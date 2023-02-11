package labeler

import (
	"fmt"
	"strconv"
)

func NewIsMergeableCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Pull Request is mergeable"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			b, err := strconv.ParseBool(matcher.Mergeable)
			if err != nil {
				return false, fmt.Errorf("mergeable is not set in config")
			}
			if b {
				return target.ghPR.GetMergeable(), nil
			}
			return !target.ghPR.GetMergeable(), nil
		},
	}
}
