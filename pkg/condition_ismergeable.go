package labeler

import (
	"fmt"
	"strconv"
)

func IsMergeableCondition() Condition {
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

			//  Check both the mergeable state and the mergeable flag
			isMergeable := target.ghPR.GetMergeable() && target.ghPR.GetMergeableState() == "clean"

			if b {
				return isMergeable, nil
			}
			return !isMergeable, nil
		},
	}
}
