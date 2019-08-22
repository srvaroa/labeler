package labeler

import (
	"log"
	"regexp"

	gh "github.com/google/go-github/v27/github"
)

type LabelerConfig map[string]LabelMatcher
type LabelMatcher struct {
	Title string
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

// HandleEvent takes a GitHub Event and its raw payload (see link below)
// to trigger an update to the issue / PR's labels.
//
// https://developer.github.com/v3/activity/events/types/
func (l *Labeler) HandleEvent(
	eventName string,
	payload *[]byte) error {

	// Workaround for https://github.com/google/go-github/issues/1254
	// should be removable soon-ish according to
	// https://github.com/google/go-github/issues/1254#issuecomment-523701383
	re := regexp.MustCompile(`\s+"\w+_at": "[\d\/ :APM]+",`)
	fixedPayload := []byte(re.ReplaceAllString(string(*payload), ``))

	event, err := gh.ParseWebHook(eventName, fixedPayload)
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
	for label, matcher := range *config {
		log.Printf("Matching `%s` against: `%s`", matcher.Title, pr.GetTitle())
		isMatched, _ := regexp.Match(matcher.Title, []byte(pr.GetTitle()))
		labelUpdates.set[label] = isMatched
		if isMatched {
			log.Printf("Matched on %s", label)
		}
	}
	return labelUpdates, nil
}
