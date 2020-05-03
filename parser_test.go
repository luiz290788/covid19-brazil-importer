package covid19brazilimporter

import (
	"testing"
	"time"
)

func TestValidDateFormatOne(t *testing.T) {
	expectedDate, _ := time.Parse("2006-01-02", "1988-07-29")

	d, err := parseDate("29/07/1988")

	if err != nil {
		t.Errorf("Unexpacted error found while parsing %v", err)
	}

	if d != expectedDate {
		t.Errorf("Date %v does not match with the expected date %v", d, expectedDate)
	}
}

func TestValidDateFormatTwo(t *testing.T) {
	expectedDate, _ := time.Parse("2006-01-02", "1988-07-29")

	d, err := parseDate("1988-07-29")

	if err != nil {
		t.Errorf("Unexpacted error found while parsing %v", err)
	}

	if d != expectedDate {
		t.Errorf("Date %v does not match with the expected date %v", d, expectedDate)
	}
}

func TestUknownDateFormat(t *testing.T) {
	_, err := parseDate("07/29/1988")

	if err == nil {
		t.Errorf("No error returned for an unknown date format")
	}
}

func TestSerializeDate(t *testing.T) {
	date, _ := time.Parse("02/01/2006", "29/07/1988")
	expectedString := "1988-07-29"

	if serializeDate(date) != expectedString {
		t.Errorf("Serialized date \"%v\" does not match with the expected \"%v\"",
			date, expectedString)
	}
}

func TestSingleDigitDay(t *testing.T) {
	date, _ := time.Parse("02/01/2006", "02/05/2020")
	expectedString := "2020-05-02"

	if serializeDate(date) != expectedString {
		t.Errorf("Serialized date \"%v\" does not match with the expected \"%v\"",
			date, expectedString)
	}
}
