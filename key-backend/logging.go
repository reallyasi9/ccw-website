package backend

import (
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/errorreporting"
)

func logAndWriteError(client *errorreporting.Client, w http.ResponseWriter, msg string, err error) {
	client.Report(errorreporting.Entry{
		Error: err,
	})
	res := responseMessage{
		ResultStatus: "error",
		ErrorMessage: msg,
	}
	log.Print(err)
	w.WriteHeader(http.StatusInternalServerError)
	if err2 := json.NewEncoder(w).Encode(res); err2 != nil {
		log.Fatal(err2)
	}
}

func logError(client *errorreporting.Client, err error) {
	client.Report(errorreporting.Entry{
		Error: err,
	})
	log.Print(err)
}
