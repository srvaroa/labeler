package labeler

import (
	"fmt"
	"strconv"

	gh "github.com/google/go-github/v35/github"
)

func NewIsDraftCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Pull Request is draft"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			b, err := strconv.ParseBool(matcher.Draft)
			if err != nil {
				return false, fmt.Errorf("draft is not set in config")
			}
			if b {
				return pr.GetDraft(), nil
			}
			return !pr.GetDraft(), nil
		},
	}
}
