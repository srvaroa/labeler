package labeler

import (
	"fmt"
	"log"
	"regexp"
)

func NewBodyCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Body matches regex"
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
