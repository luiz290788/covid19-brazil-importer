package covid19brazilimporter

import (
	"context"
	"testing"
)

func TestImporter(t *testing.T) {
	ImportData(context.Background(), PubSubMessage{})
}
