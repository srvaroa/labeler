package labeler

import (
	"fmt"
	"log"
)

func TypeCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Target type matches defined type"
		},
		CanEvaluate: func(target *Target) bool {
			return true
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if len(matcher.Type) <= 0 {
				return false, fmt.Errorf("type is not set in config")
			} else if matcher.Type != "pull_request" && matcher.Type != "issue" {
				return false, fmt.Errorf("type musst be of value 'pull_request' or 'issue'")
			}

			var targetType string
			if target.ghPR != nil {
				targetType = "pull_request"
			} else if target.ghIssue != nil {
				targetType = "issue"
			} else {
				return false, fmt.Errorf("target is neither pull_request nor issue")
			}

			log.Printf("Matching `%s` against: `%s`", matcher.Type, targetType)
			return matcher.Type == targetType || matcher.Type == "all", nil
		},
	}
}
