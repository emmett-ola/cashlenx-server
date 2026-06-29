package statistic_service

import (
	"errors"
	"time"

	"github.com/macar-x/cashlenx-server/util"
)

func parsePeriodDate(period, value string) (time.Time, error) {
	var layouts []string

	switch period {
	case "daily":
		return util.ParseDate(value)
	case "monthly":
		layouts = []string{"200601", "2006-01"}
	case "yearly":
		layouts = []string{"2006"}
	default:
		return time.Time{}, errors.New("period must be 'daily', 'monthly', or 'yearly'")
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}

	// Preserve compatibility with callers that provide a complete date for
	// monthly or yearly statistics.
	return util.ParseDate(value)
}
