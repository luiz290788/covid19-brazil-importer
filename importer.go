package covid19brazilimporter

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

func buildFileContents(fileURL string) ([]byte, error) {
	dataChannel, err := ReadData(fileURL)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(make([]byte, 0))
	writer := csv.NewWriter(buffer)

	writer.Write([]string{"Date", "Region", "Cases", "Deaths"})

	for newEntry := range dataChannel {
		newLine := []string{serializeDate(newEntry.Date), newEntry.Region,
			fmt.Sprintf("%d", newEntry.Cases), fmt.Sprintf("%d", newEntry.Deaths)}
		writer.Write(newLine)
	}

	writer.Flush()
	return buffer.Bytes(), nil
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
		log.Panicln(firestoreError.Error())
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

	fileData, contentError := buildFileContents(metadata.File.URL)

	if contentError != nil {
		log.Panicln(contentError.Error())
		return contentError
	}

	properties, propertiesError := getProperties(ctx, firestoreClient)
	if propertiesError != nil {
		return propertiesError
	}
	client := buildGithubClient(ctx, os.Getenv("GITHUB_TOKEN"))
	_, updateFileError := updateFile(ctx, client, properties, fileData)
	if updateFileError != nil {
		log.Panicln(updateFileError.Error())
		return updateFileError
	}
	setLastUpdate(ctx, firestoreClient)
	log.Println("update done")
	return nil
}
