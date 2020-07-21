package covid19brazilimporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-git/go-git/v5"
)

// PubSubMessage is the struct that define the received through the PubSub trigger
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type LastUpdated struct {
	Date time.Time `json:"date"`
}

func writeToFile(filepath string, data interface{}) (err error) {
	var file *os.File
	file, err = os.Create(filepath)
	if err != nil {
		return
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(data)
	return file.Close()
}

func dumpRegions(folder string, regions Regions) (err error) {
	for key, value := range regions {
		filepath := fmt.Sprintf("%s/%d.json", folder, key)
		err = writeToFile(filepath, value)
		if err != nil {
			return
		}
	}
	return
}

func dumpIndexes(folder string, pages []*Page) error {
	return writeToFile(fmt.Sprintf("%s/page-index.json", folder), pages)
}

func dumpDataLists(folder string, lists map[int][]*DataListing) (err error) {
	for key, value := range lists {
		filepath := fmt.Sprintf("%s/list-%d.json", folder, key)
		err = writeToFile(filepath, value)
		if err != nil {
			return
		}
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

	dataFolder := fmt.Sprintf("%s/data", folder)
	os.MkdirAll(dataFolder, 0700)

	pages, lists, indexesError := buildIndexes(regions)
	if indexesError != nil {
		return indexesError
	}

	err = dumpRegions(dataFolder, regions)
	if err != nil {
		return err
	}

	err = dumpIndexes(dataFolder, pages)
	if err != nil {
		return err
	}

	err = dumpDataLists(dataFolder, lists)
	if err != nil {
		return err
	}

	err = writeToFile(fmt.Sprintf("%s/last-updated.json", dataFolder), &LastUpdated{Date: time.Now()})
	if err != nil {
		return err
	}

	_, err = commitAll(repository, properties.CommitMessage, properties.CommiterName, properties.CommiterEmail)
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
