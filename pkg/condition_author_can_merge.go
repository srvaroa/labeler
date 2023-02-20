package labeler

import (
	"fmt"
)

func AuthorCanMergeCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Author can merge"
		},
		CanEvaluate: func(target *Target) bool {
			return true
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if len(matcher.AuthorCanMerge) <= 0 {
				return false, fmt.Errorf("AuthorCanMerge not set in repository")
			}
			ghRepo := target.ghPR.GetAuthorAssociation()
			canMerge := ghRepo == "OWNER"
			fmt.Printf("User: %s can merge? %t\n", target.Author, canMerge)
			return canMerge, nil
		},
	}
}
