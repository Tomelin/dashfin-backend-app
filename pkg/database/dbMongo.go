package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB implements the DatabaseService interface for MongoDB.
type MongoDB struct {
	client         *mongo.Client
	database       *mongo.Database
	collection     *mongo.Collection
	collectionName string // Store collection name from config
	ctx            context.Context
	cancelCtx      context.CancelFunc // To manage context lifecycle
}

// InitializeMongoDB creates and initializes a new MongoDB instance.
// It doesn't connect immediately; connection happens in the Connect method.
func InitializeMongoDB() (*MongoDB, error) {
	// A cancellable context for operations
	ctx, cancel := context.WithCancel(context.Background())
	return &MongoDB{
		ctx:       ctx,
		cancelCtx: cancel,
	}, nil
}

// Connect establishes a connection to MongoDB.
// config is expected to be a MongoConfig struct.
func (db *MongoDB) Connect(config interface{}) error {
	cfg, ok := config.(MongoConfig)
	if !ok {
		return fmt.Errorf("invalid config type for MongoDB: expected MongoConfig")
	}

	if cfg.ConnectionString == "" {
		return fmt.Errorf("ConnectionString is required for MongoDB")
	}
	if cfg.DatabaseName == "" {
		return fmt.Errorf("DatabaseName is required for MongoDB")
	}
	if cfg.CollectionName == "" {
		return fmt.Errorf("CollectionName is required for MongoDB")
	}

	clientOptions := options.Client().ApplyURI(cfg.ConnectionString)
	client, err := mongo.Connect(db.ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("mongo.Connect: %w", err)
	}

	// Ping the primary to verify connection.
	// Use a timeout for the ping.
	pingCtx, cancelPing := context.WithTimeout(db.ctx, 5*time.Second)
	defer cancelPing()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		client.Disconnect(db.ctx) // Disconnect if ping fails
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}

	db.client = client
	db.database = client.Database(cfg.DatabaseName)
	db.collection = db.database.Collection(cfg.CollectionName)
	db.collectionName = cfg.CollectionName // Store for potential later use if needed
	log.Printf("Successfully connected to MongoDB, database: %s, collection: %s", cfg.DatabaseName, cfg.CollectionName)
	return nil
}

// Get retrieves a single document by its ID.
func (db *MongoDB) Get(id string) (interface{}, error) {
	if db.collection == nil {
		return nil, fmt.Errorf("MongoDB collection not initialized. Call Connect first.")
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format for MongoDB: %w", err)
	}

	var result bson.M // Using bson.M for flexibility, could be a specific struct
	err = db.collection.FindOne(db.ctx, bson.M{"_id": objectID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("document not found with ID %s: %w", id, err)
		}
		return nil, fmt.Errorf("MongoDB FindOne error: %w", err)
	}
	return result, nil
}

// Create adds a new document to the collection.
func (db *MongoDB) Create(data interface{}) (interface{}, error) {
	if db.collection == nil {
		return nil, fmt.Errorf("MongoDB collection not initialized. Call Connect first.")
	}
	insertResult, err := db.collection.InsertOne(db.ctx, data)
	if err != nil {
		return nil, fmt.Errorf("MongoDB InsertOne error: %w", err)
	}
	// Return the ID of the inserted document
	return insertResult.InsertedID, nil
}

// Update modifies an existing document identified by its ID.
func (db *MongoDB) Update(id string, data interface{}) (interface{}, error) {
	if db.collection == nil {
		return nil, fmt.Errorf("MongoDB collection not initialized. Call Connect first.")
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format for MongoDB: %w", err)
	}

	// Using bson.M{"$set": data} to update specified fields.
	// For more complex updates, the 'data' structure would need to reflect MongoDB update operators.
	updateResult, err := db.collection.UpdateOne(db.ctx, bson.M{"_id": objectID}, bson.M{"$set": data})
	if err != nil {
		return nil, fmt.Errorf("MongoDB UpdateOne error: %w", err)
	}
	if updateResult.MatchedCount == 0 {
		return nil, fmt.Errorf("no document found with ID %s to update", id)
	}
	// Optionally, one could fetch and return the updated document.
	// For now, returning the input data upon successful update or nil.
	return data, nil // Or just return nil if not fetching the document
}

// Delete removes a document from the collection by its ID.
func (db *MongoDB) Delete(id string) error {
	if db.collection == nil {
		return fmt.Errorf("MongoDB collection not initialized. Call Connect first.")
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid ID format for MongoDB: %w", err)
	}
	deleteResult, err := db.collection.DeleteOne(db.ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("MongoDB DeleteOne error: %w", err)
	}
	if deleteResult.DeletedCount == 0 {
		return fmt.Errorf("no document found with ID %s to delete", id)
	}
	return nil
}

// GetByFilter retrieves multiple documents based on a set of filters.
// Filters map keys are field names, values are the criteria.
func (db *MongoDB) GetByFilter(filters map[string]interface{}) ([]interface{}, error) {
	if db.collection == nil {
		return nil, fmt.Errorf("MongoDB collection not initialized. Call Connect first.")
	}

	// Convert map[string]interface{} to bson.M for the MongoDB driver
	bsonFilters := bson.M{}
	for k, v := range filters {
		bsonFilters[k] = v
	}

	cursor, err := db.collection.Find(db.ctx, bsonFilters)
	if err != nil {
		return nil, fmt.Errorf("MongoDB Find error: %w", err)
	}
	defer cursor.Close(db.ctx)

	var results []interface{}
	// It's generally better to decode into a specific struct slice (e.g., []MyStruct)
	// but using []bson.M for flexibility here to match interface{}.
	// Or []map[string]interface{}.
	for cursor.Next(db.ctx) {
		var elem bson.M
		if err := cursor.Decode(&elem); err != nil {
			log.Printf("Error decoding document in GetByFilter: %v", err)
			// Decide if one error should stop the whole process or just skip the doc
			continue
		}
		results = append(results, elem)
	}
	if err := cursor.Err(); err != nil {
		return results, fmt.Errorf("MongoDB cursor error: %w", err)
	}
	return results, nil
}

// Close terminates the MongoDB connection and cancels the context.
func (db *MongoDB) Close() error {
	if db.client != nil {
		// Call cancel on the context used by this DB instance
		if db.cancelCtx != nil {
			db.cancelCtx()
		}
		err := db.client.Disconnect(db.ctx) // Use the original context for Disconnect
		if err != nil {
			return fmt.Errorf("error disconnecting MongoDB client: %w", err)
		}
		db.client = nil
		db.database = nil
		db.collection = nil
		log.Println("MongoDB connection closed.")
		return nil
	}
	return fmt.Errorf("MongoDB client not initialized or already closed")
}
