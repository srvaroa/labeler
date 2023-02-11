package labeler

import (
	"fmt"
	"log"
	"regexp"
)

func NewBaseBranchCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Base branch matches regex"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if len(matcher.BaseBranch) <= 0 {
				return false, fmt.Errorf("branch is not set in config")
			}
			prBranchName := target.ghPR.Base.GetRef()
			log.Printf("Matching `%s` against: `%s`", matcher.Branch, prBranchName)
			isMatched, _ := regexp.Match(matcher.BaseBranch, []byte(prBranchName))
			return isMatched, nil
		},
	}
}
