package labeler

import (
	"fmt"
	"time"

	"github.com/google/go-github/v50/github"
)

func LastModifiedCondition(l *Labeler) Condition {
	return Condition{
		GetName: func() string {
			return "Last modification of issue/PR"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghIssue != nil || target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if matcher.LastModified == nil {
				return false, fmt.Errorf("no last modified conditions are set in config")
			}
			// Determine the last modification time of the issue or PR
			var lastModifiedAt *github.Timestamp
			if target.ghIssue != nil {
				lastModifiedAt = target.ghIssue.UpdatedAt
			} else if target.ghPR != nil {
				lastModifiedAt = target.ghPR.UpdatedAt
			} else {
				return false, fmt.Errorf("no issue or PR found in target")
			}
			duration := time.Since(lastModifiedAt.Time)

			if matcher.LastModified.AtMost != "" {
				maxDuration, err := parseExtendedDuration(matcher.LastModified.AtMost)
				if err != nil {
					return false, fmt.Errorf("failed to parse `last-modified.at-most` parameter in configuration: %v", err)
				}
				return duration <= maxDuration, nil
			}

			if matcher.LastModified.AtLeast != "" {
				minDuration, err := parseExtendedDuration(matcher.LastModified.AtLeast)
				if err != nil {
					return false, fmt.Errorf("failed to parse `last-modified.at-least` parameter in configuration: %v", err)
				}
				return duration >= minDuration, nil
			}

			return false, fmt.Errorf("no last modified conditions are set in config")

		},
	}
}
