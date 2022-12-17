package main

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"os"
	"study-bot/pkg/bot"
	"study-bot/pkg/users"
)

func getToken() string {
	viper.SetConfigName("token")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return ""
	}
	return viper.GetString("token.tester")
}

func NewPostgresDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres password=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PASSWORD")))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	botAPI, err := tgbotapi.NewBotAPI(getToken())
	if err != nil {
		log.Fatal(err)
	}
	db, err := NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}
	b := bot.NewBot(botAPI, db)
	users.CsvToPostgres("tasks.csv", b.DB)
	b.Chapters = users.CountChapters(b.DB)
	users.CreateUsers(b.DB)
	users.CreateNotes(b.DB)
	//
	//Start timer to send messages&callbacks
	go b.TimeStart()

	// Start checking for updates
	b.Run()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(b.DB)
}
