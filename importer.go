package covid19brazilimporter

import (
	"bytes"
	"context"
	"encoding/csv"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

func buildFileContents(fileURL string) []byte {
	dataChannel := readData(fileURL)

	buffer := bytes.NewBuffer(make([]byte, 0))
	writer := csv.NewWriter(buffer)

	writer.Write([]string{"Date", "Region", "Cases", "Deaths"})

	// ignore header
	<-dataChannel

	for line := range dataChannel {
		date, _ := parseDate(line[2])
		newLine := []string{serializeDate(date), line[1], line[4], line[6]}
		writer.Write(newLine)
	}

	writer.Flush()
	return buffer.Bytes()
}

// PubSubMessage is the struct that define the received through the PubSub trigger
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// ImportData is the main function that will trigger the import of data
func ImportData(ctx context.Context, message PubSubMessage) error {
	firestoreClient, firestoreError := firestore.NewClient(ctx, projectID)
	defer firestoreClient.Close()
	if firestoreError != nil {
		return firestoreError
	}
	metadata := getMetaData()

	lastUpdate, lastUpdateError := getLastUpdate(ctx, firestoreClient)
	if lastUpdateError != nil {
		return lastUpdateError
	}

	if lastUpdate.After(metadata.UpdatedAt) {
		log.Println("already updated")
		return nil
	}

	properties, propertiesError := getProperties(ctx, firestoreClient)
	if propertiesError != nil {
		return propertiesError
	}
	client := buildGithubClient(ctx, os.Getenv("GITHUB_TOKEN"))
	updateFile(ctx, client, properties, buildFileContents(metadata.File.URL))
	setLastUpdate(ctx, firestoreClient)
	log.Println("update done")
	return nil
}
