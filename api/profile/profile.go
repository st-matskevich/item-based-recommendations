package profile

import (
	"errors"

	"github.com/st-matskevich/item-based-recommendations/db"
)

type UserProfile struct {
	Name       string `json:"name"`
	IsCustomer bool   `json:"customer"`
}

func GetUserProfileReader(client *db.SQLClient, id int) (db.ResponseReader, error) {
	return client.Query("SELECT name, is_customer FROM users WHERE user_id = $1", id)
}

func GetUserProfile(reader db.ResponseReader) (UserProfile, error) {
	result := UserProfile{}
	found, err := reader.Next(&result.Name, &result.IsCustomer)
	if !found && err == nil {
		err = errors.New("db has not returned any rows")
	}
	return result, err
}

func SetUserProfile(client *db.SQLClient, id int, profile UserProfile) error {
	_, err := client.Query("UPDATE users SET name = $2, is_customer = $3 WHERE user_id = $1; ", id, profile.Name, profile.IsCustomer)
	return err
}
