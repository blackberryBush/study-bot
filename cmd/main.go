package main

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"log"
	"study-bot/pkg/bot"
	"study-bot/pkg/users"
)

func getToken() string {
	viper.SetConfigName("token")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return ""
	}
	return viper.GetString("token")
}

func main() {
	botAPI, err := tgbotapi.NewBotAPI(getToken())
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatal(err)
	}
	b := bot.NewBot(botAPI, db)
	users.CsvToSQLite("tasks.csv", b.DB)
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
