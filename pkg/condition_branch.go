package labeler

import (
	"fmt"
	"log"
	"regexp"
)

func BranchCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Branch matches regex"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if len(matcher.Branch) <= 0 {
				return false, fmt.Errorf("branch is not set in config")
			}
			prBranchName := target.ghPR.Head.GetRef()
			log.Printf("Matching `%s` against: `%s`", matcher.Branch, prBranchName)
			isMatched, _ := regexp.Match(matcher.Branch, []byte(prBranchName))
			return isMatched, nil
		},
	}
}
