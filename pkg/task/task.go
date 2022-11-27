package task

import (
	"crypto/rand"
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math/big"
	"os"
	"strconv"
)

// Question - вопрос в формате: категория, номер, вопрос, номер корректного, варианты
type Question struct {
	Category int
	Number   int
	Problem  string
	Correct  int
	Picture  int
	Variants []string
}

func NewQuestion(category int, number int, problem string, correct int, picture int, variants ...string) *Question {
	return &Question{
		Category: category,
		Number:   number,
		Problem:  problem,
		Correct:  correct,
		Picture:  picture,
		Variants: variants,
	}
}

func cutString(text string) string {
	if len(text) >= 99 {
		return text[:99]
	}
	return text
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func checkTask(data []string) (int, int, int, int, bool) {
	chapter, err := strconv.Atoi(data[0])
	if err != nil {
		chapter, err = strconv.Atoi(data[0][3:])
		if err != nil {
			return 0, 0, 0, 0, false
		}
	}
	ID, err := strconv.Atoi(data[1])
	if err != nil {
		return 0, 0, 0, 0, false
	}
	correct, err := strconv.Atoi(data[7])
	if err != nil {
		return 0, 0, 0, 0, false
	}
	picture, err := strconv.Atoi(data[8])
	if err != nil {
		return chapter, ID, correct, -1, true
	}
	if !fileExists(fmt.Sprintf("pics\\%d.png", picture)) {
		picture = -1
	}
	return chapter, ID, correct, picture, true
}

func CsvToSQLite(filename string, db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS tasks")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE tasks (" +
		"chapter	INTEGER NOT NULL," +
		"ID	INTEGER PRIMARY KEY NOT NULL," +
		"question	TEXT," +
		"option1	TEXT," +
		"option2	TEXT," +
		"option3	TEXT," +
		"option4	TEXT," +
		"correct	INTEGER NOT NULL," +
		"picture	INTEGER)")
	if err != nil {
		log.Fatal(err)
	}
	/// file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)
	reader := csv.NewReader(file)
	reader.Comma = ';'
	///
	for {
		data, err := reader.Read()
		if err != nil {
			break
		}
		if chapter, ID, correct, picture, check := checkTask(data); check {
			_, err = db.Exec("INSERT INTO tasks(chapter, ID, question, option1, option2, option3, option4, correct, picture) values (?,?,?,?,?,?,?,?,?)",
				chapter, ID, data[2], cutString(data[3]), cutString(data[4]), cutString(data[5]), cutString(data[6]), correct, picture)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func GetQuestion(db *sql.DB, number int) Question {
	row := db.QueryRow("SELECT * FROM tasks WHERE ID = $1", number)
	var chapter, ID, correct, picture int
	var question, option1, option2, option3, option4 string
	err := row.Scan(&chapter, &ID, &question, &option1, &option2, &option3, &option4, &correct, &picture)
	if err != nil {
		log.Println("invalid question number")
	}
	return *NewQuestion(chapter, ID, question, correct, picture, option1, option2, option3, option4)
}

func CheckQuestion(db *sql.DB, number int, answer int) bool {
	row := db.QueryRow("SELECT correct FROM tasks WHERE ID = $1", number)
	if row.Err() != nil {
		log.Println("invalid question number")
	}
	var correct int
	err := row.Scan(&correct)
	if err != nil {
		log.Println(err)
	}
	if answer != correct {
		return false
	}
	return true
}

func GenerateRandomQuestion(category int, number int) Question {
	length, _ := rand.Int(rand.Reader, big.NewInt(8))
	length.Add(length, big.NewInt(2))
	correct, _ := rand.Int(rand.Reader, length)
	ans := make([]string, length.Int64())
	for i := range ans {
		ans[i] = generatePassword(20)
	}
	return *NewQuestion(
		category,
		number,
		generatePassword(20),
		int(correct.Int64())-1,
		-1,
		ans...,
	)
}

func generatePassword(length int) string {
	kit := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,.%*()?@#$~"
	res := make([]byte, length)
	for i := range res {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(len(kit))))
		if err != nil {
			log.Fatal(err)
		}
		res[i] = kit[r.Int64()]
	}
	return string(res)
}