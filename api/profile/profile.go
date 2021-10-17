package profile

import "github.com/st-matskevich/item-based-recommendations/db"

type UserProfile struct {
	name       string
	isCustomer bool
}

func GetUserProfileReader(client *db.SQLClient, id int) (db.ResponseReader, error) {
	return client.Query("SELECT name, is_customer FROM users WHERE user_id = $1", id)
}

func GetUserProfile(reader db.ResponseReader) (UserProfile, error) {
	result := UserProfile{}
	_, err := reader.Next(&result.name, &result.isCustomer)
	return result, err
}
