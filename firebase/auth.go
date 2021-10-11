package firebase

import (
	"context"
	"strings"

	"firebase.google.com/go/auth"
)

type AuthProvider interface {
	Verify(authorizationHeader string) (string, error)
}

type FirebaseAuth struct {
	fbAuth *auth.Client
}

var authClient *FirebaseAuth

func (client *FirebaseAuth) Verify(authorizationHeader string) (string, error) {
	tokenString := strings.TrimSpace(strings.Replace(authorizationHeader, "Bearer", "", 1))
	token, err := client.fbAuth.VerifyIDToken(context.Background(), tokenString)

	if err != nil {
		return "", err
	} else {
		return token.UID, nil
	}
}

func GetFirebaseAuth() *FirebaseAuth {
	return authClient
}
