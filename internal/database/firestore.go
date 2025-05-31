package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/grpc/codes" // For GetProfile not found error
	"google.golang.org/grpc/status" // For GetProfile not found error

	"example.com/profile-service/internal/domain"
)

const ProfileCollection = "profiles"

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
	_, err := client.Collection(ProfileCollection).Doc(userID).Set(ctx, profile)
	if err != nil {
		log.Printf("Error creating profile for userID %s: %v", userID, err)
		return err
	}
	return nil
}

// GetProfile retrieves a profile document from Firestore by userID.
func GetProfile(ctx context.Context, client *firestore.Client, userID string) (*domain.Profile, error) {
	docSnap, err := client.Collection(ProfileCollection).Doc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			log.Printf("Profile not found for userID %s", userID)
			return nil, status.Errorf(codes.NotFound, "profile not found for userID %s", userID)
		}
		log.Printf("Error getting profile for userID %s: %v", userID, err)
		return nil, err
	}

	var profile domain.Profile
	if err := docSnap.DataTo(&profile); err != nil {
		log.Printf("Error unmarshalling profile data for userID %s: %v", userID, err)
		return nil, err
	}
	return &profile, nil
}

// UpdateProfile updates specific fields of a profile document in Firestore.
// profileUpdates is a map of field names to new values.
func UpdateProfile(ctx context.Context, client *firestore.Client, userID string, profileUpdates map[string]interface{}) error {
	if len(profileUpdates) == 0 {
		return status.Errorf(codes.InvalidArgument, "no updates provided")
	}

	var firestoreUpdates []firestore.Update
	for key, value := range profileUpdates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: key, Value: value})
	}

	_, err := client.Collection(ProfileCollection).Doc(userID).Update(ctx, firestoreUpdates)
	if err != nil {
		if status.Code(err) == codes.NotFound { // Check if the document to update exists
			return status.Errorf(codes.NotFound, "profile not found for userID %s to update", userID)
		}
		log.Printf("Error updating profile for userID %s: %v", userID, err)
		return err
	}
	return nil
}

// DeleteProfile deletes a profile document from Firestore by userID.
func DeleteProfile(ctx context.Context, client *firestore.Client, userID string) error {
	_, err := client.Collection(ProfileCollection).Doc(userID).Delete(ctx)
	if err != nil {
		// Firestore Delete does not return an error if the document doesn't exist,
		// but it's good practice to log or handle other potential errors.
		log.Printf("Error deleting profile for userID %s: %v", userID, err)
		return err
	}
	return nil
}
