package labeler

import (
	"testing"

	gh "github.com/google/go-github/v50/github"
)

func TestProcessAllIssuesAndPRsOnlyOpen(t *testing.T) {
	calls := struct {
		issues []int
		prs    []int
	}{[]int{}, []int{}}

	fakeIssues := []*gh.Issue{
		{Number: gh.Int(1), State: gh.String("open"), User: &gh.User{Login: gh.String("user1")}, RepositoryURL: gh.String("http://x/y/repo")},
		{Number: gh.Int(2), State: gh.String("closed"), User: &gh.User{Login: gh.String("user2")}, RepositoryURL: gh.String("http://x/y/repo")},
	}
	fakePRs := []*gh.PullRequest{
		{Number: gh.Int(10), State: gh.String("open"), User: &gh.User{Login: gh.String("user3")}, Base: &gh.PullRequestBranch{Repo: &gh.Repository{Name: gh.String("repo"), Owner: &gh.User{Login: gh.String("y")}}}},
		{Number: gh.Int(11), State: gh.String("closed"), User: &gh.User{Login: gh.String("user4")}, Base: &gh.PullRequestBranch{Repo: &gh.Repository{Name: gh.String("repo"), Owner: &gh.User{Login: gh.String("y")}}}},
	}

	l := &Labeler{
		FetchRepoConfig: func() (*LabelerConfigV1, error) {
			return &LabelerConfigV1{Version: 1, Issues: true, Labels: []LabelMatcher{}}, nil
		},
		ReplaceLabels: func(target *Target, labels []string) error {
			if target.ghIssue != nil {
				calls.issues = append(calls.issues, target.IssueNo)
			}
			if target.ghPR != nil {
				calls.prs = append(calls.prs, target.IssueNo)
			}
			return nil
		},
		GetCurrentLabels: func(target *Target) ([]string, error) { return nil, nil },
		GitHubFacade: &GitHubFacade{
			ListIssuesByRepo: func(owner, repo string) ([]*gh.Issue, error) { return fakeIssues, nil },
			ListPRs:          func(owner, repo string) ([]*gh.PullRequest, error) { return fakePRs, nil },
		},
	}

	l.ProcessAllIssues("y", "repo")
	l.ProcessAllPRs("y", "repo")

	if len(calls.issues) != 1 || calls.issues[0] != 1 {
		t.Errorf("Expected only open issue #1 to be processed, got %+v", calls.issues)
	}
	if len(calls.prs) != 1 || calls.prs[0] != 10 {
		t.Errorf("Expected only open PR #10 to be processed, got %+v", calls.prs)
	}
}

func TestProcessAllSkipsNilState(t *testing.T) {
	calls := struct {
		issues []int
		prs    []int
	}{[]int{}, []int{}}

	// Issue and PR with nil State should be skipped
	fakeIssues := []*gh.Issue{
		{Number: gh.Int(1), State: nil, User: &gh.User{Login: gh.String("user1")}, RepositoryURL: gh.String("http://x/y/repo")},
		{Number: gh.Int(2), State: gh.String("open"), User: &gh.User{Login: gh.String("user2")}, RepositoryURL: gh.String("http://x/y/repo")},
	}
	fakePRs := []*gh.PullRequest{
		{Number: gh.Int(10), State: nil, User: &gh.User{Login: gh.String("user3")}, Base: &gh.PullRequestBranch{Repo: &gh.Repository{Name: gh.String("repo"), Owner: &gh.User{Login: gh.String("y")}}}},
		{Number: gh.Int(11), State: gh.String("open"), User: &gh.User{Login: gh.String("user4")}, Base: &gh.PullRequestBranch{Repo: &gh.Repository{Name: gh.String("repo"), Owner: &gh.User{Login: gh.String("y")}}}},
	}

	l := &Labeler{
		FetchRepoConfig: func() (*LabelerConfigV1, error) {
			return &LabelerConfigV1{Version: 1, Issues: true, Labels: []LabelMatcher{}}, nil
		},
		ReplaceLabels: func(target *Target, labels []string) error {
			if target.ghIssue != nil {
				calls.issues = append(calls.issues, target.IssueNo)
			}
			if target.ghPR != nil {
				calls.prs = append(calls.prs, target.IssueNo)
			}
			return nil
		},
		GetCurrentLabels: func(target *Target) ([]string, error) { return nil, nil },
		GitHubFacade: &GitHubFacade{
			ListIssuesByRepo: func(owner, repo string) ([]*gh.Issue, error) { return fakeIssues, nil },
			ListPRs:          func(owner, repo string) ([]*gh.PullRequest, error) { return fakePRs, nil },
		},
	}

	l.ProcessAllIssues("y", "repo")
	l.ProcessAllPRs("y", "repo")

	// Only the items with explicit "open" state should be processed
	if len(calls.issues) != 1 || calls.issues[0] != 2 {
		t.Errorf("Expected only issue #2 (with 'open' state) to be processed, got %+v", calls.issues)
	}
	if len(calls.prs) != 1 || calls.prs[0] != 11 {
		t.Errorf("Expected only PR #11 (with 'open' state) to be processed, got %+v", calls.prs)
	}
}

func TestProcessAllHandlesMixedCaseState(t *testing.T) {
	calls := struct {
		issues []int
		prs    []int
	}{[]int{}, []int{}}

	// Test case-insensitive state matching
	fakeIssues := []*gh.Issue{
		{Number: gh.Int(1), State: gh.String("Open"), User: &gh.User{Login: gh.String("user1")}, RepositoryURL: gh.String("http://x/y/repo")},
		{Number: gh.Int(2), State: gh.String("OPEN"), User: &gh.User{Login: gh.String("user2")}, RepositoryURL: gh.String("http://x/y/repo")},
		{Number: gh.Int(3), State: gh.String("Closed"), User: &gh.User{Login: gh.String("user3")}, RepositoryURL: gh.String("http://x/y/repo")},
	}
	fakePRs := []*gh.PullRequest{
		{Number: gh.Int(10), State: gh.String("OPEN"), User: &gh.User{Login: gh.String("user4")}, Base: &gh.PullRequestBranch{Repo: &gh.Repository{Name: gh.String("repo"), Owner: &gh.User{Login: gh.String("y")}}}},
		{Number: gh.Int(11), State: gh.String("CLOSED"), User: &gh.User{Login: gh.String("user5")}, Base: &gh.PullRequestBranch{Repo: &gh.Repository{Name: gh.String("repo"), Owner: &gh.User{Login: gh.String("y")}}}},
	}

	l := &Labeler{
		FetchRepoConfig: func() (*LabelerConfigV1, error) {
			return &LabelerConfigV1{Version: 1, Issues: true, Labels: []LabelMatcher{}}, nil
		},
		ReplaceLabels: func(target *Target, labels []string) error {
			if target.ghIssue != nil {
				calls.issues = append(calls.issues, target.IssueNo)
			}
			if target.ghPR != nil {
				calls.prs = append(calls.prs, target.IssueNo)
			}
			return nil
		},
		GetCurrentLabels: func(target *Target) ([]string, error) { return nil, nil },
		GitHubFacade: &GitHubFacade{
			ListIssuesByRepo: func(owner, repo string) ([]*gh.Issue, error) { return fakeIssues, nil },
			ListPRs:          func(owner, repo string) ([]*gh.PullRequest, error) { return fakePRs, nil },
		},
	}

	l.ProcessAllIssues("y", "repo")
	l.ProcessAllPRs("y", "repo")

	// Both "Open" and "OPEN" should be processed (case-insensitive)
	if len(calls.issues) != 2 {
		t.Errorf("Expected issues #1 and #2 to be processed, got %+v", calls.issues)
	}
	if len(calls.prs) != 1 || calls.prs[0] != 10 {
		t.Errorf("Expected only PR #10 to be processed, got %+v", calls.prs)
	}
}
