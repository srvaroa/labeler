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
			// Backward compatibility: If "age" is provided as a string, treat it as "at-least"
			var atLeastDuration, atMostDuration time.Duration
			var err error

			//	If they have specified a legacy "age" field, use that
			//	and treat it is as "at-least"
			if matcher.Age != "" {
				atLeastDuration, err = parseExtendedDuration(matcher.Age)
				if err != nil {
					return false, fmt.Errorf("failed to parse age parameter in configuration: %v", err)
				}
			} else if matcher.AgeRange != nil {
				// Parse "at-least" if specified
				if matcher.AgeRange.AtLeast != "" {
					atLeastDuration, err = parseExtendedDuration(matcher.AgeRange.AtLeast)
					if err != nil {
						return false, fmt.Errorf("failed to parse `age.at-least` parameter in configuration: %v", err)
					}
				}

				// Parse "at-most" if specified
				if matcher.AgeRange.AtMost != "" {
					atMostDuration, err = parseExtendedDuration(matcher.AgeRange.AtMost)
					if err != nil {
						return false, fmt.Errorf("failed to parse `age.at-most` parameter in configuration: %v", err)
					}
				}
			} else {
				return false, fmt.Errorf("no age conditions are set in config")
			}

			// Determine the creation time of the issue or PR
			var createdAt time.Time
			if target.ghIssue != nil {
				createdAt = target.ghIssue.CreatedAt.Time
			} else if target.ghPR != nil {
				createdAt = target.ghPR.CreatedAt.Time
			}

			age := time.Since(createdAt)

			//	Check if the age of the issue/PR is within the specified range
			if atLeastDuration != 0 && age < atLeastDuration {
				return false, nil
			}
			if atMostDuration != 0 && age > atMostDuration {
				return false, nil
			}

			return true, nil
		},
	}
}
