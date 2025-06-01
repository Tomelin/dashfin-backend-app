package firestore

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	// For local development with a specific project ID, you might need:
	// "google.golang.org/api/option"
)

// Client wraps the Firestore client.
type Client struct {
	*firestore.Client
}

// NewClient initializes a new Firestore client.
// It expects GOOGLE_APPLICATION_CREDENTIALS environment variable to be set
// or for the application to be running on Google Cloud infrastructure.
// projectID is your Google Cloud Project ID.
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	// Ensure projectID is provided, especially for local/non-GCP env.
	if projectID == "" {
		log.Println("Warning: Firestore projectID is empty. Client initialization might rely on environment.")
		// Consider returning an error if projectID is strictly required.
	}

	// For local development, you might use:
	// sa := option.WithCredentialsFile("path/to/your/serviceAccountKey.json")
	// client, err := firestore.NewClient(ctx, projectID, sa)
	// For now, we assume credentials are set up via environment for broader compatibility.
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Error initializing Firestore client for project %s: %v\n", projectID, err)
		return nil, err
	}
	return &Client{client}, nil
}
