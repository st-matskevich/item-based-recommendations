//TODO: maybe mocks should be implemented there?
//like in api packages
// - sql reader <- mockable
// - logic
package firebase

import (
	"context"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

type FirebaseAuth struct {
	fbAuth *auth.Client
}

var authClient *FirebaseAuth

func mapFirebaseUIDToUserID(UID string) (utils.UID, error) {
	var result int64

	reader, err := db.GetSQLClient().Query(
		`WITH new_user AS (
			INSERT INTO users (firebase_uid) 
			VALUES ($1)
			ON CONFLICT (firebase_uid) DO NOTHING
			RETURNING user_id
		) SELECT COALESCE(
			(SELECT user_id FROM new_user),
			(SELECT user_id FROM users WHERE firebase_uid = $1)
		)`, UID,
	)

	if err != nil {
		return utils.UID(result), err
	}

	err = reader.GetRow(&result)
	reader.Close()
	return utils.UID(result), err
}

func (client *FirebaseAuth) Verify(authorizationHeader string) (utils.UID, error) {
	var result int64
	tokenString := strings.TrimSpace(strings.Replace(authorizationHeader, "Bearer", "", 1))
	token, err := client.fbAuth.VerifyIDToken(context.Background(), tokenString)

	if err != nil {
		return utils.UID(result), err
	}

	return mapFirebaseUIDToUserID(token.UID)
}

func GetFirebaseAuth() *FirebaseAuth {
	return authClient
}
