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
	erc := NewErrorReportingClient(ctx, projectName, serviceName)
	defer erc.Close()

	fsc, err := NewFirestoreClient(context.Background(), projectName)
	if err != nil {
		logError(erc, err)
		return
	}
	defer fsc.Close()

	expiryItr := fsc.Collection("keys").Where("expiry", "<=", time.Now()).Documents(ctx)
	delBatch := fsc.Batch()
	irec := 0
	for {
		doc, err := expiryItr.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logError(erc, err)
		}
		log.Printf("scheduling key %s for deletion\n", doc.Ref.ID)
		delBatch.Delete(doc.Ref)
		irec++
		// 500 record limit per Google's rules
		if irec%500 == 0 {
			_, err := delBatch.Commit(ctx)
			if err != nil {
				logError(erc, err)
			}
			delBatch = fsc.Batch()
		}
	}
	if irec%500 != 0 {
		_, err := delBatch.Commit(ctx)
		if err != nil {
			logError(erc, err)
		}
	}
}
