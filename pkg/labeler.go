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

	gh "github.com/google/go-github/v35/github"
	"github.com/waigani/diffparser"
)

type LabelMatcher struct {
	Label      string
	Title      string
	Branch     string
	BaseBranch string `yaml:"base-branch"`
	Body       string
	Files      []string
	Authors    []string
	Mergeable  string
	SizeBelow  string `yaml:"size-below"`
	SizeAbove  string `yaml:"size-above"`
}

type LabelerConfigV0 map[string]LabelMatcher

type LabelerConfigV1 struct {
	Version int32
	// when set to true, we will only add labels when they match a rule
	// but it will NOT remove labels that were previously set and stop
	// matching a rule
	AppendOnly bool
	Labels     []LabelMatcher
}

// LabelUpdates Represents a request to update the set of labels
type LabelUpdates struct {
	set map[string]bool
}

type Labeler struct {
	FetchRepoConfig    func(owner string, repoName string) (*LabelerConfigV1, error)
	ReplaceLabelsForPr func(owner string, repoName string, prNumber int, labels []string) error
	GetCurrentLabels   func(owner string, repoName string, prNumber int) ([]string, error)
	Client             HttpClient
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

func NewBaseBranchCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Base branch matches regex"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.BaseBranch) <= 0 {
				return false, fmt.Errorf("branch is not set in config")
			}
			prBranchName := pr.Base.GetRef()
			log.Printf("Matching `%s` against: `%s`", matcher.Branch, prBranchName)
			isMatched, _ := regexp.Match(matcher.BaseBranch, []byte(prBranchName))
			return isMatched, nil
		},
	}
}

func NewBodyCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Body matches regex"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.Body) <= 0 {
				return false, fmt.Errorf("body is not set in config")
			}
			log.Printf("Matching `%s` against: `%s`", matcher.Body, pr.GetBody())
			isMatched, _ := regexp.Match(matcher.Body, []byte(pr.GetBody()))
			return isMatched, nil
		},
	}
}

func NewFilesCondition(l *Labeler) Condition {
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
				prFiles, err = l.getPrFileNames(pr)
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

func NewAuthorCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Author matches"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.Authors) <= 0 {
				return false, fmt.Errorf("Users are not set in config")
			}

			prAuthor := pr.GetUser().Login

			log.Printf("Matching `%s` against: `%v`", matcher.Authors, *prAuthor)
			for _, author := range matcher.Authors {
				if strings.ToLower(author) == strings.ToLower(*prAuthor) {
					return true, nil
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
		err = l.ExecuteOn(event.PullRequest)
	case *gh.PullRequestTargetEvent:
		err = l.ExecuteOn(event.PullRequest)
	}
	return err
}

func (l *Labeler) ExecuteOn(pr *gh.PullRequest) error {
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

	// update, adding new ones and unflagging those to remove if
	// necessary
	for label, isDesired := range labelUpdates.set {
		if config.AppendOnly {
			// If we DO NOT allow deletions, then we will respect
			// labels that were already set in the current set
			// but add new ones that matched the repo
			intentions[label] = intentions[label] || isDesired
		} else {
			// If we allow deletions, then we set / unset the label
			// based on the result of the rule checks
			intentions[label] = isDesired
		}
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
func (l *Labeler) findMatches(pr *gh.PullRequest, config *LabelerConfigV1) (LabelUpdates, error) {

	labelUpdates := LabelUpdates{
		set: map[string]bool{},
	}
	conditions := []Condition{
		NewTitleCondition(),
		NewBranchCondition(),
		NewBaseBranchCondition(),
		NewIsMergeableCondition(),
		NewSizeCondition(),
		NewBodyCondition(),
		NewFilesCondition(l),
		NewAuthorCondition(),
	}

	for _, matcher := range config.Labels {
		label := matcher.Label

		if labelUpdates.set[label] {
			// This label was already matched in another matcher
			// so we already decided to apply it and need to
			// evaluate no more matchers.
			//
			// Note that multiple matchers for the same label
			// are combined with an OR.
			continue
		}

		// Reset the label as we're going to re-evaluate it in a new
		// condition
		delete(labelUpdates.set, label)

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
func (l *Labeler) getPrFileNames(pr *gh.PullRequest) ([]string, error) {
	log.Printf("getPrFileNames for pr - " + pr.GetURL())
	ghToken := os.Getenv("GITHUB_TOKEN")
	diffReq, err := http.NewRequest("GET", pr.GetURL(), nil)

	if err != nil {
		return nil, err
	}

	if ghToken != "" {
		diffReq.Header.Add("Authorization", "Bearer "+ghToken)
	} else {
		log.Printf("Env var GITHUB_TOKEN is missing, using annonymous request")
	}
	diffReq.Header.Add("Accept", "application/vnd.github.v3.diff")
	diffRes, err := l.Client.Do(diffReq)

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

		diff, err := diffparser.Parse(string(diffRaw))
		if err != nil {
			return nil, err
		}

		log.Printf("got diff %s, parsed %+v", string(diffRaw), diff)
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
		log.Printf("diff files %s", prFiles)
	} else {
		log.Printf("failed with status %s", diffRes.Status)
	}

	return prFiles, nil
}
