package web

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	fb "github.com/user/supportservice/internal/platform/firebase" // Adjust import path if module name is different
)

// AuthMiddleware creates a Gin middleware for Firebase authentication and header checking.
func AuthMiddleware(authClient *fb.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check for X-APP and X-USERID headers
		appHeader := c.GetHeader("X-APP")
		userIDHeader := c.GetHeader("X-USERID")

		if appHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-APP header is missing"})
			c.Abort()
			return
		}

		if userIDHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-USERID header is missing"})
			c.Abort()
			return
		}

		// 2. Check for X-AUTHORIZATION header (Firebase token)
		authHeader := c.GetHeader("X-AUTHORIZATION")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-AUTHORIZATION header is missing"})
			c.Abort()
			return
		}

		// Extract token (typically "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-AUTHORIZATION header format must be Bearer {token}"})
			c.Abort()
			return
		}
		idToken := parts[1]

		// 3. Verify Firebase token
		token, err := authClient.VerifyToken(context.Background(), idToken) // Use c.Request.Context() in real scenarios
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Firebase token", "details": err.Error()})
			c.Abort()
			return
		}

		// 4. Validate X-USERID against token UID
		if token.UID != userIDHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-USERID does not match authenticated user"})
			c.Abort()
			return
		}

		// You can store validated information in the Gin context if needed
		c.Set("firebase_uid", token.UID)
		c.Set("app_name", appHeader)

		c.Next()
	}
}
