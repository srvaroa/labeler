package labeler

import (
	"fmt"
	"log"
	"regexp"

	gh "github.com/google/go-github/v35/github"
)

func NewBodyCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Body matches regex"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.Body) <= 0 {
				return false, fmt.Errorf("body is not set in config")
			}
			log.Printf("Matching `%s` against: `%s`", matcher.Body, pr.GetBody())
			isMatched, _ := regexp.Match(matcher.Body, []byte(pr.GetBody()))
			return isMatched, nil
		},
	}
}
