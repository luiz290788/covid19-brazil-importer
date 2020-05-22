package covid19brazilimporter

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const stateCol = 1
const cityCol = 2
const codMunCol = 4
const dateCol = 7
const casesCol = 10
const deathsCol = 11

// XLSXReader is a data reader for xlsx files
type XLSXReader struct{}

type xlsxC struct {
	R string `xml:"r,attr,omitempty"`
	V string `xml:"v,omitempty"`
}

type xlsxSi struct {
	T string `xml:"t"`
}

type xslxSst struct {
	Count int `xml:"count,attr,omitempty"`
}

func (XLSXReader) read(fileURL string) (chan *Entry, error) {
	dataChannel := make(chan *Entry)

	go func() {
		defer close(dataChannel)
		resp, _ := http.Get(fileURL)
		body, _ := ioutil.ReadAll(resp.Body)
		reader, _ := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		var stringTable []*string = nil
		for _, file := range reader.File {
			if file.Name == "xl/sharedStrings.xml" {
				stringTable, _ = readSharedStrings(file)
			} else if file.Name == "xl/worksheets/sheet1.xml" {
				rows, _ := readRows(file, stringTable)
				for _, row := range rows {
					state := row[stateCol]
					if state == nil {
						// ignore country entries
						continue
					}

					city := row[cityCol]
					if city != nil {
						// ignore city entries for now
						continue
					}

					codMun := row[codMunCol]
					if codMun != nil {
						// ignore numbers of the state without city
						continue
					}

					date, _ := parseDate(*row[dateCol])
					cases, _ := strconv.Atoi(*row[casesCol])
					deaths, _ := strconv.Atoi(*row[deathsCol])
					dataChannel <- &Entry{
						Region: *row[stateCol],
						Date:   date,
						Cases:  cases,
						Deaths: deaths,
					}
				}
			}
		}
	}()

	return dataChannel, nil
}

func (XLSXReader) supports(fileURL string) bool {
	return strings.HasSuffix(fileURL, ".xlsx")
}

func readSharedStrings(file *zip.File) (strs []*string, err error) {
	reader, _ := file.Open()
	decoder := xml.NewDecoder(reader)

	index := 0
	for {
		token, _ := decoder.Token()
		if token == nil {
			return
		}
		switch actualToken := token.(type) {
		case xml.StartElement:
			if actualToken.Name.Local == "si" {
				sharedString := &xlsxSi{}
				decoder.DecodeElement(sharedString, &actualToken)
				strs[index] = &sharedString.T
				index++
			} else if actualToken.Name.Local == "sst" {
				for _, attr := range actualToken.Attr {
					if attr.Name.Local == "count" {
						count, _ := strconv.Atoi(attr.Value)
						strs = make([]*string, count)
					}
				}
			}
			break
		}
	}
}

func readRows(file *zip.File, stringTable []*string) ([][]*string, error) {
	reader, sheetErr := file.Open()
	if sheetErr != nil {
		log.Fatalf("error opening file: %v", sheetErr.Error())
	}

	decoder := xml.NewDecoder(reader)

	var rows [][]*string

	for {
		token, _ := decoder.Token()

		if token == nil {
			break
		}
		switch actualToken := token.(type) {
		case xml.StartElement:
			if actualToken.Name.Local == "row" {
				row, _ := readRow(decoder, actualToken, stringTable)
				rows = append(rows, row)
			}
			break
		}
	}
	return rows, nil
}

func readRow(decoder *xml.Decoder, start xml.StartElement, stringTable []*string) (row []*string, err error) {
	for {
		token, tokenErr := decoder.Token()
		if tokenErr != nil {
			return
		}

		if token == nil {
			break
		}

		switch actualToken := token.(type) {
		case xml.StartElement:
			if actualToken.Name.Local == "c" {
				value := &xlsxC{}
				decoder.DecodeElement(value, &actualToken)
				index, indexErr := strconv.Atoi(value.V)
				if indexErr != nil {
					row = append(row, nil)
				} else {
					row = append(row, stringTable[index])
				}
			}
		case xml.EndElement:
			if actualToken.Name.Local == "row" {
				return
			}
			break
		}
	}
	return
}
