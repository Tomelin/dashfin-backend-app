package auth

import (
	"context"
	"log"
	"strings"

	firebase "firebase.google.com/go/v4"
	// "firebase.google.com/go/v4/auth" // This import will be needed by FirebaseAuthMiddleware, but it's good practice to only have used imports. The linter/tidy might remove it if not used yet.
	// "firebase.google.com/go/v4/auth" // Already implicitly used by firebaseApp.Auth()
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)


// Define a custom context key for userID
type contextKey string
const userIDKey contextKey = "userID"

// NewContextWithUserID creates a new context with the userID value.
func NewContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext retrieves the userID from context.
// Returns empty string if not found.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

// InitFirebase initializes the Firebase Admin SDK.
// It expects the GOOGLE_APPLICATION_CREDENTIALS environment variable to be set.
func InitFirebase(ctx context.Context) (*firebase.App, error) {
	conf := &firebase.Config{
		// ProjectID will be inferred from the credentials if not set.
	}
	// If GOOGLE_APPLICATION_CREDENTIALS is set, it's used automatically.
	// Otherwise, specify credentials using option.WithCredentialsFile("path/to/serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		// log.Fatalf causes exit. For a library function, better to return the error.
		log.Printf("error initializing app: %v\n", err)
		return nil, err
	}
	return app, nil
}

// FirebaseAuthMiddleware creates a Gin middleware for Firebase authentication.
func FirebaseAuthMiddleware(firebaseApp *firebase.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.GetHeader("X-AUTHORIZATION")
		appHeader := c.GetHeader("X-APP")
		userIDHeader := c.GetHeader("X-USERID")

		if authToken == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "X-AUTHORIZATION header required"})
			return
		}

		if appHeader == "" {
			// As per requirements, X-APP just needs to be present.
			c.AbortWithStatusJSON(401, gin.H{"error": "X-APP header required"})
			return
		}

		if userIDHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "X-USERID header required"})
			return
		}

		// Remove "Bearer " prefix if present
		if strings.HasPrefix(authToken, "Bearer ") {
			authToken = strings.TrimPrefix(authToken, "Bearer ")
		}

		client, err := firebaseApp.Auth(context.Background())
		if err != nil {
			log.Printf("Error getting Auth client: %v\n", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal server error initializing auth"})
			return
		}

		token, err := client.VerifyIDToken(context.Background(), authToken)
		if err != nil {
			log.Printf("Error verifying ID token: %v\n", err)
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid X-AUTHORIZATION token"})
			return
		}

		if token.UID != userIDHeader {
			log.Printf("X-USERID (%s) does not match token UID (%s)\n", userIDHeader, token.UID)
			c.AbortWithStatusJSON(401, gin.H{"error": "X-USERID does not match authenticated user"})
			return
		}

		// Store the userID in context for downstream handlers
		c.Set("userID", token.UID)
		c.Next()
	}
}

// UnaryAuthInterceptor provides gRPC unary server interceptor for Firebase authentication.
func UnaryAuthInterceptor(firebaseApp *firebase.App) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "Retrieving metadata failed")
		}

		authHeaders := md.Get("x-authorization")
		appHeaders := md.Get("x-app")
		userIDHeaders := md.Get("x-userid")

		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "x-authorization header required")
		}
		authToken := authHeaders[0]

		if len(appHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "x-app header required")
		}
		// appHeader := appHeaders[0] // Value not used yet, just presence check

		if len(userIDHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "x-userid header required")
		}
		userIDHeader := userIDHeaders[0]

		if strings.HasPrefix(authToken, "Bearer ") {
			authToken = strings.TrimPrefix(authToken, "Bearer ")
		}

		client, err := firebaseApp.Auth(context.Background()) // Use ctx from interceptor? Usually Background for SDK init calls
		if err != nil {
			log.Printf("Error getting Auth client for gRPC: %v\n", err)
			return nil, status.Errorf(codes.Internal, "Internal server error initializing auth")
		}

		token, err := client.VerifyIDToken(context.Background(), authToken) // Use ctx?
		if err != nil {
			log.Printf("Error verifying ID token for gRPC: %v\n", err)
			return nil, status.Errorf(codes.Unauthenticated, "Invalid x-authorization token")
		}

		if token.UID != userIDHeader {
			log.Printf("gRPC: X-USERID (%s) does not match token UID (%s)\n", userIDHeader, token.UID)
			return nil, status.Errorf(codes.Unauthenticated, "x-userid does not match authenticated user")
		}

		// Add userID to context for downstream handlers
		newCtx := NewContextWithUserID(ctx, token.UID)
		return handler(newCtx, req)
	}
}
