package covid19brazilimporter

import (
	"errors"
	"time"
)

var formats = [...]string{
	"02/01/2006",
	"2006-01-02",
}

const outputFormat = "2006-01-02"

func parseDate(dateStr string) (time.Time, error) {
	for i := 0; i < len(formats); i++ {
		if possibleDate, err := time.Parse(formats[i], dateStr); err == nil {
			return possibleDate, nil
		}
	}
	return time.Time{}, errors.New("Could not parse without any format")
}

func serializeDate(t time.Time) string {
	return t.Format(outputFormat)
}
