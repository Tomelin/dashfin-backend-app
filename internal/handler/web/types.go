package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Tomelin/dashfin-backend-app/pkg/authenticatior"
)

const AuthTokenKey = "Authorization"

func GetRequiredHeaders(authClient authenticatior.Authenticator, r *http.Request) (userID string, authorization string, err error) {
	userID = r.Header.Get("X-Userid")
	authorization = r.Header.Get("X-Authorization")

	if userID == "" || authorization == "" {
		return "", "", fmt.Errorf("X-Userid and X-Authorization headers are required")
	}

	if len(authorization) > 7 && authorization[:7] == "Bearer " {
		authorization = authorization[7:]
	}

	tokenUserID, err := validAuth(authClient, userID, authorization)
	if err != nil {
		return "", "", err
	}

	// Verify that the user ID from the token matches the X-Userid header
	if tokenUserID != userID {
		return "", "", fmt.Errorf("user ID mismatch: token user ID (%s) does not match X-Userid header (%s)", tokenUserID, userID)
	}

	return tokenUserID, authorization, nil
}

func validAuth(authClient authenticatior.Authenticator, userID, authToken string) (string, error) {
	claims, err := authClient.ValidateToken(context.TODO(), userID, authToken)
	if err != nil {
		return "", err
	}

	// Extract user ID from token claims
	tokenUserID := ""
	if uid, ok := claims["user_id"].(string); ok {
		tokenUserID = uid
	} else if sub, ok := claims["sub"].(string); ok {
		tokenUserID = sub
	}

	if tokenUserID == "" {
		return "", fmt.Errorf("user ID not found in token claims")
	}

	return tokenUserID, nil
}
