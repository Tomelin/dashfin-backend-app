package database

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type FirebaseDBInterface interface {
	Get(ctx context.Context, id string) (interface{}, error)
	Create(ctx context.Context, data interface{}, collection string) (interface{}, error)
	Update(id string, data interface{}) (interface{}, error)
	Delete(id string) error
}

// FirebaseDB implements the DatabaseService interface for Firebase Firestore.
type FirebaseDB struct {
	client *firestore.Client
}

// InitializeFirebaseDB creates and initializes a new FirebaseDB instance.
// It doesn't connect immediately; connection happens in the Connect method.
func InitializeFirebaseDB(config FirebaseConfig) (FirebaseDBInterface, error) {

	fdb := &FirebaseDB{}

	err := fdb.connect(config)
	if err != nil {
		return nil, err
	}

	return fdb, nil
}

// Connect establishes a connection to Firebase Firestore.
// config is expected to be a FirebaseConfig struct.
func (db *FirebaseDB) connect(cfg FirebaseConfig) error {

	if cfg.ProjectID == "" {
		return fmt.Errorf("ProjectID is required for Firebase connection")
	}

	var opt option.ClientOption
	if cfg.ServiceAccountKeyPath != "" {
		opt = option.WithCredentialsFile(cfg.ServiceAccountKeyPath)
	} else {
		// If no credentials file is provided, Firestore client will try to use
		// Application Default Credentials (ADC) if available.
		log.Println("Firebase CredentialsFile not provided, attempting to use Application Default Credentials.")
	}

	databaseName := "default"
	if cfg.DatabaseURL != "" {
		databaseName = cfg.DatabaseURL
	}

	// Firestore doesn't have a concept of a "database name" like traditional RDBMS.
	// Connections are made to the project ID, and then you interact with collections and documents within that project.
	client, err := firestore.NewClientWithDatabase(
		context.Background(),
		cfg.ProjectID,
		databaseName,
		opt,
	)
	if err != nil {
		return fmt.Errorf("firestore.NewClientWithDatabase: %w", err)
	}

	db.client = client
	log.Printf("Successfully connected to Firestore database: %s in project: %s", databaseName, cfg.ProjectID)

	return nil
}

// Get retrieves a single document by its ID from a default collection.
// Note: Firestore is schemaless, but often a default collection is used.
// This implementation assumes a collection name might be needed or configured elsewhere.
// For this example, let's assume Get needs a collection name.
// This highlights a potential mismatch with the generic interface if not handled carefully.
// We might need to adjust the interface or how collection names are passed.
// For now, this is a placeholder.
func (db *FirebaseDB) Get(ctx context.Context, id string) (interface{}, error) {
	// Placeholder: A real implementation would specify the collection.
	doc, err := db.client.Collection("your_collection_name").Doc(id).Get(ctx)
	if db.client == nil {
		return nil, fmt.Errorf("Firestore client not initialized. Call Connect first.")
	}
	log.Println(doc, err)
	return nil, fmt.Errorf("Get not fully implemented for Firebase: collection name needed")
}

// Create adds a new document to a default collection.
// Placeholder: Collection name needed.
func (db *FirebaseDB) Create(ctx context.Context, data interface{}, collection string) (interface{}, error) {
	if db.client == nil {
		return nil, fmt.Errorf("Firestore client not initialized. Call Connect first.")
	}

	// Placeholder:
	colRef := db.client.Collection(collection)
	log.Println("collection", colRef)
	log.Println("data", data)
	docRef, _, err := colRef.Add(ctx, data)
	log.Println("collecolRef.Addction", err, docRef)
	return docRef.ID, err
	// return nil, fmt.Errorf("Create not fully implemented for Firebase: collection name needed")
}

// Update modifies an existing document in a default collection.
// Placeholder: Collection name needed.
func (db *FirebaseDB) Update(id string, data interface{}) (interface{}, error) {
	if db.client == nil {
		return nil, fmt.Errorf("Firestore client not initialized. Call Connect first.")
	}
	// Placeholder: _, err := db.client.Collection("your_collection_name").Doc(id).Set(db.ctx, data, firestore.MergeAll)
	// return data, err
	return nil, fmt.Errorf("Update not fully implemented for Firebase: collection name needed")
}

// Delete removes a document from a default collection by its ID.
// Placeholder: Collection name needed.
func (db *FirebaseDB) Delete(id string) error {
	if db.client == nil {
		return fmt.Errorf("Firestore client not initialized. Call Connect first.")
	}
	// Placeholder: _, err := db.client.Collection("your_collection_name").Doc(id).Delete(db.ctx)
	// return err
	return fmt.Errorf("Delete not fully implemented for Firebase: collection name needed")
}

// GetByFilter retrieves multiple documents based on a set of filters from a default collection.
// Placeholder: Collection name needed.
func (db *FirebaseDB) GetByFilter(filters map[string]interface{}) ([]interface{}, error) {
	if db.client == nil {
		return nil, fmt.Errorf("Firestore client not initialized. Call Connect first.")
	}
	// Placeholder: query := db.client.Collection("your_collection_name").Query
	// for key, value := range filters {
	//	 query = query.Where(key, "==", value)
	// }
	// iter := query.Documents(db.ctx)
	// defer iter.Stop()
	// var results []interface{}
	// for {
	//	 doc, err := iter.Next()
	//	 if err == iterator.Done {
	//		 break
	//	 }
	//	 if err != nil {
	//		 return nil, err
	//	 }
	//	 results = append(results, doc.Data())
	// }
	// return results, nil
	return nil, fmt.Errorf("GetByFilter not fully implemented for Firebase: collection name needed")
}

// Close terminates the Firebase connection.
func (db *FirebaseDB) Close() error {
	if db.client != nil {
		err := db.client.Close()
		if err != nil {
			return fmt.Errorf("error closing Firestore client: %w", err)
		}
		db.client = nil
		log.Println("Firebase Firestore connection closed.")
		return nil
	}
	return fmt.Errorf("Firestore client not initialized or already closed")
}
