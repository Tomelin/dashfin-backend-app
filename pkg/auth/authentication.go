package auth

import (
	"context"
	"errors"
	"log"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	// "google.golang.org/api/option"
)

// Authenticator defines the interface for authentication operations.
type Authenticator interface {
	ValidateToken(ctx context.Context, userID string, authToken string) (map[string]interface{}, error)
	IsExpired(ctx context.Context, authToken string) (bool, error)
	IsValid(ctx context.Context, authToken string) (bool, error)
}

// firebaseAuthenticator implements the Authenticator interface using Firebase.
type firebaseAuthenticator struct {
	client *auth.Client
}

// InitializeAuth initializes the Firebase application and returns an Authenticator.
func InitializeAuth(ctx context.Context) (Authenticator, error) {
	// ... (implementation as before)
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		if _, ok := err.(*firebase.DefaultAppExistsError); ok {
			log.Println("Default Firebase app already exists, attempting to get it.")
			app, err = firebase.GetApp("")
			if err != nil {
				log.Printf("Error getting existing Firebase app: %v\n", err)
				return nil, errors.New("error getting existing Firebase app: " + err.Error())
			}
		} else {
			log.Printf("Error initializing new Firebase app: %v\n", err)
			return nil, errors.New("error initializing new Firebase app: " + err.Error())
		}
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Printf("Error getting Auth client: %v\n", err)
		return nil, errors.New("error getting Auth client: " + err.Error())
	}

	return &firebaseAuthenticator{client: client}, nil
}

// ValidateToken method ... (as before)
func (fa *firebaseAuthenticator) ValidateToken(ctx context.Context, userID string, authToken string) (map[string]interface{}, error) {
	// ... (implementation as before)
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}
	if fa.client == nil {
		return nil, errors.New("auth client not initialized")
	}

	verifiedToken, err := fa.client.VerifyIDToken(ctx, authToken)
	if err != nil {
		return nil, errors.New("error verifying ID token: " + err.Error())
	}

	claims := verifiedToken.Claims
	if claims == nil {
		return nil, errors.New("verified token contained no claims")
	}

	tokenUID := verifiedToken.UID
	if tokenUID == "" {
		uidFromClaims, ok := claims["user_id"].(string)
		if ok && uidFromClaims != "" {
			tokenUID = uidFromClaims
		} else {
			return nil, errors.New("UID not found in verified token or its claims")
		}
	}

	if tokenUID != userID {
		return nil, errors.New("userID mismatch: token UID (" + tokenUID + ") does not match provided userID (" + userID + ")")
	}

	return claims, nil
}

// IsExpired checks if the given Firebase authToken is expired.
func (fa *firebaseAuthenticator) IsExpired(ctx context.Context, authToken string) (bool, error) {
	if fa.client == nil {
		return true, errors.New("auth client not initialized")
	}

	_, err := fa.client.VerifyIDToken(ctx, authToken) // We only need to check the error for expiry
	if err != nil {
		// Check if the error is specifically due to token expiry.
		// The Firebase Admin SDK for Go might wrap errors or have specific error types/codes.
		// For example, auth.ErrIDTokenExpired exists in firebase.google.com/go/v4/auth
		// Using errors.As to check for the specific error type is more robust.
		var idTokenExpiredError *auth.IDTokenExpired
		if errors.As(err, &idTokenExpiredError) {
			return true, nil // Token is confirmed expired
		}
		// For other verification errors, it's invalid but not necessarily because it's "expired".
		return false, errors.New("token verification failed, not necessarily due to expiry: " + err.Error())
	}
	// If VerifyIDToken is successful, the token is not expired.
	return false, nil
}

// IsValid checks if the given Firebase authToken is valid (not expired, correctly signed, etc.).
func (fa *firebaseAuthenticator) IsValid(ctx context.Context, authToken string) (bool, error) {
	if fa.client == nil {
		return false, errors.New("auth client not initialized")
	}

	_, err := fa.client.VerifyIDToken(ctx, authToken)
	if err != nil {
		// Any error from VerifyIDToken means the token is not valid.
		// We return false and the original error so the caller can inspect it.
		return false, err
	}
	// If no error, the token is valid.
	return true, nil
}
