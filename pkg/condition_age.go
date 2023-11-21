package labeler

import (
	"fmt"
	"strconv"
	"strings"
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

func parseExtendedDuration(s string) (time.Duration, error) {
	multiplier := time.Hour * 24 // default to days

	if strings.HasSuffix(s, "w") {
		multiplier = time.Hour * 24 * 7 // weeks
		s = strings.TrimSuffix(s, "w")
	} else if strings.HasSuffix(s, "y") {
		multiplier = time.Hour * 24 * 365 // years
		s = strings.TrimSuffix(s, "y")
	} else if strings.HasSuffix(s, "d") {
		s = strings.TrimSuffix(s, "d") // days
	} else {
		return time.ParseDuration(s) // default to time.ParseDuration for hours, minutes, seconds
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	return time.Duration(value) * multiplier, nil
}
