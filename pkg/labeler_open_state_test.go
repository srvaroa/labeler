package labeler

import (
	   "testing"
	   gh "github.com/google/go-github/v50/github"
)

func TestProcessAllIssuesAndPRsOnlyOpen(t *testing.T) {
	calls := struct{
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
					   ListPRs: func(owner, repo string) ([]*gh.PullRequest, error) { return fakePRs, nil },
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
