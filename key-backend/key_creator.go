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

	"cloud.google.com/go/firestore"
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
	ctx := context.Background()
	ec := NewErrorReportingClient(ctx, projectName, serviceName)
	defer ec.Close()

	fs, err := NewFirestoreClient(ctx, projectName)
	if err != nil {
		logAndWriteError(ec, w, "failed to create firestore client", err)
	}
	defer fs.Close()

	var req requestMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logAndWriteError(ec, w, "failed to decode json", err)
		return
	}

	if req.Requester == "" {
		logAndWriteError(ec, w, "no requester supplied", fmt.Errorf("no requester supplied"))
		return
	}

	key, err := generateKey()
	if err != nil {
		logAndWriteError(ec, w, "failed to generate key", err)
		return
	}

	if err := writeKey(ctx, fs, key, req.Requester); err != nil {
		logAndWriteError(ec, w, "failed to write key to datastore", err)
		return
	}

	res := responseMessage{
		ResultStatus: "success",
		Key:          key,
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		logAndWriteError(ec, w, "failed to encode response", err)
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

func writeKey(ctx context.Context, client *firestore.Client, key, requester string) error {
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
