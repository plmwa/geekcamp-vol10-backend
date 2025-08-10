package database

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

const HostEnvKey = "FIRESTORE_EMULATOR_HOST" // URLlocalhost:8080
const Project = "my-first-firebase-57a45"    // projectID

// Firestoreクライアント生成
func NewFirestoreClient(ctx context.Context) *firestore.Client {
	_, url := os.LookupEnv(HostEnvKey)
	if !url {
		log.Fatalf("required env %s", HostEnvKey)
	}

	client, err := firestore.NewClient(ctx, Project)
	if err != nil {
		log.Fatalf("Firestoreクライアントでエラー: %v", err)
	}
	return client
}
