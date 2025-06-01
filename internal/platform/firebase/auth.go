package firebase

import (
	"context"
	"log"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// AuthClient wraps the Firebase Auth client
type AuthClient struct {
	Client *auth.Client
}

// NewAuthClient initializes a new Firebase Auth client.
// It expects FIREBASE_APPLICATION_CREDENTIALS environment variable to be set
// or for the application to be running on Google Cloud infrastructure.
func NewAuthClient(ctx context.Context) (*AuthClient, error) {
	// Initialize Firebase Admin SDK
	// Ensure you have set up GOOGLE_APPLICATION_CREDENTIALS environment variable
	// or are running on a GCP environment with appropriate service account permissions.
	// For local development, you might use:
	// opt := option.WithCredentialsFile("path/to/your/serviceAccountKey.json")
	// app, err := firebase.NewApp(ctx, nil, opt)
	// For now, we'll assume credentials are set up via environment for broader compatibility.
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Printf("Error initializing Firebase app: %v\n", err)
		return nil, err
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Printf("Error getting Firebase Auth client: %v\n", err)
		return nil, err
	}
	return &AuthClient{Client: client}, nil
}

// VerifyToken verifies a Firebase ID token.
// Returns the verified token or an error.
func (ac *AuthClient) VerifyToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := ac.Client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return token, nil
}
