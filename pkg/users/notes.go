package users

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type Note struct {
	UserID  int
	PollID  string
	TaskID  int
	Answer  int
	Correct int
}

func NewNote(userID int, pollID string, taskID int, answer int, correct int) *Note {
	return &Note{
		UserID:  userID,
		PollID:  pollID,
		TaskID:  taskID,
		Answer:  answer,
		Correct: correct,
	}
}

func CreateNotes(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS notes(" +
		"userID		INTEGER," +
		"pollID		TEXT PRIMARY KEY," +
		"taskID 	INTEGER," +
		"answer		INTEGER," +
		"correct 	INTEGER)")
	if err != nil {
		log.Fatal(err)
	}
}

func InputNote(db *sql.DB, note Note) {
	_, err := db.Exec("INSERT INTO notes(userID, pollID, taskID, answer, correct) VALUES ($1,$2,$3,$4,$5)",
		note.UserID, note.PollID, note.TaskID, note.Answer, note.Correct)
	if err != nil {
		log.Println(err)
	}
}

func UpdateAnswer(db *sql.DB, userID int, pollID string, answer int) {
	_, err := db.Exec("UPDATE notes SET answer=$1 WHERE userID=$2 AND pollID=$3", answer, userID, pollID)
	if err != nil {
		log.Println(err)
	}
}

func GetTask(db *sql.DB, pollID string, userID int) (int, int, int) {
	taskRow := db.QueryRow("SELECT * FROM notes WHERE pollID = $1 AND userID = $2", pollID, userID)
	if taskRow.Err() != nil {
		return -1, -1, -1
	}
	var taskNote Note
	err := taskRow.Scan(&taskNote.UserID, &taskNote.PollID, &taskNote.TaskID, &taskNote.Answer, &taskNote.Correct)
	if err != nil {
		return -1, -1, -1
	}
	t, err := GetQuestion(db, taskNote.TaskID)
	if err != nil {
		return -1, -1, -1
	}
	return taskNote.TaskID, t.Category, taskNote.Correct
}

func ClearUser(db *sql.DB, userID int) {
	_, err := db.Exec("DELETE FROM notes WHERE userID = $1", userID)
	if err != nil {
		log.Println(err)
	}
}

func GetLastStats(db *sql.DB, user *User) {
	rows, err := db.Query("SELECT taskID, answer, correct FROM notes WHERE userID = $1", user.ID)
	if err != nil || rows == nil {
		return
	}
	user.PollRun = 0
	user.Corrects = 0
	user.Worst = make(map[int]int)
	user.Chapters = make(map[int]int)
	for rows.Next() {
		var taskID, answer, correct int
		err := rows.Scan(&taskID, &answer, &correct)
		if err != nil || answer == 0 {
			continue
		}
		user.PollRun++
		t, err := GetQuestion(db, taskID)
		if err != nil {
			continue
		}
		chapter := t.Category
		user.Chapters[chapter]++
		if correct == answer {
			user.Corrects++
		} else {
			user.Worst[chapter]++
		}
	}
}
