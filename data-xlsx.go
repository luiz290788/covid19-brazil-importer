package covid19brazilimporter

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	excelize "github.com/360EntSecGroup-Skylar/excelize/v2"
)

// XLSXReader is a data reader for xlsx files
type XLSXReader struct{}

func (XLSXReader) read(fileURL string) (Regions, error) {
	regions := make(Regions)

	resp, _ := http.Get(fileURL)
	file, fileError := excelize.OpenReader(resp.Body)
	if fileError != nil {
		return nil, fileError
	}

	rows, rowsError := file.Rows("Sheet 1")
	if rowsError != nil {
		return nil, rowsError
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
		pointerRows := make([]*string, len(row))
		for index, value := range row {
			valuePointer := value
			pointerRows[index] = &valuePointer
		}
		err := regions.ProcessRow(pointerRows)
		if err != nil {
			fmt.Printf("error processing row %v\n", err.Error())
		}
	}

	return regions, nil
}

func (XLSXReader) supports(fileURL string) bool {
	return strings.HasSuffix(fileURL, ".xlsx")
}
