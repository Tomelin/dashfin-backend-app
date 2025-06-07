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

	err = validAuth(authClient, &userID, &authorization)
	if err != nil {
		return "", "", err
	}

	return userID, authorization, nil
}

func validAuth(authClient authenticatior.Authenticator, user, auth *string) error {

	_, err := authClient.ValidateToken(context.TODO(), *user, *auth)
	return err

}
