package labeler

import (
	"fmt"
	"log"
	"strconv"
)

func NewIsMergeableCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Pull Request is mergeable"
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if target.ghPR == nil {
				log.Printf("IsMergeable only applies on PRs, skip condition")
				return false, nil
			}
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
