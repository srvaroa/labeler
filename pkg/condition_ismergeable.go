package labeler

import (
	"fmt"
	"strconv"

	gh "github.com/google/go-github/v35/github"
)

func NewIsMergeableCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Pull Request is mergeable"
		},
		Evaluate: func(pr *gh.PullRequest, matcher LabelMatcher) (bool, error) {
			b, err := strconv.ParseBool(matcher.Mergeable)
			if err != nil {
				return false, fmt.Errorf("mergeable is not set in config")
			}
			if b {
				return pr.GetMergeable(), nil
			}
			return !pr.GetMergeable(), nil
		},
	}
}
