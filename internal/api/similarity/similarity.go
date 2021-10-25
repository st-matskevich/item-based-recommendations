package similarity

import (
	"math"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
	"github.com/st-matskevich/item-based-recommendations/internal/firebase"
)

//TODO: similarity field should be private in production
type TaskSimilarity struct {
	Id         utils.UID `json:"id"`
	Similarity float32   `json:"similarity"`
}

type TaskTagLink struct {
	TaskID utils.UID
	TagID  utils.UID
}

type ProfilesReaders struct {
	UserProfileReader, TasksTagsReader db.ResponseReader
}

func getUserProfileReader(client *db.SQLClient, userID utils.UID) (db.ResponseReader, error) {
	return client.Query(`SELECT task_tag.task_id, task_tag.tag_id 
						FROM likes 
						JOIN task_tag 
						ON likes.task_id = task_tag.task_id 
						AND likes.user_id = $1`, userID)
}

func getTasksTagsReader(client *db.SQLClient, userID utils.UID) (db.ResponseReader, error) {
	return client.Query(`SELECT task_tag.task_id, task_tag.tag_id 
						FROM likes 
						RIGHT JOIN task_tag 
						ON likes.task_id = task_tag.task_id 
						AND likes.user_id = $1 
						WHERE likes.user_id IS NULL`, userID)
}

func normalizeVector(vector map[utils.UID]float32) {
	magnitude := float32(0)
	for _, val := range vector {
		magnitude += val * val
	}
	magnitude = float32(math.Sqrt(float64(magnitude)))

	for id, val := range vector {
		val /= magnitude
		vector[id] = val
	}
}

func readUserProfile(reader db.ResponseReader) (map[utils.UID]float32, error) {
	result := map[utils.UID]float32{}

	uniqueTasks := map[utils.UID]struct{}{}
	row := TaskTagLink{}

	ok, err := reader.NextRow(&row.TaskID, &row.TagID)
	for ; ok; ok, err = reader.NextRow(&row.TaskID, &row.TagID) {
		//TODO: can initial value be not 0? if 0 is guranteed, if statement can be removed
		if _, contains := result[row.TagID]; !contains {
			result[row.TagID] = 1
		} else {
			result[row.TagID] += 1
		}
		uniqueTasks[row.TaskID] = struct{}{}
	}

	if err != nil {
		return nil, err
	}

	for tagID := range result {
		result[tagID] /= float32(len(uniqueTasks))
	}

	normalizeVector(result)

	return result, nil
}

func readTasksTags(reader db.ResponseReader) (map[utils.UID]map[utils.UID]float32, error) {
	result := map[utils.UID]map[utils.UID]float32{}

	row := TaskTagLink{}
	ok, err := reader.NextRow(&row.TaskID, &row.TagID)
	for ; ok; ok, err = reader.NextRow(&row.TaskID, &row.TagID) {
		if _, contains := result[row.TaskID]; !contains {
			result[row.TaskID] = map[utils.UID]float32{}
		}

		result[row.TaskID][row.TagID] = 1
	}

	if err != nil {
		return nil, err
	}

	for taskID := range result {
		normalizeVector(result[taskID])
	}

	return result, nil
}

func getSimilarTasks(readers ProfilesReaders, topSize int) ([]TaskSimilarity, error) {
	userProfile, err := readUserProfile(readers.UserProfileReader)
	if err != nil {
		return nil, err
	}

	tasksTags, err := readTasksTags(readers.TasksTagsReader)
	if err != nil {
		return nil, err
	}

	result := []TaskSimilarity{}
	similarity := float32(0)

	for taskID, tagsMap := range tasksTags {
		//TODO: use go coroutines
		similarity = 0
		for tagID, tagWeight := range tagsMap {
			if val, ok := userProfile[tagID]; ok {
				similarity += tagWeight * val
			}
		}

		if len(result) < topSize {
			result = append(result, TaskSimilarity{taskID, similarity})
		} else if similarity > result[len(result)-1].Similarity {
			result[len(result)-1] = TaskSimilarity{taskID, similarity}
		} else {
			continue
		}

		for idx := len(result) - 1; idx > 0 && result[idx].Similarity > result[idx-1].Similarity; idx-- {
			result[idx], result[idx-1] = result[idx-1], result[idx]
		}
	}
	return result, nil
}

const MAX_RECOMMENDED_POSTS = 5

func HandleGetRecommendations(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	profileReader, err := getUserProfileReader(db.GetSQLClient(), uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer profileReader.Close()

	tasksReader, err := getTasksTagsReader(db.GetSQLClient(), uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer tasksReader.Close()

	readers := ProfilesReaders{
		UserProfileReader: profileReader,
		TasksTagsReader:   tasksReader,
	}

	topList, err := getSimilarTasks(readers, MAX_RECOMMENDED_POSTS)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, topList, nil)
}
