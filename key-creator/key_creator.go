package keycreator

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/errorreporting"
	firebase "firebase.google.com/go"
)

type requestMessage struct {
	Requester string `json:"requester"`
}

type responseMessage struct {
	ResultStatus string `json:"resultStatus"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Key          string `json:"key,omitempty"`
}

func init() {
	alphabetLength = big.NewInt(int64(len(alphabet)))
}

var serviceName = "keycreator"

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

// CreateKey is an HTTP Cloud Function that generates an upload key.
func CreateKey(w http.ResponseWriter, r *http.Request) {
	var req requestMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logAndPrintError(w, "failed to decode json", err)
		return
	}

	key, err := generateKey()
	if err != nil {
		logAndPrintError(w, "failed to generate key", err)
		return
	}

	if err := writeKey(key, req.Requester); err != nil {
		logAndPrintError(w, "failed to write key to datastore", err)
		return
	}

	res := responseMessage{
		ResultStatus: "success",
		Key:          key,
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		logAndPrintError(w, "failed to encode response", err)
		return
	}
}

var alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var alphabetLength *big.Int

const keyLength = 8

func generateKey() (string, error) {
	var sb strings.Builder
	for i := 0; i < keyLength; i++ {
		idx, err := rand.Int(rand.Reader, alphabetLength)
		if err != nil {
			return "", err
		}
		sb.WriteByte(alphabet[idx.Int64()])
	}
	return sb.String(), nil
}

type keyDocument struct {
	Key       string    `firestore:"key"`
	Requester string    `firestore:"requester"`
	Expiry    time.Time `firestore:"expiry"`
}

func writeKey(key, requester string) error {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: os.Getenv("GCP_PROJECT")}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	doc := keyDocument{
		Key:       key,
		Requester: requester,
		Expiry:    time.Now().AddDate(0, 0, 1),
	}
	_, _, err2 := client.Collection("keys").Add(ctx, doc)
	if err2 != nil {
		return err2
	}

	return nil
}
