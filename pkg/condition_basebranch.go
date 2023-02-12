package labeler

import (
	"fmt"
	"log"
	"regexp"

	gh "github.com/google/go-github/v35/github"
)

func NewBaseBranchCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Base branch matches regex"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			if len(matcher.BaseBranch) <= 0 {
				return false, fmt.Errorf("branch is not set in config")
			}
			prBranchName := pr.Base.GetRef()
			log.Printf("Matching `%s` against: `%s`", matcher.Branch, prBranchName)
			isMatched, _ := regexp.Match(matcher.BaseBranch, []byte(prBranchName))
			return isMatched, nil
		},
	}
}
