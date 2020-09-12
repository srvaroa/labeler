package labeler

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	gh "github.com/google/go-github/v27/github"
	"github.com/waigani/diffparser"
)

type LabelerConfig map[string]LabelMatcher
type LabelMatcher struct {
	Title     string
	Branch    string
	Files     []string
	Mergeable string
	SizeBelow string `yaml:"size-below"`
	SizeAbove string `yaml:"size-above"`
}

// LabelUpdates Represents a request to update the set of labels
type LabelUpdates struct {
	set map[string]bool
}

type Labeler struct {
	FetchRepoConfig    func(owner string, repoName string) (*LabelerConfig, error)
	ReplaceLabelsForPr func(owner string, repoName string, prNumber int, labels []string) error
	GetCurrentLabels   func(owner string, repoName string, prNumber int) ([]string, error)
}

type Condition struct {
	Evaluate func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error)
	GetName  func() string
}

func NewTitleCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Title matches regex"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.Title) <= 0 {
				return false, fmt.Errorf("title is not set in config")
			}
			log.Printf("Matching `%s` against: `%s`", matcher.Title, pr.GetTitle())
			isMatched, _ := regexp.Match(matcher.Title, []byte(pr.GetTitle()))
			return isMatched, nil
		},
	}
}

func NewBranchCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Branch matches regex"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.Branch) <= 0 {
				return false, fmt.Errorf("branch is not set in config")
			}
			prBranchName := pr.Head.GetRef()
			log.Printf("Matching `%s` against: `%s`", matcher.Branch, prBranchName)
			isMatched, _ := regexp.Match(matcher.Branch, []byte(prBranchName))
			return isMatched, nil
		},
	}
}

func NewFilesCondition() Condition {
	prFiles := []string{}

	return Condition{
		GetName: func() string {
			return "File matches regex"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.Files) <= 0 {
				return false, fmt.Errorf("Files are not set in config")
			}

			if len(prFiles) == 0 {
				var err error
				prFiles, err = getPrFileNames(pr)
				if err != nil {
					return false, err
				}
			}

			log.Printf("Matching `%s` against: %s", strings.Join(matcher.Files, ", "), strings.Join(prFiles, ", "))
			for _, fileMatcher := range matcher.Files {
				for _, prFile := range prFiles {
					isMatched, _ := regexp.Match(fileMatcher, []byte(prFile))
					if isMatched {
						log.Printf("Matched `%s` against: `%s`", prFile, fileMatcher)
						return isMatched, nil
					}
				}
			}
			return false, nil
		},
	}
}

func NewIsMergeableCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Pull Request is mergeable"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			b, err := strconv.ParseBool(matcher.Mergeable)
			if err != nil {
				return false, fmt.Errorf("mergeable is not set in config")
			}
			if b {
				return pr.GetMergeable(), nil
			}
			return !pr.GetMergeable(), nil
		},
	}
}

func NewSizeCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Pull Request contains a number of changes"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.SizeBelow) == 0 && len(matcher.SizeAbove) == 0 {
				return false, fmt.Errorf("size-above and size-below are not set in config")
			}
			upperBound, err := strconv.ParseInt(matcher.SizeBelow, 0, 64)
			if err != nil {
				upperBound = math.MaxInt64
				log.Printf("Upper boundary set to %d (config has invalid or empty value)", upperBound)
			}
			lowerBound, err := strconv.ParseInt(matcher.SizeAbove, 0, 32)
			if err != nil || lowerBound < 0 {
				lowerBound = 0
				log.Printf("Lower boundary set to 0 (config has invalid or empty value)")
			}
			totalChanges := int64(math.Abs(float64(pr.GetAdditions() + pr.GetDeletions())))
			log.Printf("Matching %d changes in PR against bounds: (%d, %d)", totalChanges, lowerBound, upperBound)
			isWithinBounds := totalChanges > lowerBound && totalChanges < upperBound
			return isWithinBounds, nil
		},
	}
}

// HandleEvent takes a GitHub Event and its raw payload (see link below)
// to trigger an update to the issue / PR's labels.
//
// https://developer.github.com/v3/activity/events/types/
func (l *Labeler) HandleEvent(
	eventName string,
	payload *[]byte) error {

	event, err := gh.ParseWebHook(eventName, *payload)
	if err != nil {
		return err
	}
	switch event := event.(type) {
	case *gh.PullRequestEvent:
		err = l.executeOn(event.PullRequest)
	}
	return err
}

func (l *Labeler) executeOn(pr *gh.PullRequest) error {
	owner := pr.Base.Repo.GetOwner().GetLogin()
	repoName := *pr.Base.Repo.Name

	config, err := l.FetchRepoConfig(owner, repoName)

	labelUpdates, err := l.findMatches(pr, config)
	if err != nil {
		return err
	}

	currLabels, err := l.GetCurrentLabels(owner, repoName, *pr.Number)
	if err != nil {
		return err
	}

	// intentions(label) tells whether `label` should be set in the PR
	intentions := map[string]bool{}

	// initialize with current labels
	for _, label := range currLabels {
		intentions[label] = true
	}

	// update, adding new ones and unflagging those to remove
	for label, isDesired := range labelUpdates.set {
		intentions[label] = isDesired
	}

	// filter out only labels that must be set
	desiredLabels := []string{}
	for k, v := range intentions {
		if v {
			desiredLabels = append(desiredLabels, k)
		}
	}
	log.Printf("Desired labels: %s", desiredLabels)

	return l.ReplaceLabelsForPr(owner, repoName, *pr.Number, desiredLabels)
}

// findMatches returns all updates to be made to labels for the given PR
func (l *Labeler) findMatches(pr *gh.PullRequest, config *LabelerConfig) (LabelUpdates, error) {

	labelUpdates := LabelUpdates{
		set: map[string]bool{},
	}
	conditions := []Condition{
		NewTitleCondition(),
		NewBranchCondition(),
		NewIsMergeableCondition(),
		NewSizeCondition(),
		NewFilesCondition(),
	}

	for label, matcher := range *config {
		for _, c := range conditions {
			isMatched, err := c.Evaluate(pr, matcher)
			if err != nil {
				log.Printf("%s: condition %s skipped (%s)", label, c.GetName(), err)
				continue
			}

			prev, ok := labelUpdates.set[label]

			if ok { // Other conditions were evaluated for the label
				labelUpdates.set[label] = prev && isMatched
			} else { // First condition evaluated for this label
				labelUpdates.set[label] = isMatched
			}

			log.Printf("%s: condition %s yields %t", label, c.GetName(), isMatched)
			if isMatched {
				continue
			}
		}
	}

	return labelUpdates, nil
}

// getPrFileNames returns all of the file names (old and new) of files changed in the given PR
func getPrFileNames(pr *gh.PullRequest) ([]string, error) {
	ghToken := os.Getenv("GITHUB_TOKEN")
	diffReq, err := http.NewRequest("GET", pr.GetDiffURL(), nil)

	if err != nil {
		return nil, err
	}

	diffReq.Header.Add("Authorization", "Bearer "+ghToken)
	diffRes, err := http.DefaultClient.Do(diffReq)

	if err != nil {
		return nil, err
	}

	defer diffRes.Body.Close()

	var diffRaw []byte
	prFiles := make([]string, 0)
	if diffRes.StatusCode == http.StatusOK {
		diffRaw, err = ioutil.ReadAll(diffRes.Body)

		if err != nil {
			return nil, err
		}

		diff, _ := diffparser.Parse(string(diffRaw))
		prFilesSet := map[string]struct{}{}
		// Place in a set to remove duplicates
		for _, file := range diff.Files {
			prFilesSet[file.OrigName] = struct{}{}
			prFilesSet[file.NewName] = struct{}{}
		}
		// Convert to list to make it easier to consume
		for k := range prFilesSet {
			prFiles = append(prFiles, k)
		}
	}

	return prFiles, nil
}
