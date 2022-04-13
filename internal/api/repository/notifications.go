package repository

import (
	"time"

	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

type Notification struct {
	ID        utils.UID        `json:"id"`
	TriggerID utils.UID        `json:"-"`
	Type      int              `json:"type"`
	Content   utils.JSONObject `json:"content"`
	CreatedAt time.Time        `json:"createdAt"`
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
		`SELECT notifications.notification_id, notifications.type, notifications.created_at,
		(CASE WHEN notifications.type=0 THEN JSON_BUILD_OBJECT('id', ENCODE(task_trigger.task_id::text::bytea, 'base64'), 'name', task_trigger.name, 'customer', JSON_BUILD_OBJECT('name', users.name))
		WHEN notifications.type=10000 THEN JSON_BUILD_OBJECT('task', JSON_BUILD_OBJECT('id', ENCODE(tasks.task_id::text::bytea, 'base64'), 'name', tasks.name), 'reply', JSON_BUILD_OBJECT('creator', JSON_BUILD_OBJECT('name', users.name), 'text', reply_trigger.text))
		ELSE (NULL) END) AS content
		FROM notifications
		LEFT JOIN replies AS reply_trigger
		ON notifications.trigger_id = reply_trigger.reply_id
		LEFT JOIN tasks AS task_trigger
		ON notifications.trigger_id = task_trigger.task_id
		LEFT JOIN tasks
		ON reply_trigger.task_id = tasks.task_id
		LEFT JOIN users
		ON task_trigger.customer_id = users.user_id
		OR reply_trigger.creator_id = users.user_id
		WHERE notifications.user_id = $1
		ORDER BY notifications.notification_id DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	notifications := []Notification{}
	row := Notification{}
	for {
		ok, err := reader.NextRow(&row.ID, &row.Type, &row.CreatedAt, &row.Content)
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
