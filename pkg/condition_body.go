package labeler

import (
	"fmt"
	"log"
	"regexp"
)

func BodyCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Body matches regex"
		},
		CanEvaluate: func(target *Target) bool {
			return true
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if len(matcher.Body) <= 0 {
				return false, fmt.Errorf("body is not set in config")
			}
			log.Printf("Matching `%s` against: `%s`", matcher.Body, target.Body)
			isMatched, _ := regexp.Match(matcher.Body, []byte(target.Body))
			return isMatched, nil
		},
	}
}
