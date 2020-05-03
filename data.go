package covid19brazilimporter

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

const portalGeralURL = "https://xx9p7hp1p7.execute-api.us-east-1.amazonaws.com/prod/PortalGeral"

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

func getMetaData() result {
	req, _ := http.NewRequest("GET", portalGeralURL, nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, _ := http.DefaultClient.Do(req)
	data, _ := ioutil.ReadAll(resp.Body)
	var results results
	json.Unmarshal(data, &results)
	return results.Results[0]
}

func getData(fileURL string) *csv.Reader {
	resp, _ := http.Get(fileURL)
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	return reader
}

func readData(fileURL string) chan []string {
	dataChannel := make(chan []string)

	go func() {
		reader := getData(fileURL)
		for line, _ := reader.Read(); line != nil; line, _ = reader.Read() {
			dataChannel <- line
		}
		close(dataChannel)
	}()

	return dataChannel
}
