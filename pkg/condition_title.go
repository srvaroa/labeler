package labeler

import (
	"fmt"
	"log"
	"regexp"

	gh "github.com/google/go-github/v35/github"
)

func NewTitleCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Title matches regex"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.Title) <= 0 {
				return false, fmt.Errorf("title is not set in config")
			}
			log.Printf("Matching `%s` against: `%s`", matcher.Title, pr.GetTitle())
			isMatched, _ := regexp.Match(matcher.Title, []byte(pr.GetTitle()))
			return isMatched, nil
		},
	}
}
