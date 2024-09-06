package labeler

import (
	"strconv"
	"strings"
	"time"
)

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
