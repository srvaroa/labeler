package labeler

import (
	"fmt"
	"strconv"
)

func IsDraftCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Pull Request is draft"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			b, err := strconv.ParseBool(matcher.Draft)
			if err != nil {
				return false, fmt.Errorf("draft is not set in config")
			}
			if b {
				return target.ghPR.GetDraft(), nil
			}
			return !target.ghPR.GetDraft(), nil
		},
	}
}
