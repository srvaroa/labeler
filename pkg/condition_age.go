package labeler

import (
	"fmt"
	"time"
)

func AgeCondition(l *Labeler) Condition {
	return Condition{
		GetName: func() string {
			return "Age of issue/PR"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghIssue != nil || target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			// Parse the age from the configuration
			ageDuration, err := parseExtendedDuration(matcher.Age)
			if err != nil {
				return false, fmt.Errorf("failed to parse age parameter in configuration: %v", err)
			}

			// Determine the creation time of the issue or PR
			var createdAt time.Time
			if target.ghIssue != nil {
				createdAt = target.ghIssue.CreatedAt.Time
			} else if target.ghPR != nil {
				createdAt = target.ghPR.CreatedAt.Time
			}

			age := time.Since(createdAt)

			return age > ageDuration, nil
		},
	}
}
