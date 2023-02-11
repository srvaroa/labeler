package labeler

import (
	"log"
	"strings"

	gh "github.com/google/go-github/v35/github"
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
	Draft      string
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
	FetchRepoConfig  func() (*LabelerConfigV1, error)
	ReplaceLabels    func(target *Target, labels []string) error
	GetCurrentLabels func(target *Target) ([]string, error)
	Client           HttpClient
}

type Condition struct {
	CanEvaluate func(target *Target) bool
	Evaluate    func(target *Target, matcher LabelMatcher) (bool, error)
	GetName     func() string
}

type Target struct {
	Author   string
	Body     string
	IssueNo  int
	Title    string
	Owner    string
	RepoName string
	ghPR     *gh.PullRequest
	ghIssue  *gh.Issue
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
		err = l.ExecuteOn(WrapPrAsTarget(event.PullRequest))
	case *gh.PullRequestTargetEvent:
		err = l.ExecuteOn(WrapPrAsTarget(event.PullRequest))
	case *gh.IssuesEvent:
		err = l.ExecuteOn(WrapIssueAsTarget(event.Issue))
	default:
		log.Printf("Event type is not supported, please review your workflow config")
	}
	return err
}

func WrapPrAsTarget(pr *gh.PullRequest) *Target {
	return &Target{
		Author:   *pr.GetUser().Login,
		Body:     pr.GetBody(),
		IssueNo:  *pr.Number,
		Title:    pr.GetTitle(),
		Owner:    pr.Base.Repo.GetOwner().GetLogin(),
		RepoName: *pr.Base.Repo.Name,
		ghPR:     pr,
		ghIssue:  nil,
	}
}

func WrapIssueAsTarget(issue *gh.Issue) *Target {

	// TODO: go-github@v50 has a Repository property that
	// avoids this.
	repoUrlSplit := strings.Split(*issue.RepositoryURL, "/")
	repoName := repoUrlSplit[len(repoUrlSplit)-1]
	owner := repoUrlSplit[len(repoUrlSplit)-2]

	return &Target{
		Author:   *issue.GetUser().Login,
		Body:     issue.GetBody(),
		IssueNo:  *issue.Number,
		Title:    issue.GetTitle(),
		Owner:    owner,
		RepoName: repoName,
		ghPR:     nil,
		ghIssue:  issue,
	}
}

func (l *Labeler) ExecuteOn(target *Target) error {

	log.Printf("Matching labels on target %+v", target)

	config, err := l.FetchRepoConfig()

	labelUpdates, err := l.findMatches(target, config)
	if err != nil {
		return err
	}

	currLabels, err := l.GetCurrentLabels(target)
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

	return l.ReplaceLabels(target, desiredLabels)
}

// findMatches returns all updates to be made to labels for the given target
func (l *Labeler) findMatches(target *Target, config *LabelerConfigV1) (LabelUpdates, error) {

	labelUpdates := LabelUpdates{
		set: map[string]bool{},
	}
	conditions := []Condition{
		NewTitleCondition(),
		NewBranchCondition(),
		NewBaseBranchCondition(),
		NewIsMergeableCondition(),
		NewIsDraftCondition(),
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
			if !c.CanEvaluate(target) {
				log.Printf("[%s] skip, event not supported by condition", c.GetName())
				continue
			}
			isMatched, err := c.Evaluate(target, matcher)
			if err != nil {
				log.Printf("[%s] skip, %s", c.GetName(), err)
				continue
			}
			log.Printf("[%s] yields %t", c.GetName(), isMatched)

			prev, ok := labelUpdates.set[label]
			if ok { // Other conditions were evaluated for the label
				labelUpdates.set[label] = prev && isMatched
			} else { // First condition evaluated for this label
				labelUpdates.set[label] = isMatched
			}

			if isMatched {
				continue
			}
		}
	}

	return labelUpdates, nil
}
