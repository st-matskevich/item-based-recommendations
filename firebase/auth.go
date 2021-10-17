//TODO: maybe mocks should be implemented there?
//like in api packages
// - sql reader <- mockable
// - logic
package firebase

import (
	"context"
	"errors"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/lib/pq"
	"github.com/st-matskevich/item-based-recommendations/api/utils"
	"github.com/st-matskevich/item-based-recommendations/db"
)

type FirebaseAuth struct {
	fbAuth *auth.Client
}

var authClient *FirebaseAuth

func mapFirebaseUIDToUserID(UID string) (int64, error) {
	result := utils.GetNextSnowflakeID()
	_, err := db.GetSQLClient().Query("INSERT INTO users (user_id, firebase_uid) VALUES ($1, $2)", result, UID)
	if err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code.Name() == "unique_violation" {
			reader, err := db.GetSQLClient().Query("SELECT user_id FROM users WHERE firebase_uid = $1", UID)
			if err != nil {
				return result, err
			}

			found, err := reader.Next(&result)
			if !found && err == nil {
				err = errors.New(utils.SQL_NO_RESULT)
			}
			return result, err
		} else {
			return result, err
		}
	}

	return result, err
}

func (client *FirebaseAuth) Verify(authorizationHeader string) (int64, error) {
	var result int64
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
