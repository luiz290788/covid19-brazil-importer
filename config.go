package covid19brazilimporter

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

type properties struct {
	Filename      string
	Repo          string
	RepoOwner     string
	Branch        string
	CommitMessage string
	CommiterName  string
	CommiterEmail string
}

const projectID = "covid19-radar-276100"

func getProperties(ctx context.Context, firestoreClient *firestore.Client) (*properties, error) {
	doc, getDocError := firestoreClient.Collection("Config").Doc("properties").Get(ctx)
	if getDocError != nil {
		return nil, getDocError
	}

	p := &properties{}
	doc.DataTo(p)

	return p, nil
}

func getLastUpdate(ctx context.Context, firestoreClient *firestore.Client) (*time.Time, error) {
	doc, getDocError := firestoreClient.Collection("Config").Doc("lastUpdate").Get(ctx)
	if getDocError != nil {
		return nil, getDocError
	}

	possibleLastUpdate, found := doc.Data()["lastUpdate"]
	var lastUpdate time.Time
	if !found {
		lastUpdate = time.Unix(0, 0)
	} else {
		lastUpdate = possibleLastUpdate.(time.Time)
	}

	return &lastUpdate, nil
}

func setLastUpdate(ctx context.Context, firestoreClient *firestore.Client) (err error) {
	_, err = firestoreClient.Collection("Config").Doc("lastUpdate").Set(ctx, map[string]interface{}{
		"lastUpdate": time.Now(),
	})

	return
}
