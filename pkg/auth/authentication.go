package auth

import (
	"context"
	"errors"
	"log" // Added for logging potential issues during app init

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	// "google.golang.org/api/option" // Would be needed for initialization with service account
)

// ValidateToken validates the given Firebase userID and authToken.
// It returns the token claims if valid, or an error otherwise.
func ValidateToken(ctx context.Context, userID string, authToken string) (map[string]interface{}, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	// Firebase app initialization:
	// Option 1: Initialize default app (if not already done)
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		// If error is "firebase app already exists", try to get the existing default app
		// This specific error string might vary, adjust if necessary based on SDK behavior
		if _, ok := err.(*firebase.DefaultAppExistsError); ok || (err != nil && err.Error() == "firebase: app already exists") {
			app, err = firebase.GetApp("") // Get default app
			if err != nil {
				log.Printf("Error getting existing Firebase app: %v\n", err)
				return nil, errors.New("error getting Firebase app: " + err.Error())
			}
		} else {
			log.Printf("Error initializing new Firebase app: %v\n", err)
			return nil, errors.New("error initializing Firebase app: " + err.Error())
		}
	}
	// Option 2: If you initialize with a specific config:
	// opt := option.WithCredentialsFile("path/to/serviceAccountKey.json")
	// config := &firebase.Config{ProjectID: "my-project-id"}
	// app, err := firebase.NewApp(context.Background(), config, opt) // Or firebase.NewApp(ctx, nil, opt) for default app with options
	// if err != nil {
	//     log.Fatalf("error initializing app: %v
", err)
	// }


	client, err := app.Auth(ctx)
	if err != nil {
		return nil, errors.New("error getting Auth client: " + err.Error())
	}

	verifiedToken, err := client.VerifyIDToken(ctx, authToken)
	if err != nil {
		return nil, errors.New("error verifying ID token: " + err.Error())
	}

	// Extract claims
	claims := verifiedToken.Claims
	if claims == nil {
		return nil, errors.New("token contained no claims")
	}

	// Validate User ID against token UID
	tokenUID, ok := claims["user_id"].(string) // Firebase typically uses "user_id" or "uid" in standard claims.
                                                // If custom claims are used, this key might be different.
                                                // Also, Firebase populates verifiedToken.UID directly.
	if !ok {
        // Fallback to check verifiedToken.UID if "user_id" is not in claims map directly
        if verifiedToken.UID == "" {
             return nil, errors.New("user_id not found in token claims or token UID is empty")
        }
        tokenUID = verifiedToken.UID
	}


	if tokenUID != userID {
		return nil, errors.New("userID mismatch: token UID does not match provided userID")
	}

	// Return all claims from the token
	return claims, nil
}
