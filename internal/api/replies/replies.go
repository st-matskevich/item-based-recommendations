package replies

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
	"github.com/st-matskevich/item-based-recommendations/internal/firebase"
)

type Reply struct {
	Id             int64     `json:"id"`
	Text           string    `json:"text"`
	DoerName       string    `json:"doerName"`
	DoerID         int64     `json:"-"`
	TaskCustomerID int64     `json:"-"`
	Hidden         bool      `json:"hidden"`
	CreatedAt      time.Time `json:"createdAt"`
}

func getRepliesReader(client *db.SQLClient, taskID int64) (db.ResponseReader, error) {
	return client.Query(`SELECT reply_id, text, users.name, customer_id, hidden, tasks.created_at
						FROM tasks JOIN replies 
						ON tasks.task_id = replies.task_id
						AND tasks.task_id = $1
						JOIN users
						ON replies.creator_id = users.user_id
						ORDER BY created_at`, taskID)
}

func getReplies(reader db.ResponseReader, userID int64) ([]Reply, error) {
	result := []Reply{}
	row := Reply{}
	ok, err := reader.Next(&row.Id, &row.Text, &row.DoerName, &row.TaskCustomerID, &row.Hidden, &row.CreatedAt)
	for ; ok; ok, err = reader.Next(&row.Id, &row.Text, &row.DoerName, &row.TaskCustomerID, &row.Hidden, &row.CreatedAt) {
		if userID == row.TaskCustomerID || userID == row.DoerID {
			result = append(result, row)
		}
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func HandleGetReplies(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	taskID, err := strconv.ParseInt(mux.Vars(r)["task"], 10, 64)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	reader, err := getRepliesReader(db.GetSQLClient(), taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer reader.Close()

	replies, err := getReplies(reader, uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, replies, nil)
}

func createReply(client *db.SQLClient, reply Reply, userID int64, taskID int64) error {
	reader, err := client.Query("INSERT INTO replies(task_id, text, creator_id) VALUES ($1, $2, $3)", taskID, reply.Text, userID)
	reader.Close()
	return err
}

func parseReply(reply Reply) error {
	if reply.Text == "" {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(reply.Text)) > 512 {
		return errors.New(utils.INVALID_INPUT)
	}

	return nil
}

func HandleCreateReply(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	input := Reply{}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = parseReply(input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	taskID, err := strconv.ParseInt(mux.Vars(r)["task"], 10, 64)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = createReply(db.GetSQLClient(), input, uid, taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}
