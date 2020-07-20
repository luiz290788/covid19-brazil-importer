package covid19brazilimporter

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

// Entry read from the data file
type Entry struct {
	Region string
	Cases  int
	Date   time.Time
	Deaths int
}

type dataReader interface {
	read(fileURL string) (Regions, error)
	supports(fileURL string) bool
}

var readers = []dataReader{XLSXReader{}}

const portalGeralURL = "https://xx9p7hp1p7.execute-api.us-east-1.amazonaws.com/prod/PortalGeralApi"

var headers = map[string]string{
	"Pragma":                 "no-cache",
	"Cache-Control":          "no-cache",
	"X-Parse-Application-Id": "unAFkcaNDeXajurGB7LChj8SgQYS2ptm",
}

type file struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type result struct {
	File      file      `json:"arquivo"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type results struct {
	Results []result `json:"results"`
}

type spreadSheet struct {
	File      file      `json:"arquivo"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type portalGeralAPI struct {
	SpreadSheet spreadSheet `json:"planilha"`
}

func GetMetaData() result {
	req, _ := http.NewRequest("GET", portalGeralURL, nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, _ := http.DefaultClient.Do(req)
	data, _ := ioutil.ReadAll(resp.Body)
	var portalGeralAPI portalGeralAPI
	json.Unmarshal(data, &portalGeralAPI)
	return result{
		File:      portalGeralAPI.SpreadSheet.File,
		UpdatedAt: portalGeralAPI.SpreadSheet.UpdatedAt,
	}
}

// ReadData checks the file extensions ans uses the correct
// method to read the data.
func ReadData(fileURL string) (Regions, error) {
	for _, reader := range readers {
		if reader.supports(fileURL) {
			return reader.read(fileURL)
		}
	}

	return nil, errors.New("unsupported data file")
}
