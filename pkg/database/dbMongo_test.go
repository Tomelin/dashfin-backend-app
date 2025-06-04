package database

import (
	"os"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Helper function to get MongoConfig from environment variables for testing
func getTestMongoConfig() (MongoConfig, bool) {
	connStr := os.Getenv("MONGO_CONNECTION_STRING")
	dbName := os.Getenv("MONGO_DATABASE_NAME")
	collName := os.Getenv("MONGO_COLLECTION_NAME")

	if connStr == "" || dbName == "" || collName == "" {
		return MongoConfig{}, false
	}
	return MongoConfig{
		ConnectionString: connStr,
		DatabaseName:     dbName,
		CollectionName:   collName,
	}, true
}

// TestInitializeMongoDB tests the initialization of MongoDB.
func TestInitializeMongoDB(t *testing.T) {
	db, err := InitializeMongoDB()
	if err != nil {
		t.Fatalf("InitializeMongoDB() error = %v", err)
	}
	if db == nil {
		t.Fatalf("InitializeMongoDB() returned nil db instance")
	}
	// Check if context and cancelFunc are initialized
	if db.ctx == nil {
		t.Error("InitializeMongoDB() did not initialize context")
	}
	if db.cancelCtx == nil {
		t.Error("InitializeMongoDB() did not initialize cancelCtx")
	}
	t.Log("InitializeMongoDB successful.")
}

// TestMongoDBConnect tests the Connect method of MongoDB.
func TestMongoDBConnect(t *testing.T) {
	db, _ := InitializeMongoDB()
	defer db.Close() // Ensure context is cancelled even if test fails early

	// Test with invalid config type
	err := db.Connect("not a mongo config")
	if err == nil {
		t.Errorf("Connect() with invalid config type, expected error, got nil")
	}

	// Test with valid config struct but missing ConnectionString
	err = db.Connect(MongoConfig{DatabaseName: "test", CollectionName: "test"})
	if err == nil {
		t.Errorf("Connect() with empty ConnectionString, expected error, got nil")
	}

	cfg, আছে := getTestMongoConfig()
	if !আছে {
		t.Skip("MONGO_CONNECTION_STRING, MONGO_DATABASE_NAME, or MONGO_COLLECTION_NAME not set, skipping real connection test for MongoDB.")
	}

	err = db.Connect(cfg)
	if err != nil {
		t.Fatalf("Connect() with valid config failed: %v", err)
	}
	if db.client == nil || db.database == nil || db.collection == nil {
		t.Errorf("Connect() did not properly initialize client, database, or collection")
	}
	t.Log("MongoDBConnect successful.")
}

// TestMongoDBOperations is a placeholder for CRUD tests.
// It requires a running MongoDB instance.
func TestMongoDBOperations(t *testing.T) {
	cfg, আছে := getTestMongoConfig()
	if !আছে {
		t.Skip("MongoDB environment variables not set, skipping CRUD tests.")
	}

	db, err := InitializeMongoDB()
	if err != nil {
		t.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer db.Close()

	if err := db.Connect(cfg); err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// --- Test Create ---
	type TestItem struct {
		Name  string `bson:"name"`
		Value int    `bson:"value"`
	}
	itemToCreate := TestItem{Name: "Test Item " + time.Now().String(), Value: 123}
	insertedID, err := db.Create(itemToCreate)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	if insertedID == nil {
		t.Fatal("Create() returned nil ID")
	}
	docID, ok := insertedID.(primitive.ObjectID)
	if !ok {
		t.Fatalf("Create() returned ID of unexpected type: %T", insertedID)
	}
	t.Logf("Create() successful, ID: %s", docID.Hex())

	// --- Test Get ---
	retrievedItem, err := db.Get(docID.Hex())
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if retrievedItem == nil {
		t.Fatal("Get() returned nil item")
	}
	retrievedMap, ok := retrievedItem.(bson.M)
	if !ok {
		t.Fatalf("Get() returned item of unexpected type: %T", retrievedItem)
	}
	if retrievedMap["name"] != itemToCreate.Name {
		t.Errorf("Get() retrieved name = %v, want %v", retrievedMap["name"], itemToCreate.Name)
	}
	t.Logf("Get() successful: %+v", retrievedMap)

	// --- Test Update ---
	updateData := bson.M{"value": 456, "updated": true}
	updatedItem, err := db.Update(docID.Hex(), updateData)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	if updatedItem == nil { // Assuming Update returns the data map passed in on success
		t.Fatal("Update() returned nil item")
	}
	// Verify by Getting again
	retrievedAfterUpdate, _ := db.Get(docID.Hex())
	retrievedMapAfterUpdate := retrievedAfterUpdate.(bson.M)
	if retrievedMapAfterUpdate["value"] != int32(456) || retrievedMapAfterUpdate["updated"] != true { // BSON numbers might be int32
		t.Errorf("Update() did not correctly update item. Got: %+v, value type: %T", retrievedMapAfterUpdate, retrievedMapAfterUpdate["value"])
	}
	t.Logf("Update() successful.")


	// --- Test GetByFilter ---
	filter := map[string]interface{}{"name": itemToCreate.Name}
	filteredItems, err := db.GetByFilter(filter)
	if err != nil {
		t.Fatalf("GetByFilter() failed: %v", err)
	}
	if len(filteredItems) != 1 {
		t.Fatalf("GetByFilter() expected 1 item, got %d", len(filteredItems))
	}
	t.Logf("GetByFilter() successful, found %d item(s).", len(filteredItems))


	// --- Test Delete ---
	err = db.Delete(docID.Hex())
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}
	// Verify by trying to Get again
	_, err = db.Get(docID.Hex())
	if err == nil {
		t.Errorf("Get() after Delete() should have failed, but it succeeded.")
	} else if !strings.Contains(err.Error(), "document not found") { // check specific error if possible
        t.Errorf("Get() after Delete() failed with unexpected error: %v", err)
    }
	t.Log("Delete() successful.")
}


func TestMongoDBClose(t *testing.T) {
	db, _ := InitializeMongoDB()
	// Test close on non-connected client (should be graceful)
	err := db.Close()
	// Allow specific error for not being initialized, or no error if already closed by init
	if err != nil && !strings.Contains(err.Error(), "MongoDB client not initialized or already closed") {
		t.Errorf("Close() on non-connected client returned unexpected error: %v", err)
	}


	cfg, আছে := getTestMongoConfig()
	if !আছে {
		t.Skip("MongoDB environment variables not set, skipping connected Close() test.")
	}

	// Re-initialize for this specific test part to ensure clean state
	db, _ = InitializeMongoDB()
	defer func() { // Ensure this deferred close doesn't panic if db is nil
		if db != nil {
			db.Close()
		}
	}()


	if errConnect := db.Connect(cfg); errConnect == nil {
		errClose := db.Close()
		if errClose != nil {
			t.Errorf("Close() after successful connect failed: %v", errClose)
		}
		// Try to use client after close (should fail or client should be nil)
		if db.client != nil {
			pingCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			// Ping should fail as client is disconnected.
			// mongo.ErrClientDisconnected is the typical error.
			if errPing := db.client.Ping(pingCtx, readpref.Primary()); errPing == nil {
				t.Error("MongoDB client Ping succeeded after Close(), expected error.")
			} else {
				t.Logf("Ping after close failed as expected: %v", errPing)
			}
		} else {
			t.Log("Client is nil after close, as expected.")
		}
	} else {
		t.Logf("Skipping connected Close() test as connection failed: %v", errConnect)
	}
	t.Log("TestMongoDBClose finished.")
}
