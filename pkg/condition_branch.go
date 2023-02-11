package labeler

import (
	"fmt"
	"log"
	"regexp"
)

func NewBranchCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Branch matches regex"
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if target.ghPR == nil {
				log.Printf("Branch only applies on PRs, skip condition")
				return false, nil
			}
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
