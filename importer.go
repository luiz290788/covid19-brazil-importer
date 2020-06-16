package covid19brazilimporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/go-git/go-git/v5"
)

// PubSubMessage is the struct that define the received through the PubSub trigger
type PubSubMessage struct {
	Data []byte `json:"data"`
}

func dumpRegions(folder string, regions Regions) (err error) {
	os.MkdirAll(folder, 0700)
	for key, value := range regions {
		var file *os.File
		filename := fmt.Sprintf("%s/%d.json", folder, key)
		file, err = os.Create(filename)
		if err != nil {
			return
		}
		encoder := json.NewEncoder(file)
		encoder.Encode(value)
		file.Close()
	}
	return
}

// ImportData is the main function that will trigger the import of data
func ImportData(ctx context.Context, message PubSubMessage) error {
	firestoreClient, firestoreError := firestore.NewClient(ctx, projectID)
	defer firestoreClient.Close()
	if firestoreError != nil {
		log.Panicln(firestoreError.Error())
		return firestoreError
	}
	metadata := GetMetaData()

	lastUpdate, lastUpdateError := getLastUpdate(ctx, firestoreClient)
	if lastUpdateError != nil {
		return lastUpdateError
	}

	if lastUpdate.After(metadata.UpdatedAt) {
		log.Println("already updated")
		return nil
	}

	fmt.Println(metadata.File.URL)

	properties, propertiesError := getProperties(ctx, firestoreClient)
	if propertiesError != nil {
		return propertiesError
	}

	var regions Regions
	var err error
	regions, err = ReadData(metadata.File.URL)
	if err != nil {
		return err
	}

	folder, _ := ioutil.TempDir("", "regions-*")
	fmt.Println(folder)

	repository, _ := cloneRepo(folder, os.Getenv("GITHUB_TOKEN"), properties.Branch)

	err = dumpRegions(fmt.Sprintf("%s/src/data/covid", folder), regions)
	if err != nil {
		return err
	}

	_, err = commitAll(repository, properties.CommitMessage, properties.CommiterName,
		properties.CommiterEmail)
	if err != nil {
		return err
	}

	err = repository.Push(&git.PushOptions{})
	if err != nil {
		return err
	}

	setLastUpdate(ctx, firestoreClient)
	log.Println("update done")

	return nil
}
