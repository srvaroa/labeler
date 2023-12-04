package labeler

import (
	"fmt"
	"strconv"
)

func AuthorCanMergeCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Author can merge"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			expected, err := strconv.ParseBool(matcher.AuthorCanMerge)
			if err != nil {
				return false, fmt.Errorf("author-can-merge doesn't have a valid value in config")
			}

			ghRepo := target.ghPR.GetAuthorAssociation()
			canMerge := ghRepo == "OWNER"

			if expected && canMerge {
				fmt.Printf("User: %s can merge %t, condition matched\n",
					target.Author, canMerge)
				return true, nil
			}

			if !expected && !canMerge {
				fmt.Printf("User: %s can not merge %t, condition matched\n",
					target.Author, canMerge)
				return true, nil
			}

			fmt.Printf("Condition not matched")
			return false, nil
		},
	}
}
