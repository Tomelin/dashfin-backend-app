package support

import (
	"context"
	"fmt"
	// "log" // If more detailed logging is needed within the service
)

// ServiceInterface defines the interface for support request operations.
// This allows for different implementations if needed (e.g., mock for testing).
type ServiceInterface interface {
	CreateSupportRequest(ctx context.Context, req SupportRequest, userID, appName string) (string, error)
}

// Service implements the business logic for support requests.
type Service struct {
	repository *RepositoryFirestore // Using pointer as per note
}

// NewService creates a new support service instance.
func NewService(repo *RepositoryFirestore) *Service {
	return &Service{
		repository: repo, // Store as pointer
	}
}

// CreateSupportRequest handles the business logic for creating a new support request.
// For now, it primarily validates and then passes data to the repository.
func (s *Service) CreateSupportRequest(ctx context.Context, req SupportRequest, userID, appName string) (string, error) {
	// Additional business logic or validation can go here if needed.
	// For example, checking against a user's quota, enriching data, etc.

	// The primary validation (format, length, enum) is expected to have been done
	// by the handlers (Gin binding, gRPC validation, GraphQL resolver validation).
	// If there were service-specific validation that doesn't fit the handlers, it would be here.

	// log.Printf("Service: Processing support request for UserID: %s, AppName: %s", userID, appName)

	// Persist to Firestore
	requestID, err := s.repository.Save(ctx, req, userID, appName)
	if err != nil {
		// It's good practice to return errors that are not too specific about underlying issues
		// unless the caller needs to know. For now, we'll wrap it.
		return "", fmt.Errorf("failed to save support request: %w", err)
	}

	// log.Printf("Service: Support request saved with ID: %s", requestID)
	return requestID, nil
}
