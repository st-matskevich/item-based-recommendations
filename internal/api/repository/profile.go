package repository

import (
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

type UserData struct {
	ID         utils.UID `json:"id"`
	Name       string    `json:"name"`
	IsCustomer *bool     `json:"customer,omitempty"`
}

type ProfileRepository interface {
	GetProfile(userID utils.UID) (*UserData, error)
	SetProfile(userID utils.UID, profile UserData) error
}

type ProfileSQLRepository struct {
	SQLClient *db.SQLClient
}

func (repo *ProfileSQLRepository) GetProfile(userID utils.UID) (*UserData, error) {
	reader, err := repo.SQLClient.Query("SELECT name, is_customer FROM users WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	row := UserData{}
	err = reader.GetRow(&row.Name, &row.IsCustomer)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (repo *ProfileSQLRepository) SetProfile(userID utils.UID, profile UserData) error {
	return repo.SQLClient.Exec("UPDATE users SET name = $2, is_customer = $3 WHERE user_id = $1", userID, profile.Name, profile.IsCustomer)
}
