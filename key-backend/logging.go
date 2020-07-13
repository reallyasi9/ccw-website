package backend

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/errorreporting"
)

func logAndPrintError(w http.ResponseWriter, msg string, err error) {
	ctx := context.Background()

	errorClient, err2 := errorreporting.NewClient(ctx, os.Getenv("GCP_PROJECT"), errorreporting.Config{
		ServiceName: serviceName,
		OnError: func(err error) {
			log.Printf("Could not log error: %v", err)
		},
	})
	if err2 != nil {
		log.Fatal(err2)
	}
	defer errorClient.Close()

	errorClient.Report(errorreporting.Entry{
		Error: err,
	})
	res := responseMessage{
		ResultStatus: "error",
		ErrorMessage: msg,
	}
	w.WriteHeader(http.StatusInternalServerError)
	if err3 := json.NewEncoder(w).Encode(res); err3 != nil {
		log.Fatal(err3)
	}
}
