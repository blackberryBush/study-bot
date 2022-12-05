package users

import (
	"database/sql"
	"encoding/json"
	"fmt"
	//_ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
	"log"
)

type User struct {
	ID       int
	PollRun  int
	Corrects int
	Regime   int
	Worst    map[int]int
	Chapters map[int]int
}

func NewUser(ID int) *User {
	return &User{
		ID:       ID,
		PollRun:  -1,
		Corrects: 0,
		Regime:   0,
		Worst:    make(map[int]int),
		Chapters: make(map[int]int),
	}
}

// таблица юзер - статус опроса,кол-во вопросов - кол-во верных - режим - главы по качеству

func CreateUsers(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users(" +
		"userID 	INTEGER PRIMARY KEY, " +
		"pollRun	INTEGER, " +
		"corrects 	INTEGER, " +
		"regime 	INTEGER, " +
		"worst 		jsonb, " +
		"chapters 	jsonb)")
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
	c, err := json.Marshal(user.Chapters)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec("INSERT INTO users(userID, pollRun, corrects, regime, worst, chapters) VALUES ($1,$2,$3,$4,$5,$6)",
		user.ID, user.PollRun, user.Corrects, user.Regime, m, c)
	if err != nil {
		log.Println(err)
	}
}

func GetUser(db *sql.DB, userID int) (User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE userID = $1", userID)
	if row.Err() != nil {
		return User{ID: 0, PollRun: 0, Corrects: 0, Regime: 0, Worst: nil, Chapters: nil}, row.Err()
	}
	user1 := *NewUser(userID)
	var m, c []byte
	err := row.Scan(&user1.ID, &user1.PollRun, &user1.Corrects, &user1.Regime, &m, &c)
	if err != nil {
		return User{ID: 0, PollRun: 0, Corrects: 0, Regime: 0, Worst: nil, Chapters: nil}, fmt.Errorf("unknown user")
	}
	err = json.Unmarshal(m, &user1.Worst)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(c, &user1.Chapters)
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
	c, err := json.Marshal(user.Chapters)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec("UPDATE users SET pollRun = $1, corrects = $2, regime = $3, worst = $4, chapters = $5 where userID = $6",
		user.PollRun, user.Corrects, user.Regime, m, c, user.ID)
	if err != nil {
		log.Println(err)
	}
}
