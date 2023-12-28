package labeler

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	gh "github.com/google/go-github/v50/github"
)

func SizeCondition(l *Labeler) Condition {
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

			totalChanges, err := l.getModifiedLinesCount(target.ghPR, realMatcher.ExcludeFiles)
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

func (l *Labeler) getModifiedLinesCount(pr *gh.PullRequest, exclusions []string) (int64, error) {

	if len(exclusions) == 0 {
		// no exclusions so we can just rely on GH's summary which is
		// more lightweight
		return int64(math.Abs(float64(pr.GetAdditions() + pr.GetDeletions()))), nil
	}

	// Get the diff for the pull request
	urlParts := strings.Split(pr.GetBase().GetRepo().GetHTMLURL(), "/")
	owner := urlParts[len(urlParts)-2]
	repo := urlParts[len(urlParts)-1]
	diff, err := l.GitHubFacade.GetRawDiff(owner, repo, pr.GetNumber())
	if err != nil {
		return 0, err
	}

	// Count the number of lines that start with "+" or "-"
	var count int64
	var countFile = false
	for _, line := range strings.Split(diff, "\n") {
		if line == "+++ /dev/null" || line == "--- /dev/null" {
			// ignore, these are removed or added files
			continue
		}
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			// We're in a file's block
			path := strings.TrimPrefix(line, "---")
			path = strings.TrimPrefix(path, "+++")
			path = strings.TrimPrefix(path, "a/")
			path = strings.TrimPrefix(path, "b/")
			path = strings.TrimSpace(path)
			// Check if the file path matches any of the excluded files
			countFile = !isFileExcluded(path, exclusions)
			if countFile {
				log.Printf("Counting changes in file %s", path)
			} else {
				log.Printf("Ignoring file %s", path)
			}
			continue
		}
		if countFile && (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-")) {
			log.Printf("Count line %s", line)
			count++
		}
	}

	log.Printf("Total count %d", count)

	return count, nil
}

func isFileExcluded(path string, exclusions []string) bool {
	for _, exclusion := range exclusions {
		exclusionRegex, err := regexp.Compile(exclusion)
		if err != nil {
			log.Printf("Error compiling file exclusion regex %s: %s", exclusion, err)
		} else if exclusionRegex.MatchString(path) {
			return true
		}
	}
	return false
}
