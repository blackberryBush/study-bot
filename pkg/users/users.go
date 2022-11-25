package users

import (
	"database/sql"
	"fmt"
	"log"
)

type User struct {
	ID       int
	pollRun  int
	corrects int
	regime   bool
	worst    []int
}

type Note struct {
	User   int
	PollID string
	TaskID int
}

func NewNote(user int, pollID string, taskID int) *Note {
	return &Note{
		User:   user,
		PollID: pollID,
		TaskID: taskID,
	}
}

// таблица юзер - статус опроса,кол-во вопросов - кол-во верных - режим - главы по качеству

func CreateUsers(db *sql.DB) {
	query := "CREATE TABLE IF NOT EXISTS users(" +
		"user		INTEGER PRIMARY KEY," +
		"poll_run	INTEGER," +
		"corrects 	INTEGER," +
		"regime 	INTEGER," +
		"worst 		TEXT)"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

/*
func SetUserPoll(db *sql.DB, poll int, userID int) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE chatID = %d", tableName, chatID)
	row := db.QueryRow(query)
	if row.Err() == nil {
		query = fmt.Sprintf("UPDATE users SET poll_run = %d WHERE user = %d", poll, UserID)
		_, err := db.Exec(query)
	} else {

	}
}*/

// таблица юзер - id опроса - id вопроса

func CreateNotes(db *sql.DB) {
	query := "CREATE TABLE IF NOT EXISTS notes(" +
		"user		INTEGER," +
		"poll_ID	TEXT PRIMARY KEY," +
		"task_ID 	INTEGER)"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func InputNote(db *sql.DB, note Note) {
	query := fmt.Sprintf("INSERT INTO notes(user, poll_id, task_ID) VALUES (%d,%s,%d);",
		note.User, note.PollID, note.TaskID)
	_, err := db.Exec(query)
	if err != nil {
		log.Println(err)
	}
}

func HasQuestion(db *sql.DB, userID int, taskID int) bool {
	query := fmt.Sprintf("SELECT * FROM notes WHERE user = %d AND task_ID = %d", userID, taskID)
	taskNote := db.QueryRow(query)
	if taskNote.Err() != nil {
		return true
	}
	return false
}

func ClearUser(db *sql.DB, userID int) {
	query := fmt.Sprintf("DELETE FROM notes WHERE user = %d", userID)
	_, err := db.Exec(query)
	if err != nil {
		log.Println(err)
	}
}
