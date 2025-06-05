package authenticatior

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
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

// Custom errors for better error handling
var (
	ErrEmptyUserID    = errors.New("userID cannot be empty")
	ErrEmptyToken     = errors.New("authToken cannot be empty")
	ErrClientNotInit  = errors.New("auth client not initialized")
	ErrNoClaimsFound  = errors.New("verified token contained no claims")
	ErrUIDNotFound    = errors.New("UID not found in verified token or its claims")
	ErrUserIDMismatch = errors.New("userID mismatch")
)

// FirebaseConfig holds the configuration for Firebase initialization
type FirebaseConfig struct {
	// Path to the service account key file
	ServiceAccountKeyPath string
	// Project ID (optional if using service account key)
	ProjectID string `json:"projectId" yaml:"projectId"`
	// Database URL (optional, for Realtime Database)
	DatabaseURL string `json:"database" yaml:"database"`

	APIKey            string `json:"apiKey" yaml:"apiKey"`
	AuthDomain        string `json:"authDomain" yaml:"authDomain"`
	StorageBucket     string `json:"storageBucket" yaml:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId" yaml:"messagingSenderId"`
	AppID             string `json:"appId" yaml:"appId"`
}

// InitializeAuth initializes the Firebase application and returns an Authenticator.
func InitializeAuth(ctx context.Context, config *FirebaseConfig) (Authenticator, error) {
	if config == nil {
		return nil, errors.New("firebase config cannot be nil")
	}

	var app *firebase.App
	var err error

	// Option 1: Using service account key file
	if config.APIKey != "" {
		opt := option.WithAPIKey(config.APIKey)
		conf := &firebase.Config{
			ProjectID:     config.ProjectID,
			StorageBucket: config.StorageBucket,
		}
		if config.DatabaseURL != "" {
			conf.DatabaseURL = config.DatabaseURL
		}
		app, err = firebase.NewApp(ctx, conf, opt)
	} else {
		// Option 2: Using default credentials (ADC - Application Default Credentials)
		// This works when running on Google Cloud or with GOOGLE_APPLICATION_CREDENTIALS env var
		conf := &firebase.Config{}
		if config.ProjectID != "" {
			conf.ProjectID = config.ProjectID
		}
		if config.DatabaseURL != "" {
			conf.DatabaseURL = config.DatabaseURL
		}
		app, err = firebase.NewApp(ctx, conf)
	}

	if err != nil {
		log.Printf("Error initializing Firebase app: %v", err)
		return nil, fmt.Errorf("error initializing Firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Printf("Error getting Auth client: %v", err)
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}

	return &firebaseAuthenticator{client: client}, nil
}

// ValidateToken validates the Firebase ID token and ensures it belongs to the specified user.
func (fa *firebaseAuthenticator) ValidateToken(ctx context.Context, userID string, authToken string) (map[string]interface{}, error) {
	// Input validation
	if userID == "" {
		return nil, ErrEmptyUserID
	}
	if authToken == "" {
		return nil, ErrEmptyToken
	}
	if fa.client == nil {
		return nil, ErrClientNotInit
	}

	// Verify the ID token
	verifiedToken, err := fa.client.VerifyIDToken(ctx, authToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying ID token: %w", err)
	}

	// Check claims
	claims := verifiedToken.Claims
	if claims == nil {
		return nil, ErrNoClaimsFound
	}

	// Get UID from token
	tokenUID := fa.extractUID(verifiedToken, claims)
	if tokenUID == "" {
		return nil, ErrUIDNotFound
	}

	// Validate UID matches
	if tokenUID != userID {
		return nil, fmt.Errorf("%w: token UID (%s) does not match provided userID (%s)",
			ErrUserIDMismatch, tokenUID, userID)
	}

	return claims, nil
}

// extractUID extracts UID from token, falling back to claims if necessary.
func (fa *firebaseAuthenticator) extractUID(token *auth.Token, claims map[string]interface{}) string {
	if token.UID != "" {
		return token.UID
	}

	// Fallback to claims
	if uidFromClaims, ok := claims["user_id"].(string); ok && uidFromClaims != "" {
		return uidFromClaims
	}

	// Try standard 'sub' claim as well
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub
	}

	return ""
}

// isTokenExpiredError checks if the error indicates an expired token
func (fa *firebaseAuthenticator) isTokenExpiredError(err error) bool {
	if err == nil {
		return false
	}

	// Check common error messages that indicate token expiry
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "expired") ||
		strings.Contains(errStr, "token has expired") ||
		strings.Contains(errStr, "id token has expired")
}

// IsExpired checks if the given Firebase authToken is expired.
func (fa *firebaseAuthenticator) IsExpired(ctx context.Context, authToken string) (bool, error) {
	if authToken == "" {
		return true, ErrEmptyToken
	}
	if fa.client == nil {
		return true, ErrClientNotInit
	}

	_, err := fa.client.VerifyIDToken(ctx, authToken)
	if err != nil {
		// Check if the error indicates token expiry
		if fa.isTokenExpiredError(err) {
			return true, nil // Token is confirmed expired
		}

		// For other verification errors, it's invalid but not necessarily expired
		return false, fmt.Errorf("token verification failed, not necessarily due to expiry: %w", err)
	}

	// If VerifyIDToken is successful, the token is not expired
	return false, nil
}

// IsValid checks if the given Firebase authToken is valid (not expired, correctly signed, etc.).
func (fa *firebaseAuthenticator) IsValid(ctx context.Context, authToken string) (bool, error) {
	if authToken == "" {
		return false, ErrEmptyToken
	}
	if fa.client == nil {
		return false, ErrClientNotInit
	}

	_, err := fa.client.VerifyIDToken(ctx, authToken)
	if err != nil {
		// Any error from VerifyIDToken means the token is not valid
		return false, fmt.Errorf("token validation failed: %w", err)
	}

	// If no error, the token is valid
	return true, nil
}
