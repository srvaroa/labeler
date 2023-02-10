package labeler

import (
	"fmt"
	"log"
	"regexp"

	gh "github.com/google/go-github/v35/github"
)

func NewBranchCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Branch matches regex"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.Branch) <= 0 {
				return false, fmt.Errorf("branch is not set in config")
			}
			prBranchName := pr.Head.GetRef()
			log.Printf("Matching `%s` against: `%s`", matcher.Branch, prBranchName)
			isMatched, _ := regexp.Match(matcher.Branch, []byte(prBranchName))
			return isMatched, nil
		},
	}
}
