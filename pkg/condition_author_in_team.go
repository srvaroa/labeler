package labeler

import (
	"fmt"
)

func AuthorInTeamCondition(l *Labeler) Condition {
	return Condition{
		GetName: func() string {
			return "Author is member of team"
		},
		CanEvaluate: func(target *Target) bool {
			return true
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if len(matcher.AuthorInTeam) <= 0 {
				return false, fmt.Errorf("author-in-team is not set in config")
			}
			// check if author is a member of team
			return l.GitHubFacade.IsUserMemberOfTeam(
				target.Owner,
				target.Author,
				matcher.AuthorInTeam, // this is the team slug
			)
		},
	}
}
