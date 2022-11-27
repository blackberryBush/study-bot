package users

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type User struct {
	ID       int
	PollRun  int
	Corrects int
	Regime   int
	Worst    map[int]int
}

func NewUser(ID int) *User {
	return &User{
		ID:       ID,
		PollRun:  -1,
		Corrects: 0,
		Regime:   0,
		Worst:    make(map[int]int),
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
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users(" +
		"userID 	INTEGER PRIMARY KEY, " +
		"pollRun	INTEGER, " +
		"corrects 	INTEGER, " +
		"regime 	INTEGER, " +
		"worst 		BLOB)")
	if err != nil {
		log.Fatal(err)
	}
}

func InsertUser(db *sql.DB, user User) {
	m, err := json.Marshal(user.Worst)
	if err != nil {
		fmt.Println(err)
		return
	}
	// запрос наверно переделать
	//		"ON DUPLICATE KEY UPDATE PollRun=%d, Corrects=%d, Regime=%d, Worst=\"%s\"",
	_, err = db.Exec("INSERT INTO users(userID, pollRun, corrects, regime, worst) VALUES (?,?,?,?,?)",
		user.ID, user.PollRun, user.Corrects, user.Regime, m)
	if err != nil {
		log.Println(err)
	}
}

func GetUser(db *sql.DB, userID int) (User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE userID = $1", userID)
	if row.Err() != nil {
		return User{ID: 0, PollRun: 0, Corrects: 0, Regime: 0, Worst: nil}, row.Err()
	}
	user1 := *NewUser(userID)
	var m []byte
	err := row.Scan(&user1.ID, &user1.PollRun, &user1.Corrects, &user1.Regime, &m)
	if err != nil {
		return User{ID: 0, PollRun: 0, Corrects: 0, Regime: 0, Worst: nil}, fmt.Errorf("unknown user")
	}
	err = json.Unmarshal(m, &user1.Worst)
	if err != nil {
		fmt.Println(err)
	}
	return user1, err
}

func UpdateUser(db *sql.DB, user User) {
	m, err := json.Marshal(user.Worst)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec("UPDATE users SET pollRun = $1, corrects = $2, regime = $3, worst = $4 where userID = $5",
		user.PollRun, user.Corrects, user.Regime, m, user.ID)
	if err != nil {
		log.Println(err)
	}
}

/*
func UpdatePollRun(db *sql.DB, poll int, userID int) {
	query := fmt.Sprintf("SELECT * FROM users WHERE userID = %d", userID)
	row := db.QueryRow(query)
	if row.Err() == nil {
		query = fmt.Sprintf("UPDATE users SET PollRun = %d WHERE userID = %d", poll, UserID)
		_, err := db.Exec(query)
	} else {

	}
}*/

// таблица юзер - id опроса - id вопроса

func CreateNotes(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS notes(" +
		"userID		INTEGER," +
		"pollID		TEXT PRIMARY KEY," +
		"taskID 	INTEGER)")
	if err != nil {
		log.Fatal(err)
	}
}

func InputNote(db *sql.DB, note Note) {
	_, err := db.Exec("INSERT INTO notes(userID, pollID, taskID) VALUES (?,?,?)",
		note.UserID, note.PollID, note.TaskID)
	if err != nil {
		log.Println(err)
	}
}

func GetTask(db *sql.DB, pollID string, userID int) int {
	taskRow := db.QueryRow("SELECT * FROM notes WHERE pollID = $1 AND userID = $2", pollID, userID)
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
