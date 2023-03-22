package labeler

import (
	"fmt"
	"log"
	"math"
	"strconv"
)

func SizeCondition() Condition {
	return Condition{
		GetName: func() string {
			return "Pull Request contains a number of changes"
		},
		CanEvaluate: func(target *Target) bool {
			return target.ghPR != nil
		},
		Evaluate: func(target *Target, matcher LabelMatcher) (bool, error) {

			if isNewConfig(matcher) && isOldConfig(matcher) {
				log.Printf("WARNING: you are using both the old " +
					"`size-above`/`size-below` settings together with " +
					"the newer `size`. You should use only the latter. " +
					"This condition will apply the configurations set in `Size` " +
					"and ignore the rest")
			}

			realMatcher := matcher.Size
			if realMatcher == nil {
				if matcher.SizeBelow == "" && matcher.SizeAbove == "" {
					return false, fmt.Errorf("no size conditions are set in config")
				}
				realMatcher = &SizeConfig{
					Above: matcher.SizeAbove,
					Below: matcher.SizeBelow,
				}
			}

			log.Printf("Checking PR size using config: %+v", realMatcher)

			upperBound, err := strconv.ParseInt(realMatcher.Below, 0, 64)
			if err != nil {
				upperBound = math.MaxInt64
				log.Printf("Upper boundary set to %d (config has invalid or empty value)", upperBound)
			}
			lowerBound, err := strconv.ParseInt(realMatcher.Above, 0, 32)
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

func isNewConfig(matcher LabelMatcher) bool {
	return matcher.Size != nil
}

func isOldConfig(matcher LabelMatcher) bool {
	return matcher.SizeAbove != "" || matcher.SizeBelow != ""
}
