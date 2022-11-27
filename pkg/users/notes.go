package users

import (
	"database/sql"
	"fmt"
	"log"
	"study-bot/pkg/task"
)

type Note struct {
	UserID int
	PollID string
	TaskID int
	Answer int
}

func NewNote(userID int, pollID string, taskID int, answer int) *Note {
	return &Note{
		UserID: userID,
		PollID: pollID,
		TaskID: taskID,
		Answer: answer,
	}
}

func CreateNotes(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS notes(" +
		"userID		INTEGER," +
		"pollID		TEXT PRIMARY KEY," +
		"taskID 	INTEGER," +
		"answer		INTEGER)")
	if err != nil {
		log.Fatal(err)
	}
}

func InputNote(db *sql.DB, note Note) {
	_, err := db.Exec("INSERT INTO notes(userID, pollID, taskID, answer) VALUES (?,?,?,?)",
		note.UserID, note.PollID, note.TaskID, note.Answer)
	if err != nil {
		log.Println(err)
	}
}

func UpdateAnswer(db *sql.DB, userID int, pollID string, answer int) {
	fmt.Println(answer, userID)
	_, err := db.Exec("UPDATE notes SET answer=$1 WHERE userID=$2 AND pollID=$3", answer, userID, pollID)
	if err != nil {
		log.Println(err)
	}
}

func GetTask(db *sql.DB, pollID string, userID int) (int, int) {
	taskRow := db.QueryRow("SELECT * FROM notes WHERE pollID = $1 AND userID = $2", pollID, userID)
	if taskRow.Err() != nil {
		return -1, -1
	}
	var taskNote Note
	err := taskRow.Scan(&taskNote.UserID, &taskNote.PollID, &taskNote.TaskID, &taskNote.Answer)
	if err != nil {
		return -1, -1
	}
	t, err := task.GetQuestion(db, taskNote.TaskID)
	if err != nil {
		return -1, -1
	}
	return taskNote.TaskID, t.Category
}

// для генерации вопросов
// нуждается в корректировке

func HasTask(db *sql.DB, userID int, taskID int) bool {
	taskNote := db.QueryRow("SELECT * FROM notes WHERE userID = $1 AND taskID = $2", userID, taskID)
	if taskNote.Err() == nil {
		return true
	}
	return false
}

func ClearUser(db *sql.DB, userID int) {
	_, err := db.Exec("DELETE FROM notes WHERE userID = $1", userID)
	if err != nil {
		log.Println(err)
	}
}

func GetLastStats(db *sql.DB, user *User) {
	rows, err := db.Query("SELECT taskID, answer FROM notes WHERE userID = $1", user.ID)
	if err != nil || rows == nil {
		return
	}
	user.PollRun = 0
	user.Corrects = 0
	user.Worst = make(map[int]int)
	user.Chapters = make(map[int]int)
	for rows.Next() {
		var taskID, answer int
		err := rows.Scan(&taskID, &answer)
		if err != nil || answer == 0 {
			continue
		}
		user.PollRun++
		t, err := task.GetQuestion(db, taskID)
		if err != nil {
			continue
		}
		chapter := t.Category
		user.Chapters[chapter]++
		if task.CheckQuestion(db, taskID, answer) {
			user.Corrects++
		} else {
			user.Worst[chapter]++
		}
	}
}
