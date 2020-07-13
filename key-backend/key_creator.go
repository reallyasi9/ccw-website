package backend

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"
)

type requestMessage struct {
	Requester string `json:"requester"`
}

type responseMessage struct {
	ResultStatus string `json:"resultStatus"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Key          string `json:"key,omitempty"`
}

var serviceName string
var projectName string
var alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var alphabetLength *big.Int

const keyLength = 8

const expiryDays = 7

func init() {
	alphabetLength = big.NewInt(int64(len(alphabet)))
	serviceName = os.Getenv("FUNCTION_NAME")
	projectName = os.Getenv("GCP_PROJECT")
}

// CreateKey is an HTTP Cloud Function that generates an upload key.
func CreateKey(w http.ResponseWriter, r *http.Request) {
	var req requestMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logAndPrintError(w, serviceName, "failed to decode json", err)
		return
	}

	if req.Requester == "" {
		logAndPrintError(w, serviceName, "no requester supplied", fmt.Errorf("no requester supplied"))
		return
	}

	key, err := generateKey()
	if err != nil {
		logAndPrintError(w, serviceName, "failed to generate key", err)
		return
	}

	if err := writeKey(key, req.Requester); err != nil {
		logAndPrintError(w, serviceName, "failed to write key to datastore", err)
		return
	}

	res := responseMessage{
		ResultStatus: "success",
		Key:          key,
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		logAndPrintError(w, serviceName, "failed to encode response", err)
		return
	}
}

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
	client, err := NewFirestoreClient(ctx, projectName)
	if err != nil {
		return err
	}
	defer client.Close()

	doc := keyDocument{
		Key:       key,
		Requester: requester,
		Expiry:    time.Now().AddDate(0, 0, expiryDays),
	}
	_, _, err2 := client.Collection("keys").Add(ctx, doc)
	if err2 != nil {
		return err2
	}

	return nil
}
