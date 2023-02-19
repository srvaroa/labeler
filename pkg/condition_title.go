package labeler

import (
	"fmt"
	"log"
	"regexp"
)

func TitleCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Title matches regex"
		},
		CanEvaluate: func(target *Target) bool {
			return true
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if len(matcher.Title) <= 0 {
				return false, fmt.Errorf("title is not set in config")
			}
			log.Printf("Matching `%s` against: `%s`", matcher.Title, target.Title)
			isMatched, _ := regexp.Match(matcher.Title, []byte(target.Title))
			return isMatched, nil
		},
	}
}
