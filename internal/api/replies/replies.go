package replies

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

const (
	ALL  = "ALL"
	USER = "USER"
	DOER = "DOER"
)

type Reply struct {
	Id        utils.UID      `json:"id"`
	Text      string         `json:"text"`
	Creator   utils.UserData `json:"creator"`
	CreatedAt time.Time      `json:"createdAt"`
}

type RepliesReaders struct {
	UserReplyReader, DoerReplyReader, AllRepliesReader db.ResponseReader
}

type TaskReplies struct {
	User *Reply  `json:"user"`
	Doer *Reply  `json:"doer"`
	All  []Reply `json:"all"`
}

//TODO: should USER and DOER be excluded from ALL?
//proably not, because if DOER is not null, ALL will not be rendered
func getRepliesReader(client *db.SQLClient, userID utils.UID, taskID utils.UID, scope string) (db.ResponseReader, error) {
	switch scope {
	case USER:
		return client.Query(`SELECT replies.reply_id, replies.text, users.user_id, users.name, replies.created_at
							FROM tasks JOIN replies 
							ON tasks.task_id = replies.task_id
							AND tasks.task_id = $1
							AND replies.creator_id = $2
							JOIN users
							ON replies.creator_id = users.user_id
							ORDER BY replies.reply_id`, taskID, userID)
	case DOER:
		return client.Query(`SELECT replies.reply_id, replies.text, users.user_id, users.name, tasks.customer_id, replies.created_at
							FROM tasks JOIN replies 
							ON tasks.task_id = replies.task_id
							AND tasks.task_id = $1
							AND tasks.doer_id = replies.creator_id
							JOIN users
							ON replies.creator_id = users.user_id
							ORDER BY replies.reply_id`, taskID)
	}
	return client.Query(`SELECT replies.reply_id, replies.text, users.user_id, users.name, tasks.customer_id, replies.created_at
						FROM tasks JOIN replies 
						ON tasks.task_id = replies.task_id
						AND tasks.task_id = $1
						AND replies.hidden = false
						JOIN users
						ON replies.creator_id = users.user_id
						ORDER BY replies.reply_id`, taskID)
}

func getReplies(readers RepliesReaders, userID utils.UID) (TaskReplies, error) {
	result := TaskReplies{}
	row := Reply{}
	taskCustomerID := utils.UID(0)

	found, err := readers.UserReplyReader.NextRow(&row.Id, &row.Text, &row.Creator.ID, &row.Creator.Name, &row.CreatedAt)
	if err != nil {
		return result, err
	}
	if found {
		result.User = &row
	}

	found, err = readers.DoerReplyReader.NextRow(&row.Id, &row.Text, &row.Creator.ID, &row.Creator.Name, &taskCustomerID, &row.CreatedAt)
	if err != nil {
		return result, err
	}
	if found && userID == taskCustomerID {
		result.Doer = &row
	}

	replies := []Reply{}
	for {
		ok, err := readers.AllRepliesReader.NextRow(&row.Id, &row.Text, &row.Creator.ID, &row.Creator.Name, &taskCustomerID, &row.CreatedAt)
		if err != nil {
			return result, err
		}
		if !ok {
			break
		}

		if userID == taskCustomerID {
			replies = append(replies, row)
		}
	}

	result.All = replies

	return result, nil
}

func HandleGetReplies(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	taskID := utils.UID(0)
	err := taskID.FromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	userReplyReader, err := getRepliesReader(db.GetSQLClient(), uid, taskID, USER)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer userReplyReader.Close()

	doerReplyReader, err := getRepliesReader(db.GetSQLClient(), uid, taskID, DOER)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer doerReplyReader.Close()

	allRepliesReader, err := getRepliesReader(db.GetSQLClient(), uid, taskID, ALL)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer allRepliesReader.Close()

	readers := RepliesReaders{
		UserReplyReader:  userReplyReader,
		DoerReplyReader:  doerReplyReader,
		AllRepliesReader: allRepliesReader,
	}

	replies, err := getReplies(readers, uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, replies, nil)
}

//TODO: check if user is customer of the task
func createReply(client *db.SQLClient, reply Reply, userID utils.UID, taskID utils.UID) error {
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

func HandleCreateReply(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	input := Reply{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = parseReply(input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	var taskID utils.UID
	err = taskID.FromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = createReply(db.GetSQLClient(), input, uid, taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}
