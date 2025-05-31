package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/grpc/codes" // For GetProfile not found error (used by this package)
	"google.golang.org/grpc/status" // For GetProfile not found error (used by this package)

	"example.com/profile-service/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes" // Alias to avoid conflict with grpc/codes
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0" // For semantic conventions
	"go.opentelemetry.io/otel/trace"
)

const ProfileCollection = "profiles"
const firestoreTracerName = "example.com/profile-service/database/firestore"

// InitFirestore initializes and returns a Firestore client.
func InitFirestore(ctx context.Context, app *firebase.App) (*firestore.Client, error) {
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("Error initializing Firestore client: %v", err)
		return nil, err
	}
	return client, nil
}

// CreateProfile creates or overwrites a profile document in Firestore.
// The userID is used as the document ID.
func CreateProfile(ctx context.Context, client *firestore.Client, userID string, profile *domain.Profile) error {
	tracer := otel.Tracer(firestoreTracerName)
	spanCtx, span := tracer.Start(ctx, "Firestore CreateProfile",
		trace.WithAttributes(
			semconv.DBSystemFirestore,
			semconv.DBOperationKey.String("SET"), // Firestore Set is like an UPSERT
			semconv.DBStatementKey.String(ProfileCollection+"/"+userID),
			attribute.String("firestore.collection", ProfileCollection),
			attribute.String("firestore.document_id", userID),
		),
		trace.WithSpanKind(trace.SpanKindClient), // It's a client call to Firestore
	)
	defer span.End()

	_, err := client.Collection(ProfileCollection).Doc(userID).Set(spanCtx, profile) // Use spanCtx
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		log.Printf("Error creating profile for userID %s: %v", userID, err)
		return err
	}
	span.SetStatus(otelCodes.Ok, "Profile created/set successfully")
	return nil
}

// GetProfile retrieves a profile document from Firestore by userID.
func GetProfile(ctx context.Context, client *firestore.Client, userID string) (*domain.Profile, error) {
	tracer := otel.Tracer(firestoreTracerName)
	spanCtx, span := tracer.Start(ctx, "Firestore GetProfile",
		trace.WithAttributes(
			semconv.DBSystemFirestore,
			semconv.DBOperationKey.String("GET"),
			semconv.DBStatementKey.String(ProfileCollection+"/"+userID),
			attribute.String("firestore.collection", ProfileCollection),
			attribute.String("firestore.document_id", userID),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	docSnap, err := client.Collection(ProfileCollection).Doc(userID).Get(spanCtx) // Use spanCtx
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error()) // Generic error status
		if status.Code(err) == codes.NotFound {
			log.Printf("Profile not found for userID %s", userID)
			// Specific gRPC status for not found is returned by the caller (handler)
			return nil, status.Errorf(codes.NotFound, "profile not found for userID %s", userID)
		}
		log.Printf("Error getting profile for userID %s: %v", userID, err)
		return nil, err // Return the original Firestore error or a wrapped one
	}

	var profileData domain.Profile
	if err := docSnap.DataTo(&profileData); err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, "Failed to unmarshal profile data")
		log.Printf("Error unmarshalling profile data for userID %s: %v", userID, err)
		return nil, err
	}
	span.SetStatus(otelCodes.Ok, "Profile retrieved successfully")
	return &profileData, nil
}

// UpdateProfile updates specific fields of a profile document in Firestore.
// profileUpdates is a map of field names to new values.
func UpdateProfile(ctx context.Context, client *firestore.Client, userID string, profileUpdates map[string]interface{}) error {
	tracer := otel.Tracer(firestoreTracerName)
	spanCtx, span := tracer.Start(ctx, "Firestore UpdateProfile",
		trace.WithAttributes(
			semconv.DBSystemFirestore,
			semconv.DBOperationKey.String("UPDATE"),
			semconv.DBStatementKey.String(ProfileCollection+"/"+userID),
			attribute.String("firestore.collection", ProfileCollection),
			attribute.String("firestore.document_id", userID),
			attribute.Int("firestore.update_param_count", len(profileUpdates)),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	if len(profileUpdates) == 0 {
        span.SetStatus(otelCodes.Error, "No updates provided") // Or OK if it's not an error for DB layer
		return status.Errorf(codes.InvalidArgument, "no updates provided")
	}

	var firestoreUpdates []firestore.Update
	for key, value := range profileUpdates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: key, Value: value})
	}

	_, err := client.Collection(ProfileCollection).Doc(userID).Update(spanCtx, firestoreUpdates) // Use spanCtx
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		if status.Code(err) == codes.NotFound { // Check if the document to update exists
			return status.Errorf(codes.NotFound, "profile not found for userID %s to update", userID)
		}
		log.Printf("Error updating profile for userID %s: %v", userID, err)
		return err
	}
	span.SetStatus(otelCodes.Ok, "Profile updated successfully")
	return nil
}

// DeleteProfile deletes a profile document from Firestore by userID.
func DeleteProfile(ctx context.Context, client *firestore.Client, userID string) error {
	tracer := otel.Tracer(firestoreTracerName)
	spanCtx, span := tracer.Start(ctx, "Firestore DeleteProfile",
		trace.WithAttributes(
			semconv.DBSystemFirestore,
			semconv.DBOperationKey.String("DELETE"),
			semconv.DBStatementKey.String(ProfileCollection+"/"+userID),
			attribute.String("firestore.collection", ProfileCollection),
			attribute.String("firestore.document_id", userID),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	_, err := client.Collection(ProfileCollection).Doc(userID).Delete(spanCtx) // Use spanCtx
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelCodes.Error, err.Error())
		// Firestore Delete does not return an error if the document doesn't exist,
		// but other errors can occur.
		log.Printf("Error deleting profile for userID %s: %v", userID, err)
		return err
	}
	span.SetStatus(otelCodes.Ok, "Profile deleted successfully")
	return nil
}
