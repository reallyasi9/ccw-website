package backend

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/errorreporting"
)

func logAndPrintError(w http.ResponseWriter, serviceName, msg string, err error) {
	client := NewErrorReportingClient(context.Background(), projectName, serviceName)
	defer errorClient.Close()

	client.Report(errorreporting.Entry{
		Error: err,
	})
	res := responseMessage{
		ResultStatus: "error",
		ErrorMessage: msg,
	}
	w.WriteHeader(http.StatusInternalServerError)
	if err2 := json.NewEncoder(w).Encode(res); err2 != nil {
		log.Fatal(err2)
	}
}
