package users

import (
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	rand2 "golang.org/x/exp/rand"
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

func GetQuestion(db *sql.DB, number int) (Question, error) {
	row := db.QueryRow("SELECT * FROM tasks WHERE ID = $1", number)
	var chapter, ID, correct, picture int
	var question, option1, option2, option3, option4 string
	err := row.Scan(&chapter, &ID, &question, &option1, &option2, &option3, &option4, &correct, &picture)
	if err != nil || correct < 1 {
		return Question{}, fmt.Errorf("invalid question number")
	}
	return *NewQuestion(chapter, ID, question, correct, picture, option1, option2, option3, option4), nil
}

func CountChapters(db *sql.DB) int {
	row := db.QueryRow("SELECT COUNT(DISTINCT chapter) FROM tasks")
	if row.Err() != nil || row == nil {
		return -1
	}
	count := 0
	err := row.Scan(&count)
	if err != nil {
		return -1
	}
	return count
}

func GetRandomInt(max int) int {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return -1
	}
	return int(r.Int64())
}

func GetRandomQuestionNumber(db *sql.DB, number int, chapters int, userID int) int {
	currentChapter := (number-1)%chapters + 1
	rows, err := db.Query("SELECT tasks.ID FROM tasks WHERE tasks.chapter = $1 EXCEPT SELECT notes.taskID FROM notes WHERE notes.userID = $2", currentChapter, userID)
	f := func(rows *sql.Rows) []int {
		var questions []int
		for rows.Next() {
			var chapter int
			err := rows.Scan(&chapter)
			if err != nil || chapter == 0 {
				continue
			}
			questions = append(questions, chapter)
		}
		return questions
	}
	if err != nil {
		return -1
	}
	questions := f(rows)
	if len(questions) == 0 {
		rows, err = db.Query("SELECT ID FROM tasks WHERE chapter = $1", currentChapter)
		if err != nil || rows == nil {
			return -1
		}
		questions = f(rows)
	}
	if len(questions) == 0 {
		return -1
	}

	return questions[GetRandomInt(len(questions))]
}

func (q *Question) MixQuestion() {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return
	}
	rand2.Seed(binary.LittleEndian.Uint64(b[:]))
	rand2.Shuffle(len(q.Variants),
		func(i, j int) {
			q.Variants[i], q.Variants[j] = q.Variants[j], q.Variants[i]
			if q.Correct == i+1 {
				q.Correct = j + 1
			} else if q.Correct == j+1 {
				q.Correct = i + 1
			}
		})

}
