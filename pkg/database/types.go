package database

// FirebaseConfig holds the configuration for Firebase connection
type FirebaseConfig struct {
	ProjectID string `json:"project_id" yaml:"projectId"`

	// Path to the service account key file
	ServiceAccountKeyPath string `json:"serviceAccountKeyPath" yaml:"serviceAccountKeyPath"`
	// Project ID (optional if using service account key)

	// Database URL (optional, for Realtime Database)
	DatabaseURL string

	APIKey            string `json:"apiKey" yaml:"apiKey"`
	AuthDomain        string `json:"authDomain" yaml:"authDomain"`
	StorageBucket     string `json:"storageBucket" yaml:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId" yaml:"messagingSenderId"`
	AppID             string `json:"appId" yaml:"appId"`
}

// MongoConfig holds the configuration for MongoDB connection
type MongoConfig struct {
	ConnectionString string `json:"connection_string"`
	DatabaseName     string `json:"database_name"`
	CollectionName   string `json:"collection_name"`
}

// DatabaseService defines the interface for database operations.
// Implementations of this interface will provide access to different database systems.
type DatabaseService interface {
	// Connect initializes the database connection.
	// The config parameter should be a pointer to FirebaseConfig or MongoConfig.
	Connect(config interface{}) error

	// Get retrieves a single document by its ID.
	Get(id string) (interface{}, error)

	// Create adds a new document to the database.
	// The 'data' parameter should be a map or a struct that can be serialized.
	Create(data interface{}) (interface{}, error) // Returns the ID of the created document or the document itself

	// Update modifies an existing document identified by its ID.
	// The 'data' parameter should be a map or a struct containing fields to be updated.
	Update(id string, data interface{}) (interface{}, error)

	// Delete removes a document from the database by its ID.
	Delete(id string) error

	// GetByFilter retrieves multiple documents based on a set of filters.
	// Filters is a map where keys are field names and values are the criteria.
	GetByFilter(filters map[string]interface{}) ([]interface{}, error)

	// Close terminates the database connection and cleans up resources.
	Close() error
}

// General validation logic (if any) related to the interface can be added here.
// For example, functions to validate config structs, though specific validation
// might be better handled within the Connect methods of the implementations.
