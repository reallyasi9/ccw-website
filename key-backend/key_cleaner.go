package backend

import (
	"context"
	"log"
	"time"

	"google.golang.org/api/iterator"
)

// PubSubMessage is the payload of a pub/sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// CleanKeys is pub/sub Cloud Function that cleans expired upload keys.
func CleanKeys(ctx context.Context, m PubSubMessage) {

	client, err := NewFirestoreClient(context.Background(), projectName)
	if err != nil {
		logError(err)
		return
	}
	defer client.Close()

	expiryItr := client.Collection("keys").Where("expiry", "<=", time.Now()).Documents(ctx)
	delBatch := client.Batch()
	irec := 0
	for {
		doc, err := expiryItr.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logError(err)
		}
		log.Printf("scheduling key %s for deletion\n", doc.Ref.ID)
		delBatch.Delete(doc.Ref)
		irec++
		if irec%500 == 0 {
			_, err := delBatch.Commit(ctx)
			if err != nil {
				logError(err)
			}
			delBatch = client.Batch()
		}
	}
	if irec%500 != 0 {
		_, err := delBatch.Commit(ctx)
		if err != nil {
			logError(err)
		}
	}
}
