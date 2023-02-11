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
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if target.ghPR == nil {
				log.Printf("Base branch only applies on PRs, skip condition")
				return false, nil
			}
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
