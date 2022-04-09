package similarity

import (
	"math"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
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
						AND likes.user_id = $1 AND likes.active = true`, userID)
}

func getTasksTagsReader(client *db.SQLClient, userID utils.UID) (db.ResponseReader, error) {
	return client.Query(`SELECT task_tag.task_id, task_tag.tag_id 
						FROM likes 
						RIGHT JOIN task_tag 
						ON likes.task_id = task_tag.task_id 
						AND likes.user_id = $1 AND likes.active = true 
						WHERE likes.user_id IS NULL`, userID)
}

func normalizeVector(vector map[utils.UID]float32) {
	magnitude := float32(0)
	for _, val := range vector {
		magnitude += val * val
	}
	magnitude = float32(math.Sqrt(float64(magnitude)))

	for id, val := range vector {
		vector[id] = val / magnitude
	}
}

func readUserProfile(reader db.ResponseReader) (map[utils.UID]float32, error) {
	result := map[utils.UID]float32{}
	uniqueTasks := map[utils.UID]struct{}{}

	row := TaskTagLink{}
	for {
		ok, err := reader.NextRow(&row.TaskID, &row.TagID)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		result[row.TagID] += 1
		uniqueTasks[row.TaskID] = struct{}{}
	}

	for tagID := range result {
		result[tagID] /= float32(len(uniqueTasks))
	}

	normalizeVector(result)

	return result, nil
}

func readTasksTags(reader db.ResponseReader) (map[utils.UID]map[utils.UID]float32, error) {
	result := map[utils.UID]map[utils.UID]float32{}
	uniqueTasks := map[utils.UID]struct{}{}
	uniqueTags := map[utils.UID]float32{}

	row := TaskTagLink{}
	for {
		ok, err := reader.NextRow(&row.TaskID, &row.TagID)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		if _, contains := result[row.TaskID]; !contains {
			result[row.TaskID] = map[utils.UID]float32{}
		}

		result[row.TaskID][row.TagID] = 0
		uniqueTags[row.TagID] += 1
		uniqueTasks[row.TaskID] = struct{}{}
	}

	for taskID, tagsMap := range result {
		for tagID := range tagsMap {
			result[taskID][tagID] = uniqueTags[tagID] / float32(len(uniqueTasks))
		}

		normalizeVector(result[taskID])
	}

	return result, nil
}

func getSimilarTasks(readers ProfilesReaders, threshold float32) ([]TaskSimilarity, error) {
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

		if similarity >= threshold {
			result = append(result, TaskSimilarity{taskID, similarity})
		}
	}
	return result, nil
}

const SIMILARITY_THRESHOLD = 0.60

func HandleGetRecommendations(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

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

	topList, err := getSimilarTasks(readers, SIMILARITY_THRESHOLD)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, topList, nil)
}
