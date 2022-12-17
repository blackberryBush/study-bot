package bot

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotGeneral struct {
	Bot       *tgbotapi.BotAPI
	SendQueue *KitToSend
	DB        *sql.DB
}

type BotInterface interface {
	Run()
	TimeStart()
	HandleUpdate(*tgbotapi.Update)
}

func NewBot(bot *tgbotapi.BotAPI, dbTasks *sql.DB) *BotGeneral {
	return &BotGeneral{
		Bot:       bot,
		SendQueue: NewKitToSend(),
		DB:        dbTasks,
	}
}
