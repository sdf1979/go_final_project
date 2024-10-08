package db

import (
	"database/sql"
	"errors"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type Task struct {
	Id      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskFrontend struct {
	Id      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func (taskFrontend *TaskFrontend) ToTask() Task {
	var task Task
	id, err := strconv.ParseInt(taskFrontend.Id, 10, 64)
	if err != nil {
		id = 0
	}
	task.Id = id
	task.Date = taskFrontend.Date
	task.Title = taskFrontend.Title
	task.Comment = taskFrontend.Comment
	task.Repeat = taskFrontend.Repeat

	return task
}

func Open(dbFile string) (*Store, error) {
	_, err_install := os.Stat(dbFile)
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err_install != nil {
		err = store.createScheduler()
		if err != nil {
			db.Close()
			return nil, err
		}
	}
	return store, nil
}

func (store *Store) Close() error {
	return store.db.Close()
}

func (store *Store) CreateTask(task *Task) error {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`

	result, err := store.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	task.Id = id

	return nil
}

func (store *Store) GetTasks(limit int) ([]Task, error) {
	rows, err := store.db.Query(getTasksQuery(), limit)
	if err != nil {
		return nil, err
	}
	return rowsToTasks(rows)
}

func (store *Store) GetTasksByDate(date string, limit int) ([]Task, error) {
	rows, err := store.db.Query(getTasksByDateQuery(), date, limit)
	if err != nil {
		return nil, err
	}
	return rowsToTasks(rows)
}

func (store *Store) SearchTasks(search string, limit int) ([]Task, error) {
	searchParam := "%" + search + "%"
	rows, err := store.db.Query(searchTasksQuery(), searchParam, searchParam, limit)
	if err != nil {
		return nil, err
	}
	return rowsToTasks(rows)
}

func (store *Store) GetTasksById(id int64) (Task, error) {
	var task Task
	err := store.db.QueryRow(getTasksByIdQuery(), id).Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (store *Store) UpdateTask(task *Task) error {
	result, err := store.db.Exec(updateTaskQuery(), task.Date, task.Title, task.Comment, task.Repeat, task.Id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("rows affected = 0")
	}

	return err
}

func (store *Store) DeleteTask(id int64) error {
	result, err := store.db.Exec(deleteTaskQuery(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("rows affected = 0")
	}

	return err
}

func (store *Store) createScheduler() error {
	query := `
    CREATE TABLE IF NOT EXISTS scheduler (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL,
        title TEXT NOT NULL,
        comment TEXT,
        repeat TEXT
    );
    `
	_, err := store.db.Exec(query)
	return err
}

func rowsToTasks(rows *sql.Rows) ([]Task, error) {
	tasks := []Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.Id, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func getTasksQuery() string {
	return `SELECT
		id,
		date,
		title,
		comment,
		repeat
	FROM
		scheduler
	ORDER BY
		date ASC
	LIMIT ?`
}

func getTasksByDateQuery() string {
	return `SELECT
		id,
		date,
		title,
		comment,
		repeat
	FROM
		scheduler
	WHERE
		date = ?
	ORDER BY
		date ASC
	LIMIT ?`
}

func searchTasksQuery() string {
	return `SELECT
		id,
		date,
		title,
		comment,
		repeat
	FROM
		scheduler
	WHERE
		title LIKE ? OR comment LIKE ?
	ORDER BY
		date ASC
	LIMIT ?`
}

func getTasksByIdQuery() string {
	return `SELECT
		id,
		date,
		title,
		comment,
		repeat
	FROM
		scheduler
	WHERE
		id = ?`
}

func updateTaskQuery() string {
	return `UPDATE
		scheduler
	SET
		date = ?,
		title = ?,
		comment = ?,
		repeat = ?
	WHERE
		id = ?`
}

func deleteTaskQuery() string {
	return `DELETE FROM
		scheduler
	WHERE
		id = ?`
}
