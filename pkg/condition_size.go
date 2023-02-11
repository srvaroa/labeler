package labeler

import (
	"fmt"
	"log"
	"math"
	"strconv"
)

func NewSizeCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Pull Request contains a number of changes"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {
			if len(matcher.SizeBelow) == 0 && len(matcher.SizeAbove) == 0 {
				return false, fmt.Errorf("size-above and size-below are not set in config")
			}
			upperBound, err := strconv.ParseInt(matcher.SizeBelow, 0, 64)
			if err != nil {
				upperBound = math.MaxInt64
				log.Printf("Upper boundary set to %d (config has invalid or empty value)", upperBound)
			}
			lowerBound, err := strconv.ParseInt(matcher.SizeAbove, 0, 32)
			if err != nil || lowerBound < 0 {
				lowerBound = 0
				log.Printf("Lower boundary set to 0 (config has invalid or empty value)")
			}
			totalChanges := int64(math.Abs(float64(target.ghPR.GetAdditions() + target.ghPR.GetDeletions())))
			log.Printf("Matching %d changes in PR against bounds: (%d, %d)", totalChanges, lowerBound, upperBound)
			isWithinBounds := totalChanges > lowerBound && totalChanges < upperBound
			return isWithinBounds, nil
		},
	}
}
