package labeler

import (
	"fmt"
	"log"
	"strings"
)

func AuthorCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Author matches"
		},
		CanEvaluate: func(target *Target) bool {
			return true
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if len(matcher.Authors) <= 0 {
				return false, fmt.Errorf("Users are not set in config")
			}

			log.Printf("Matching `%s` against: `%v`", matcher.Authors, target.Author)
			for _, author := range matcher.Authors {
				if strings.ToLower(author) == strings.ToLower(target.Author) {
					return true, nil
				}
			}
			return false, nil
		},
	}
}
