package backend

import (
	"context"
	"log"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

// NewFirestoreClient creates a Firestore client for the context and project using boilerplate code.
func NewFirestoreClient(ctx context.Context, projectID string) (*firestore.Client, error) {
	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return nil, err
	}

	return app.Firestore(ctx)
}

// NewErrorReportingClient creates an ErrorReporting client for the context and project using boilerplate code.
func NewErrorReportingClient(ctx context.Context, projectID, serviceName string) *errorreporting.Client {
	client, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
		ServiceName: serviceName,
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	return client
}
