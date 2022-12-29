package databases

import (
	"database/sql"
	"encoding/json"
	"fmt"
	//_ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
	"log"
)

type User struct {
	ID       int64
	Username string
	PollRun  int
	Corrects int
	Regime   int
	Worst    map[int]int
	Chapters map[int]int
}

func NewUser(ID int64) *User {
	return &User{
		ID:       ID,
		Username: "",
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
		"userID 	bigint PRIMARY KEY, " +
		"username	TEXT, " +
		"pollRun	INTEGER, " +
		"corrects 	INTEGER, " +
		"regime 	INTEGER, " +
		"worst 		jsonb, " +
		"chapters 	jsonb)")
	if err != nil {
		log.Fatal(err)
	}
}

func ClearUsers(db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		log.Fatal(err)
	}
	CreateUsers(db)
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
	_, err = db.Exec("INSERT INTO users(userID, username, pollRun, corrects, regime, worst, chapters) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		user.ID, user.Username, user.PollRun, user.Corrects, user.Regime, m, c)
	if err != nil {
		log.Println(err)
	}
}

func GetUser(db *sql.DB, userID int64) (User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE userID = $1", userID)
	if row.Err() != nil {
		return User{ID: 0, Username: "", PollRun: 0, Corrects: 0, Regime: 0, Worst: nil, Chapters: nil}, row.Err()
	}
	user1 := *NewUser(userID)
	var m, c []byte
	err := row.Scan(&user1.ID, &user1.Username, &user1.PollRun, &user1.Corrects, &user1.Regime, &m, &c)
	if err != nil {
		return User{ID: 0, Username: "", PollRun: 0, Corrects: 0, Regime: 0, Worst: nil, Chapters: nil}, fmt.Errorf("unknown user")
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
	_, err = db.Exec("UPDATE users SET pollRun = $1, corrects = $2, regime = $3, worst = $4, chapters = $5, username = $6 where userID = $7",
		user.PollRun, user.Corrects, user.Regime, m, c, user.Username, user.ID)
	if err != nil {
		log.Println(err)
	}
}

func GetAllUsers(db *sql.DB) string {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil || rows == nil {
		return ""
	}
	result := "userID\t| username\t\t\t| статус\t| правильных\t| ошибки в главах\t| правильные ответы\n"
	for rows.Next() {
		var userID, pollRun, corrects, regime int64
		var username, worst, chapters string
		err := rows.Scan(&userID, &username, &pollRun, &corrects, &regime, &worst, &chapters)
		if err != nil || userID == 0 {
			continue
		}
		result += fmt.Sprintf("%v\t| %v\t\t\t| %v\t| %v\t| %v\t| %v\n", userID, username, pollRun, corrects, worst, chapters)
	}
	return result
}

func GetUserByName(db *sql.DB, username string) (User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE username = $1", username)
	if row.Err() != nil {
		return User{ID: 0, Username: "", PollRun: 0, Corrects: 0, Regime: 0, Worst: nil, Chapters: nil}, row.Err()
	}
	user1 := User{0, "", 0, 0, 0, nil, nil}
	var m, c []byte
	err := row.Scan(&user1.ID, &user1.Username, &user1.PollRun, &user1.Corrects, &user1.Regime, &m, &c)
	if err != nil {
		return User{ID: 0, Username: "", PollRun: 0, Corrects: 0, Regime: 0, Worst: nil, Chapters: nil}, fmt.Errorf("unknown user")
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
