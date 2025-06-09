# Documentation: Firebase Firestore Library (dbFirebase.go)

## 1. Introduction

This document provides guidance on using the existing Firebase Firestore Go library (`pkg/database/dbFirebase.go`) for managing data, with a specific focus on a collection named "registros" (records).

The library offers functionalities to connect to Firebase Firestore and perform CRUD (Create, Read, Update, Delete) operations, as well as basic filtering. For more advanced querying capabilities not directly exposed by the library's current methods, this document will also explain how to leverage the underlying Firestore client.

**Key Design Points of the Library (Inferred):**

*   **Interface-based:** The library defines `FirebaseDBInterface` (and a more generic `DatabaseService` in `types.go`, though `dbFirebase.go` implements the former), promoting a degree of abstraction, though the current implementation is specific to Firestore.
*   **Collection Agnostic:** Most methods accept a `collection` string, allowing them to operate on any Firestore collection.
*   **Automatic Timestamps:** The `Create` and `Update` methods attempt to automatically add/update an `updatedAt` field with the server's timestamp. `Create` also sets `createdAt` if not present.
*   **ID Handling:** The `Create` method and query methods (`Get`, `GetByFilter`) attempt to include the Firestore document ID within the returned data map under the key `"id"`.
*   **Context Propagation:** Methods correctly propagate the `context.Context`.
*   **Error Handling:** Methods return errors for flow control.
*   **Simplified Filtering:** `GetByFilter` provides a simple map-based equality filter.

## 2. Setup and Initialization

To use the library, you first need to initialize `FirebaseDB`.

### 2.1. Configuration

Connection details are provided via the `FirebaseConfig` struct (defined in `pkg/database/types.go`):

```go
import "your_project_path/pkg/database"

// Example configuration
config := database.FirebaseConfig{
    ProjectID:             "your-gcp-project-id",
    ServiceAccountKeyPath: "path/to/your/serviceAccountKey.json", // Optional, uses ADC if empty
    DatabaseURL:           "your-firestore-database-url", // e.g., (default)
}
```

*   `ProjectID`: Your Google Cloud Project ID.
*   `ServiceAccountKeyPath`: Path to the JSON file containing your Firebase service account key. If not provided, the library attempts to use Application Default Credentials (ADC).
*   `DatabaseURL`: The Firestore database URL. Often, this can be the default database associated with the project.

### 2.2. Initializing the Client

Use the `InitializeFirebaseDB` function:

```go
import (
    "context"
    "log"
    "your_project_path/pkg/database"
)

func main() {
    ctx := context.Background() // Or your application's context

    config := database.FirebaseConfig{
        ProjectID: "your-gcp-project-id",
        // ServiceAccountKeyPath: "path/to/key.json", // Recommended for explicit auth
    }

    db, err := database.InitializeFirebaseDB(config)
    if err != nil {
        log.Fatalf("Failed to initialize FirebaseDB: %v", err)
    }
    log.Println("Successfully initialized FirebaseDB")

    // Ensure to close the connection when done
    // defer db.Close() // The FirebaseDB struct itself doesn't have Close(), the client inside does.
                        // The InitializeFirebaseDB returns FirebaseDBInterface.
                        // A Close() method is available on the FirebaseDB struct but not on the interface.
                        // This documentation assumes direct use of FirebaseDB or casting if needed.
                        // For now, we'll assume direct client management if needed, or that the app manages this.
}

```
**Note on `Close()`**: The `InitializeFirebaseDB` function returns a `FirebaseDBInterface`. This interface does *not* currently include a `Close()` method. The `FirebaseDB` struct *does* have a `Close()` method. To call `Close()`, you would either need to work directly with `*FirebaseDB` or type-assert the interface. For simplicity in examples, explicit closing is omitted but is crucial in production applications to release resources.

## 3. CRUD Operations for "registros"

The following examples demonstrate how to perform CRUD operations on a Firestore collection named `"registros"`.

Let's assume a `registro` (record) has the following structure:

```go
type Registro struct {
    ID          string    `json:"id,omitempty"` // Will be auto-filled by the library
    Description string    `json:"description"`
    Amount      float64   `json:"amount"`
    Date        time.Time `json:"date"`
    IsProcessed bool      `json:"isProcessed"`
    CreatedAt   time.Time `json:"createdAt,omitempty"` // Auto-filled
    UpdatedAt   time.Time `json:"updatedAt,omitempty"` // Auto-filled
}
```
The library primarily works with `map[string]interface{}` for data, but you can easily convert your structs.

### 3.1. Create a Registro

Use the `Create` method. It will automatically add `createdAt` and `updatedAt` timestamps and include the document ID in the returned data.

```go
// Assuming 'db' is your initialized FirebaseDBInterface
// and 'ctx' is your context.

registroData := map[string]interface{}{
    "description": "Salário Mensal",
    "amount":      5000.00,
    "date":        time.Now(), // Use Firestore server timestamp for more accuracy if needed
    "isProcessed": false,
}

collectionName := "registros"

createdDataBytes, err := db.Create(ctx, registroData, collectionName)
if err != nil {
    log.Printf("Error creating registro: %v", err)
    return
}

var createdRegistro map[string]interface{}
if err := json.Unmarshal(createdDataBytes, &createdRegistro); err != nil {
    log.Printf("Error unmarshalling created registro: %v", err)
    return
}

log.Printf("Registro created successfully: ID = %s, Data = %v", createdRegistro["id"], createdRegistro)
// Output will include 'id', 'createdAt', and 'updatedAt' fields.
```

### 3.2. Read Registros

#### 3.2.1. Get a Single Registro by ID

The library's `Get` method retrieves *all* documents in a collection. To get a single document by its Firestore ID, use the `GetByFilter` method, assuming the ID was stored as a field named `"id"` in the document (which this library does).

```go
// Assuming 'db' is your initialized FirebaseDBInterface, 'ctx' is your context,
// and 'registroID' is the ID of the document you want to retrieve.

registroID := "some-firestore-document-id"
collectionName := "registros"

filters := map[string]interface{}{
    "id": registroID, // Filter by the 'id' field where the document ID is stored
}

foundDataBytes, err := db.GetByFilter(ctx, filters, collectionName)
if err != nil {
    log.Printf("Error getting registro by ID '%s': %v", registroID, err)
    return
}

var foundRegistros []map[string]interface{}
if err := json.Unmarshal(foundDataBytes, &foundRegistros); err != nil {
    log.Printf("Error unmarshalling found registros: %v", err)
    return
}

if len(foundRegistros) == 0 {
    log.Printf("No registro found with ID: %s", registroID)
    return
}
if len(foundRegistros) > 1 {
    log.Printf("Warning: Multiple registros found with ID: %s. Using the first one.", registroID)
}

registro := foundRegistros[0]
log.Printf("Registro found: %v", registro)
```
**Important**: Querying by `firestore.DocumentID` directly is more efficient in Firestore. The current `GetByFilter` queries a field named `"id"`. If this field is not indexed, performance might degrade on large collections. For optimal performance, ensure an index on the `"id"` field or use the direct client for `docRef.Get()`.

#### 3.2.2. Get All Registros in a Collection

Use the `Get` method.

```go
// Assuming 'db' is your initialized FirebaseDBInterface and 'ctx' is your context.
collectionName := "registros"

allDataBytes, err := db.Get(ctx, collectionName)
if err != nil {
    log.Printf("Error getting all registros: %v", err)
    return
}

var allRegistros []map[string]interface{}
if err := json.Unmarshal(allDataBytes, &allRegistros); err != nil {
    log.Printf("Error unmarshalling all registros: %v", err)
    return
}

log.Printf("Found %d registros:", len(allRegistros))
for _, registro := range allRegistros {
    log.Printf("  ID: %s, Description: %s", registro["id"], registro["description"])
}
```

#### 3.2.3. Get Registros by Custom Filter

Use the `GetByFilter` method. This method currently supports exact equality matches.

```go
// Assuming 'db' is your initialized FirebaseDBInterface and 'ctx' is your context.
collectionName := "registros"

filters := map[string]interface{}{
    "isProcessed": false,
    "amount":      5000.00, // This will look for amount EXACTLY 5000.00
}

filteredDataBytes, err := db.GetByFilter(ctx, filters, collectionName)
if err != nil {
    log.Printf("Error filtering registros: %v", err)
    return
}

var filteredRegistros []map[string]interface{}
if err := json.Unmarshal(filteredDataBytes, &filteredRegistros); err != nil {
    log.Printf("Error unmarshalling filtered registros: %v", err)
    return
}

log.Printf("Found %d registros matching filter:", len(filteredRegistros))
for _, registro := range filteredRegistros {
    log.Printf("  ID: %s, Description: %s, Processed: %v", registro["id"], registro["description"], registro["isProcessed"])
}
```

### 3.3. Update a Registro

Use the `Update` method. It updates the document and sets the `updatedAt` server timestamp. It uses `firestore.MergeAll`, meaning only specified fields are changed; other fields remain untouched.

```go
// Assuming 'db' is your initialized FirebaseDBInterface, 'ctx' is your context,
// and 'registroIDToUpdate' is the ID of the document to update.

registroIDToUpdate := "some-firestore-document-id"
collectionName := "registros"

updateData := map[string]interface{}{
    "isProcessed": true,
    "description": "Salário Mensal (Processado)",
    // "updatedAt" will be automatically set by the library
}

err := db.Update(ctx, registroIDToUpdate, updateData, collectionName)
if err != nil {
    log.Printf("Error updating registro '%s': %v", registroIDToUpdate, err)
    return
}

log.Printf("Registro '%s' updated successfully.", registroIDToUpdate)
```

### 3.4. Delete a Registro

Use the `Delete` method.

```go
// Assuming 'db' is your initialized FirebaseDBInterface, 'ctx' is your context,
// and 'registroIDToDelete' is the ID of the document to delete.

registroIDToDelete := "some-firestore-document-id"
collectionName := "registros"

err := db.Delete(ctx, registroIDToDelete, collectionName)
if err != nil {
    log.Printf("Error deleting registro '%s': %v", registroIDToDelete, err)
    return
}

log.Printf("Registro '%s' deleted successfully.", registroIDToDelete)
```

## 4. Advanced Querying: Direct Client Usage

The `FirebaseDBInterface` provides basic CRUD and filtering. For more advanced scenarios like date range queries, ordering, and subcollection operations, you'll need to use the underlying `*firestore.Client` directly.

The `FirebaseDB` struct (which implements `FirebaseDBInterface`) holds an unexported `client` field. To access it for advanced operations, you would ideally have a way to get this client. Since the library code cannot be changed, this section assumes you might have modified the library locally to expose the client, or you are performing these operations in a context where you have access to the client instance.

**Recommendation for Future Library Design:** Consider adding a method to `FirebaseDBInterface` like `GetFirestoreClient() *firestore.Client` to allow users to perform advanced operations not covered by the simplified methods.

For the examples below, assume `fsClient` is your `*firestore.Client` instance. If you have `db *database.FirebaseDB`, then `fsClient` would be `db.client` (if it were exported or you had a getter).

### 4.1. Filtering by Date Ranges

The `GetByFilter` method only supports exact matches. To filter by date ranges (e.g., all records after a certain date), use the client directly:

```go
// Assume 'fsClient' is your *firestore.Client
// Assume 'ctx' is your context

collectionName := "registros"
startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

query := fsClient.Collection(collectionName).Where("date", ">=", startDate)
iter := query.Documents(ctx)
defer iter.Stop()

var results []map[string]interface{}
for {
    doc, err := iter.Next()
    if err == iterator.Done {
        break
    }
    if err != nil {
        log.Fatalf("Failed to iterate: %v", err)
    }
    data := doc.Data()
    data["id"] = doc.Ref.ID
    results = append(results, data)
}

log.Printf("Registros found after %s: %v", startDate, results)
```
You can use operators like `==`, `>`, `>=`, `<`, `<=`. Remember that Firestore may require specific indexes for range queries on multiple fields or if you combine them with ordering.

### 4.2. Ordering (Sorting) Results

To sort results, use the `OrderBy` method on a query. The `GetByFilter` method does not support ordering.

```go
// Assume 'fsClient' is your *firestore.Client
// Assume 'ctx' is your context

collectionName := "registros"

// Order by 'date' in descending order, then by 'amount' in ascending order
query := fsClient.Collection(collectionName).
    OrderBy("date", firestore.Desc).
    OrderBy("amount", firestore.Asc)

iter := query.Documents(ctx)
defer iter.Stop()

var orderedResults []map[string]interface{}
for {
    doc, err := iter.Next()
    if err == iterator.Done {
        break
    }
    if err != nil {
        log.Fatalf("Failed to iterate: %v", err)
    }
    data := doc.Data()
    data["id"] = doc.Ref.ID
    orderedResults = append(orderedResults, data)
}

log.Printf("Registros ordered by date (desc) and amount (asc): %v", orderedResults)
```
Firestore requires composite indexes for queries that order by one field and filter by another.

### 4.3. Working with Subcollections

The methods on `FirebaseDBInterface` are designed for top-level collections. To work with subcollections (collections nested under a document), use the client directly.

For example, if each "registro" document can have a "comments" subcollection:

```go
// Assume 'fsClient' is your *firestore.Client
// Assume 'ctx' is your context
// Assume 'parentRegistroID' is the ID of the registro document containing the subcollection

parentRegistroID := "some-parent-registro-id"
collectionName := "registros"
subCollectionName := "comments"

// Adding a document to a subcollection
commentData := map[string]interface{}{
    "text":    "This is a comment.",
    "user":    "user123",
    "postedAt": firestore.ServerTimestamp, // Use server timestamp
}
_, _, err := fsClient.Collection(collectionName).Doc(parentRegistroID).
    Collection(subCollectionName).Add(ctx, commentData)
if err != nil {
    log.Fatalf("Failed to add comment to subcollection: %v", err)
}
log.Println("Comment added successfully!")

// Querying a subcollection
iter := fsClient.Collection(collectionName).Doc(parentRegistroID).
    Collection(subCollectionName).
    OrderBy("postedAt", firestore.Desc).
    Documents(ctx)
defer iter.Stop()

var comments []map[string]interface{}
for {
    doc, err := iter.Next()
    if err == iterator.Done {
        break
    }
    if err != nil {
        log.Fatalf("Failed to iterate subcollection: %v", err)
    }
    data := doc.Data()
    data["id"] = doc.Ref.ID
    comments = append(comments, data)
}
log.Printf("Comments for registro '%s': %v", parentRegistroID, comments)
```

## 5. Recommendations and Best Practices

*   **Error Handling:** Always check for errors returned by the library methods and Firestore client operations.
*   **Context Management:** Pass appropriate `context.Context` to all methods for request lifecycle management (e.g., timeouts, cancellation).
*   **Data Modeling:**
    *   Structure your Firestore data effectively. Consider if embedding data or using subcollections is more appropriate for your use case.
    *   Be mindful of Firestore's limits (e.g., document size, write rates).
*   **Indexes:** For efficient querying, especially with filters, ranges, and ordering, configure appropriate Firestore indexes. The Firebase console usually suggests required indexes if a query fails due to missing ones.
*   **Security Rules:** Implement robust Firestore security rules to protect your data. The library itself does not handle authorization beyond what's configured for the client. The `validateWithData` and `validateWithoutData` functions in `dbFirebase.go` check for an "Authorization" value in the context, implying an external mechanism for this.
*   **Client Lifecycle:** Manage the lifecycle of the `FirebaseDB` instance (or the underlying `firestore.Client`). Initialize it once and reuse it. Ensure it's closed properly when your application shuts down to free up resources (see note in section 2.2 regarding `Close()`).
*   **Data Serialization:** The library examples use `map[string]interface{}`. For more type safety, consider using structs and tools like `firestore.DocumentData` or struct tags for marshaling/unmarshaling data to/from Firestore.
*   **Large Datasets:** For operations on very large datasets, be mindful of read/write costs and potential performance bottlenecks. Use pagination for reads (`Limit`, `StartAt`, `StartAfter` with the direct client) where appropriate. The current library methods do not support pagination.

## 6. Conclusion

The `dbFirebase.go` library provides a foundational layer for interacting with Firebase Firestore. While its direct methods cover basic CRUD and equality-based filtering, accessing the underlying `firestore.Client` is necessary for advanced querying features such as date ranges, sorting, and subcollection manipulation. This documentation provides examples for both scenarios, enabling effective use of Firestore for managing "registros" and other collections.
