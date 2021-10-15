//TODO: maybe mocks should be implemented there?
//like in api packages
// - sql reader <- mockable
// - logic
package firebase

import (
	"context"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/st-matskevich/item-based-recommendations/db"
)

type FirebaseAuth struct {
	fbAuth *auth.Client
}

var authClient *FirebaseAuth

func mapFirebaseUIDToUserID(UID string) (int, error) {
	result := -1
	err := db.GetSQLClient().QueryRow("SELECT user_id FROM users WHERE firebase_uid = $1", UID).Scan(&result)
	return result, err
}

func (client *FirebaseAuth) Verify(authorizationHeader string) (int, error) {
	result := -1
	tokenString := strings.TrimSpace(strings.Replace(authorizationHeader, "Bearer", "", 1))
	token, err := client.fbAuth.VerifyIDToken(context.Background(), tokenString)

	if err != nil {
		return result, err
	}

	return mapFirebaseUIDToUserID(token.UID)
}

func GetFirebaseAuth() *FirebaseAuth {
	return authClient
}
