package covid19brazilimporter

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	excelize "github.com/360EntSecGroup-Skylar/excelize/v2"
)

const stateCol = 1
const cityCol = 2
const codMunCol = 4
const dateCol = 7
const casesCol = 10
const deathsCol = 11

// XLSXReader is a data reader for xlsx files
type XLSXReader struct{}

func (XLSXReader) read(fileURL string) (chan *Entry, error) {
	dataChannel := make(chan *Entry)

	go func() {
		resp, _ := http.Get(fileURL)
		file, fileError := excelize.OpenReader(resp.Body)
		defer close(dataChannel)
		if fileError != nil {
			log.Panicln(fileError.Error())
			return
		}

		rows, rowsError := file.Rows("Sheet 1")
		if rowsError != nil {
			log.Panicln(rowsError.Error())
			return
		}
		// ignore header
		rows.Next()
		// we need to call columns to thrash the header columns
		rows.Columns()

		for rows.Next() {
			row, columnsErr := rows.Columns()
			if columnsErr != nil {
				log.Printf("WARN: error reading xlsx file %v", columnsErr.Error())
				continue
			}

			state := row[stateCol]
			if len(strings.TrimSpace(state)) == 0 {
				// ignore country entries
				continue
			}

			city := row[cityCol]
			if len(strings.TrimSpace(city)) > 0 {
				// ignore city entries for now
				continue
			}

			codMun := row[codMunCol]
			if len(strings.TrimSpace(codMun)) > 0 {
				// ignore numbers of the state without city
				continue
			}

			date, _ := parseDate(row[dateCol])
			cases, _ := strconv.Atoi(row[casesCol])
			deaths, _ := strconv.Atoi(row[deathsCol])
			dataChannel <- &Entry{
				Region: row[stateCol],
				Date:   date,
				Cases:  cases,
				Deaths: deaths,
			}
		}
	}()

	return dataChannel, nil
}

func (XLSXReader) supports(fileURL string) bool {
	return strings.HasSuffix(fileURL, ".xlsx")
}
