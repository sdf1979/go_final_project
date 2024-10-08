package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	constVal "github.com/sdf1979/go_final_project/const"
	"github.com/sdf1979/go_final_project/db"
)

func NextDateHandler(writer http.ResponseWriter, request *http.Request) {
	now := request.FormValue("now")
	date := request.FormValue("date")
	repeat := request.FormValue("repeat")

	nowTime, err := time.Parse(constVal.FormatDate, now)
	if err != nil {
		http.Error(writer, "Invalid 'now' date format", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.Write([]byte(nextDate))
}

func TaskHandler(store *db.Store) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			taskPost(writer, request, store)
		case http.MethodGet:
			taskGet(writer, request, store)
		case http.MethodPut:
			taskPut(writer, request, store)
		case http.MethodDelete:
			taskDelete(writer, request, store)
		default:
			http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func TasksHandler(store *db.Store) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		tasksGet(writer, request, store)
	}
}

func TaskDoneHandler(store *db.Store) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		searchId := request.URL.Query().Get("id")
		if searchId == "" {
			responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("task done - empty task id:"))
			return
		}

		id, err := strconv.ParseInt(searchId, 10, 64)
		if err != nil {
			fmt.Println("task done - invalid task id: " + err.Error())
			responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
			return
		}

		task, err := store.GetTasksById(id)
		if err != nil {
			if err == sql.ErrConnDone {
				responseWithJson(writer, http.StatusInternalServerError, formatErrorForFrontend(err.Error()))
			} else {
				responseWithJson(writer, http.StatusNotFound, formatErrorForFrontend("task done - task not found: "+err.Error()))
			}
			return
		}

		if task.Repeat == "" {
			if err := store.DeleteTask(task.Id); err != nil {
				if err == sql.ErrConnDone {
					responseWithJson(writer, http.StatusInternalServerError, formatErrorForFrontend(err.Error()))
				} else {
					responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("task done - delete error: "+err.Error()))
				}
				return
			}
		} else {
			nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				responseWithJson(writer, http.StatusNotFound, formatErrorForFrontend("task done - next date calculation error: "+err.Error()))
				return
			}
			task.Date = nextDate
			if err := store.UpdateTask(&task); err != nil {
				if err == sql.ErrConnDone {
					responseWithJson(writer, http.StatusInternalServerError, formatErrorForFrontend(err.Error()))
				} else {
					responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("task done - update error: "+err.Error()))
				}
				return
			}
		}
		responseWithJson(writer, http.StatusOK, make(map[string]string))
	}
}

func taskPost(writer http.ResponseWriter, request *http.Request, store *db.Store) {
	var task db.Task
	if err := json.NewDecoder(request.Body).Decode(&task); err != nil {
		responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend(err.Error()))
		return
	}

	if err := validateTask(&task); err != nil {
		fmt.Println("failed to validate task: " + err.Error())
		responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
		return
	}

	if err := store.CreateTask(&task); err != nil {
		if err == sql.ErrConnDone {
			responseWithJson(writer, http.StatusInternalServerError, formatErrorForFrontend(err.Error()))
		} else {
			fmt.Println("failed to create task: " + err.Error())
			responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
		}
		return
	}

	responseWithJson(writer, http.StatusCreated, formatTaskForFrontend(&task))
}

func tasksGet(writer http.ResponseWriter, request *http.Request, store *db.Store) {
	var tasks []db.Task
	var errTasks error = nil

	search := request.URL.Query().Get("search")
	if search != "" {
		searchDate, err := time.Parse(constVal.FormatDateSearch, search)
		if err == nil {
			tasks, errTasks = store.GetTasksByDate(searchDate.Format(constVal.FormatDate), constVal.LimitTasks)
		} else {
			tasks, errTasks = store.SearchTasks(search, constVal.LimitTasks)
		}
	} else {
		tasks, errTasks = store.GetTasks(constVal.LimitTasks)
	}

	if errTasks != nil {
		if errTasks == sql.ErrConnDone {
			responseWithJson(writer, http.StatusInternalServerError, formatErrorForFrontend(errTasks.Error()))
		} else {
			responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend(errTasks.Error()))
		}
		return
	}

	responseWithJson(writer, http.StatusOK, formatTasksForFrontend(tasks))
}

func taskGet(writer http.ResponseWriter, request *http.Request, store *db.Store) {
	searchId := request.URL.Query().Get("id")
	if searchId != "" {
		id, err := strconv.ParseInt(searchId, 10, 64)
		if err != nil {
			fmt.Println("task get - invalid task id: " + err.Error())
			responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
			return
		}

		task, err := store.GetTasksById(id)
		if err != nil {
			if err == sql.ErrConnDone {
				responseWithJson(writer, http.StatusInternalServerError, formatErrorForFrontend(err.Error()))
			} else {
				fmt.Println("task get - task not found: " + err.Error())
				responseWithJson(writer, http.StatusNotFound, formatErrorForFrontend("Bad request"))
			}
			return
		}
		responseWithJson(writer, http.StatusOK, formatTaskForFrontend(&task))
		return
	}

	responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("task get - no param:"))
}

func taskPut(writer http.ResponseWriter, request *http.Request, store *db.Store) {
	var taskFrontend db.TaskFrontend
	if err := json.NewDecoder(request.Body).Decode(&taskFrontend); err != nil {
		fmt.Println("task put - decoding error: " + err.Error())
		responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
		return
	}

	task := taskFrontend.ToTask()
	if err := validateTask(&task); err != nil {
		fmt.Println("task put - failed to validate task: " + err.Error())
		responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
		return
	}

	if err := store.UpdateTask(&task); err != nil {
		if err == sql.ErrConnDone {
			responseWithJson(writer, http.StatusInternalServerError, formatErrorForFrontend(err.Error()))
		} else {
			fmt.Println("task put - update error: " + err.Error())
			responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
		}
		return
	}

	responseWithJson(writer, http.StatusOK, make(map[string]string))
}

func taskDelete(writer http.ResponseWriter, request *http.Request, store *db.Store) {
	searchId := request.URL.Query().Get("id")
	if searchId != "" {
		id, err := strconv.ParseInt(searchId, 10, 64)
		if err != nil {
			fmt.Println("task delete - invalid task id: " + err.Error())
			responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
			return
		}

		if err := store.DeleteTask(id); err != nil {
			if err == sql.ErrConnDone {
				responseWithJson(writer, http.StatusInternalServerError, formatErrorForFrontend(err.Error()))
			} else {
				fmt.Println("task delete - delete error: " + err.Error())
				responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
			}
			return
		}

		responseWithJson(writer, http.StatusOK, make(map[string]string))
		return
	}
	fmt.Println("task delete - empty task id:")
	responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend("Bad request"))
}

func validateTask(task *db.Task) error {
	if task.Title == "" {
		return errors.New("validate task: empty title")
	}

	t := time.Now()
	timeNow := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

	if task.Date == "" {
		task.Date = timeNow.Format(constVal.FormatDate)
	}

	taskDateTime, err := time.Parse(constVal.FormatDate, task.Date)
	if err != nil {
		return errors.New("validate task: " + err.Error())
	}

	if taskDateTime.Before(timeNow) {
		if task.Repeat == "" {
			task.Date = timeNow.Format(constVal.FormatDate)
		} else {
			nextDate, err := NextDate(timeNow, task.Date, task.Repeat)
			if err != nil {
				return errors.New("validate task: " + err.Error())
			}
			task.Date = nextDate
		}
	} else if task.Repeat != "" {
		_, err := NextDate(timeNow, task.Date, task.Repeat)
		if err != nil {
			return errors.New("validate task: " + err.Error())
		}
	}

	return nil
}

func responseWithJson(writer http.ResponseWriter, httpCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writer.WriteHeader(httpCode)
	writer.Write(response)
}

func formatTaskForFrontend(task *db.Task) map[string]interface{} {
	return map[string]interface{}{
		"id":      strconv.FormatInt(task.Id, 10),
		"title":   task.Title,
		"comment": task.Comment,
		"date":    task.Date,
		"repeat":  task.Repeat,
	}
}

func formatErrorForFrontend(errorStr string) map[string]interface{} {
	return map[string]interface{}{
		"error": errorStr,
	}
}

func formatTasksForFrontend(tasks []db.Task) map[string]interface{} {
	formattedTasks := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		formattedTasks[i] = formatTaskForFrontend(&task)
	}
	return map[string]interface{}{
		"tasks": formattedTasks,
	}
}
