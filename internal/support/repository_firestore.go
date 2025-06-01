package support

import (
	"context"
	"errors" // Added for nil client check
	"log"
	"time"

	"cloud.google.com/go/firestore"
	// "google.golang.org/api/iterator" // If you were to implement read operations
)

const (
	firestoreCollection = "support_requests"
)

// RepositoryFirestore handles persistence of SupportRequest data in Firestore.
type RepositoryFirestore struct {
	client *firestore.Client
}

// NewRepositoryFirestore creates a new Firestore repository for support requests.
func NewRepositoryFirestore(client *firestore.Client) *RepositoryFirestore {
	return &RepositoryFirestore{client: client}
}

// FirestoreSupportRequest is a representation of the data stored in Firestore.
// It might include additional fields like timestamps or user identifiers.
type FirestoreSupportRequest struct {
	Category    string    `firestore:"category"`
	Description string    `firestore:"description"`
	UserID      string    `firestore:"user_id"` // Firebase UID
	AppName     string    `firestore:"app_name"`
	CreatedAt   time.Time `firestore:"created_at,serverTimestamp"`
	// Add other fields as necessary, e.g., status, updated_at
}

// Save stores a new support request in Firestore.
// It returns the ID of the newly created document or an error.
func (r *RepositoryFirestore) Save(ctx context.Context, req SupportRequest, userID, appName string) (string, error) {
	if r.client == nil {
		log.Println("Firestore client is nil in RepositoryFirestore")
		return "", errors.New("firestore client is not initialized")
	}

	docData := FirestoreSupportRequest{
		Category:    req.Category,
		Description: req.Description,
		UserID:      userID,
		AppName:     appName,
		// CreatedAt will be set by Firestore using ServerTimestamp
	}

	// Add a new document with a generated ID.
	docRef, _, err := r.client.Collection(firestoreCollection).Add(ctx, docData)
	if err != nil {
		log.Printf("Error adding document to Firestore: %v\n", err)
		return "", err
	}

	log.Printf("Support request saved to Firestore with ID: %s", docRef.ID)
	return docRef.ID, nil
}
