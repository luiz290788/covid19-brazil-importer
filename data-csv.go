package covid19brazilimporter

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
)

// CSVReader is a data reader for csv files
type CSVReader struct{}

func getData(fileURL string) *csv.Reader {
	resp, _ := http.Get(fileURL)
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	return reader
}

func (CSVReader) read(fileURL string) (chan *Entry, error) {
	dataChannel := make(chan *Entry)

	go func() {
		reader := getData(fileURL)
		for line, _ := reader.Read(); line != nil; line, _ = reader.Read() {

			date, _ := parseDate(line[2])
			cases, _ := strconv.Atoi(line[4])
			deaths, _ := strconv.Atoi(line[6])
			dataChannel <- &Entry{
				Region: line[1],
				Date:   date,
				Cases:  cases,
				Deaths: deaths,
			}

		}
		close(dataChannel)
	}()

	// ignore header
	<-dataChannel

	return dataChannel, nil
}

func (CSVReader) supports(fileURL string) bool {
	return strings.HasSuffix(fileURL, ".csv")
}
