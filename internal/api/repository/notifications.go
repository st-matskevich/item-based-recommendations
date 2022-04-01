package repository

import (
	"time"

	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

type Notification struct {
	ID        utils.UID   `json:"id"`
	TriggerID utils.UID   `json:"-"`
	Type      int         `json:"type"`
	Content   interface{} `json:"content"`
	CreatedAt time.Time   `json:"createdAt"`
}

const (
	TASK_CLOSE_NOTIFICATION = 0
	NEW_REPLY_NOTIFICATION  = 10000
)

type NotificationsRepository interface {
	GetNotifications(userID utils.UID) ([]Notification, error)
	CreateNotification(userID utils.UID, notificationType int, triggerID utils.UID) error
}

type NotificationsSQLRepository struct {
	SQLClient *db.SQLClient
}

func (repo *NotificationsSQLRepository) GetNotifications(userID utils.UID) ([]Notification, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT notification_id, trigger_id, type, created_at 
		FROM notifications WHERE user_id = $1
		ORDER BY notification_id DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	notifications := []Notification{}
	row := Notification{}
	for {
		ok, err := reader.NextRow(&row.ID, &row.TriggerID, &row.Type, &row.CreatedAt)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		notifications = append(notifications, row)
	}

	return notifications, nil
}

func (repo *NotificationsSQLRepository) CreateNotification(userID utils.UID, notificationType int, triggerID utils.UID) error {
	return repo.SQLClient.Exec("INSERT INTO notifications(user_id, type, trigger_id) VALUES ($1, $2, $3)", userID, notificationType, triggerID)
}
