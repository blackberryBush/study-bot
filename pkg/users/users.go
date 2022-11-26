package users

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

type User struct {
	ID       int
	pollRun  int
	corrects int
	regime   bool
	worst    map[int]int
}

func NewUser(ID int) *User {
	return &User{
		ID:       ID,
		pollRun:  -1,
		corrects: 0,
		regime:   false,
		worst:    map[int]int{},
	}
}

type Note struct {
	UserID int
	PollID string
	TaskID int
}

func NewNote(userID int, pollID string, taskID int) *Note {
	return &Note{
		UserID: userID,
		PollID: pollID,
		TaskID: taskID,
	}
}

// таблица юзер - статус опроса,кол-во вопросов - кол-во верных - режим - главы по качеству

func CreateUsers(db *sql.DB) {
	query := "CREATE TABLE IF NOT EXISTS users(" +
		"userID		INTEGER PRIMARY KEY," +
		"pollRun	INTEGER," +
		"corrects 	INTEGER," +
		"regime 	INTEGER," +
		"worst 		TEXT)"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func InputUser(db *sql.DB, user User) {
	m, err := json.Marshal(user.worst)
	if err != nil {
		fmt.Println("Ошибка тут: ", err)
	}
	fmt.Printf("%s", m)
	query := fmt.Sprintf("INSERT INTO users(userID, pollRun, corrects, regime, worst) VALUES (%d,%d,%d,%d,\"%s\")",
		user.ID, user.pollRun, user.corrects, 0, m)
	_, err = db.Exec(query)
	if err != nil {
		log.Println("или тут:", err)
	}
}

/*
func UpdatePollRun(db *sql.DB, poll int, userID int) {
	query := fmt.Sprintf("SELECT * FROM users WHERE userID = %d", userID)
	row := db.QueryRow(query)
	if row.Err() == nil {
		query = fmt.Sprintf("UPDATE users SET pollRun = %d WHERE userID = %d", poll, UserID)
		_, err := db.Exec(query)
	} else {

	}
}*/

// таблица юзер - id опроса - id вопроса

func CreateNotes(db *sql.DB) {
	query := "CREATE TABLE IF NOT EXISTS notes(" +
		"userID		INTEGER," +
		"pollID		TEXT PRIMARY KEY," +
		"taskID 	INTEGER)"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func InputNote(db *sql.DB, note Note) {
	query := fmt.Sprintf("INSERT INTO notes(userID, pollID, taskID) VALUES (%d,%s,%d);",
		note.UserID, note.PollID, note.TaskID)
	_, err := db.Exec(query)
	if err != nil {
		log.Println(err)
	}
}

func GetTask(db *sql.DB, pollID string, userID int) int {
	query := fmt.Sprintf("SELECT * FROM notes WHERE pollID = %s AND userID = %d", pollID, userID)
	taskRow := db.QueryRow(query)
	if taskRow.Err() != nil {
		return -1
	}
	var taskNote Note
	err := taskRow.Scan(&taskNote.UserID, &taskNote.PollID, &taskNote.TaskID)
	if err != nil {
		return -1
	}
	return taskNote.TaskID
}

// для генерации вопросов
func HasTask(db *sql.DB, userID int, taskID int) bool {
	query := fmt.Sprintf("SELECT * FROM notes WHERE userID = %d AND taskID = %d", userID, taskID)
	taskNote := db.QueryRow(query)
	if taskNote.Err() == nil {
		return true
	}
	return false
}

func ClearUser(db *sql.DB, userID int) {
	query := fmt.Sprintf("DELETE FROM notes WHERE userID = %d", userID)
	_, err := db.Exec(query)
	if err != nil {
		log.Println(err)
	}
}
