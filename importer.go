package covid19brazilimporter

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

func buildFileContents(fileURL string, writer *io.PipeWriter) error {
	log.Println("Building file")
	defer writer.Close()
	dataChannel, err := ReadData(fileURL)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	csvWriter.Write([]string{"Date", "Region", "Cases", "Deaths"})
	for newEntry := range dataChannel {
		newLine := []string{serializeDate(newEntry.Date), newEntry.Region,
			fmt.Sprintf("%d", newEntry.Cases), fmt.Sprintf("%d", newEntry.Deaths)}
		csvWriter.Write(newLine)
	}

	csvWriter.Flush()
	log.Println("Finished building file")
	return nil
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

	reader, writer := io.Pipe()

	go buildFileContents(metadata.File.URL, writer)

	properties, propertiesError := getProperties(ctx, firestoreClient)
	if propertiesError != nil {
		return propertiesError
	}
	client := buildGithubClient(ctx, os.Getenv("GITHUB_TOKEN"))

	_, updateFileError := updateFile(ctx, client, properties, reader)
	if updateFileError != nil {
		log.Panicln(updateFileError.Error())
		return updateFileError
	}
	setLastUpdate(ctx, firestoreClient)
	log.Println("update done")
	return nil
}
