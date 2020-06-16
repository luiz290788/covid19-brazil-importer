package covid19brazilimporter

import (
	"fmt"
	"time"
)

var formats = [...]string{
	"02/01/2006",
	"01-02-06",
	"2/1/2006",
	"2006-01-02",
}

const outputFormat = "2006-01-02"

func parseDate(dateStr string) (time.Time, error) {
	for i := 0; i < len(formats); i++ {
		if possibleDate, err := time.Parse(formats[i], dateStr); err == nil {
			return possibleDate, nil
		}
	}
	return time.Time{}, fmt.Errorf("Could not parse without any format: %s", dateStr)
}

func serializeDate(t time.Time) string {
	return t.Format(outputFormat)
}
