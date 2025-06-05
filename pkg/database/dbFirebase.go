package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FirebaseDBInterface interface {
	Get(ctx context.Context, collection string) ([]byte, error)
	Create(ctx context.Context, data interface{}, collection string) ([]byte, error)
	Update(ctx context.Context, id string, data interface{}, collection string) ([]byte, error)
	Delete(ctx context.Context, id, collection string) error
	GetByFilter(ctx context.Context, filters map[string]interface{}, collection string) ([]byte, error)
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
func (db *FirebaseDB) Get(ctx context.Context, collection string) ([]byte, error) {
	// Placeholder: A real implementation would specify the collection.
	if db.client == nil {
		return nil, errors.New("firestore client not initialized. Call Connect first")
	}

	if err := db.validateWithoutData(ctx, collection); err != nil {
		return nil, err
	}

	doc, err := db.client.Collection(collection).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var results []interface{}
	for _, d := range doc {
		results = append(results, d.Data())
	}

	b, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Create adds a new document to a default collection.
// Placeholder: Collection name needed.
func (db *FirebaseDB) Create(ctx context.Context, data interface{}, collection string) ([]byte, error) {

	if err := db.validateWithData(ctx, data, collection); err != nil {
		return nil, err
	}

	colRef := db.client.Collection(collection)
	docRef, _, err := colRef.Add(ctx, data)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(docRef)
	if err != nil {
		return nil, err
	}

	return b, err
}

// Update modifies an existing document in a default collection.
// Placeholder: Collection name needed.
func (db *FirebaseDB) Update(ctx context.Context, id string, data interface{}, collection string) ([]byte, error) {
	if db.client == nil {
		return nil, fmt.Errorf("firestore client not initialized. Call Connect first")
	}

	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}

	if err := db.validateWithData(ctx, data, collection); err != nil {
		return nil, err
	}

	result, err := db.client.Collection(collection).Doc(id).Set(ctx, data, firestore.MergeAll)
	if err != nil {
		return nil, err
	}
	log.Println("-------result------")
	log.Println("result", result)
	b, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Delete removes a document from a default collection by its ID.
// Placeholder: Collection name needed.
func (db *FirebaseDB) Delete(ctx context.Context, id, collection string) error {
	if db.client == nil {
		return fmt.Errorf("firestore client not initialized. Call Connect first")
	}
	if id == "" {
		return fmt.Errorf("id is empty")
	}

	if err := db.validateWithoutData(ctx, collection); err != nil {
		return err
	}

	_, err := db.client.Collection(collection).Doc(id).Delete(ctx)

	return err
}

// GetByFilter retrieves multiple documents based on a set of filters from a default collection.
// Placeholder: Collection name needed.
func (db *FirebaseDB) GetByFilter(ctx context.Context, filters map[string]interface{}, collection string) ([]byte, error) {
	if err := db.validateWithData(ctx, filters, collection); err != nil {
		return nil, err
	}

	query := db.client.Collection(collection).Query
	for key, value := range filters {
		query = query.Where(key, "==", value)
	}

	type resultObject struct {
		ID   string
		Data interface{}
	}

	objects := make([]resultObject, 0)
	iter := query.Documents(ctx)
	defer iter.Stop()
	for {
		log.Println("-------iter------")
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		objects = append(objects, resultObject{
			ID:   doc.Ref.ID,
			Data: doc.Data(),
		})
	}

	b, err := json.Marshal(objects)
	if err != nil {
		return nil, err
	}
	return b, nil
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
	return errors.New("firestore client not initialized or already closed")
}

// validateWithData
func (db *FirebaseDB) validateWithData(ctx context.Context, data interface{}, collection string) error {
	if db.client == nil {
		return errors.New("firestore client not initialized. Call Connect first")
	}

	if data == nil {
		return errors.New("data is nil")
	}

	if collection == "" {
		return errors.New("collection is empty")
	}

	if ctx.Value("Authorization") == "" {
		return errors.New("authorization token is nil")
	}
	return nil
}

// validateWithoutData
func (db *FirebaseDB) validateWithoutData(ctx context.Context, collection string) error {
	if db.client == nil {
		return errors.New("firestore client not initialized. Call Connect first")
	}

	if collection == "" {
		return errors.New("collection is empty")
	}

	if ctx.Value("Authorization") == "" {
		return errors.New("authorization token is nil")
	}
	return nil
}
