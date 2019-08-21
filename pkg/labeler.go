package labeler

import (
	"context"
	"regexp"

	gh "github.com/google/go-github/v26/github"
)

type LabelerConfig map[string]LabelMatcher
type LabelMatcher struct {
	Title string "json:title"
}

// LabelUpdates Represents a request to update the set of labels
type LabelUpdates struct {
	set map[string]bool
}

type Labeler struct {
	fetchRepoConfig    func(owner string, repoName string) (LabelerConfig, error)
	replaceLabelsForPr func(owner string, repoName string, prNumber int, labels []string) error
	getCurrentLabels   func(owner string, repoName string, prNumber int) ([]string, error)
}

func NewLabeler(github *gh.Client) *Labeler {
	l := Labeler{

		fetchRepoConfig: func(owner string, repoName string) (LabelerConfig, error) {
			// TODO: use the actual upstream config
			// TODO: depending on who's running the labeler, config may
			// come from different places, e.g. with Actions it might
			// make sense to have it in env variables
			// (https://developer.github.com/actions/creating-github-actions/accessing-the-runtime-environment/#environment-variables)
			return LabelerConfig{}, nil
		},

		replaceLabelsForPr: func(owner string, repoName string, prNumber int, labels []string) error {
			_, _, err := github.Issues.ReplaceLabelsForIssue(
				context.Background(), owner, repoName, prNumber, labels)
			return err
		},

		getCurrentLabels: func(owner string, repoName string, prNumber int) ([]string, error) {
			opts := gh.ListOptions{} // TODO: ignoring pagination here
			currLabels, _, err := github.Issues.ListLabelsByIssue(
				context.Background(), owner, repoName, prNumber, &opts)

			labels := []string{}
			for _, label := range currLabels {
				labels = append(labels, *label.Name)
			}
			return labels, err
		},
	}
	return &l
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

	config, err := l.fetchRepoConfig(owner, repoName)

	labelUpdates, err := l.findMatches(pr, &config)
	if err != nil {
		return err
	}

	currLabels, err := l.getCurrentLabels(owner, repoName, *pr.Number)
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

	return l.replaceLabelsForPr(owner, repoName, *pr.Number, desiredLabels)
}

// findMatches returns all updates to be made to labels for the given PR
func (l *Labeler) findMatches(pr *gh.PullRequest, config *LabelerConfig) (LabelUpdates, error) {

	labelUpdates := LabelUpdates{
		set: map[string]bool{},
	}
	for label, matcher := range *config {
		isMatched, _ := regexp.Match(matcher.Title, []byte(pr.GetTitle()))
		labelUpdates.set[label] = isMatched
	}
	return labelUpdates, nil
}
