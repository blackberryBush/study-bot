package bot

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"log"
	"time"
)

type TesterBot struct {
	Bot
}

type Bot struct {
	bot        *tgbotapi.BotAPI
	sendQueue  *KitToSend
	DB         *sql.DB
	Chapters   []int
	iterations int
	timers     map[int64]*time.Timer
}

func NewBot(bot *tgbotapi.BotAPI, dbTasks *sql.DB) *Bot {
	viper.SetConfigName("options")
	viper.AddConfigPath(".")
	iterations := 10
	if err := viper.ReadInConfig(); err == nil {
		iterations = viper.GetInt("options.iterations")
	}
	return &Bot{
		bot:        bot,
		sendQueue:  NewKitToSend(),
		DB:         dbTasks,
		Chapters:   nil,
		iterations: iterations,
		timers:     make(map[int64]*time.Timer),
	}
}

func (b *Bot) Run() {
	log.Printf("Authorized on account %s", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)
	b.bot.Self.CanJoinGroups = false
	b.SendCommands(tgbotapi.BotCommand{
		Command:     "/test",
		Description: "Запуск тестирования",
	}, tgbotapi.BotCommand{
		Command:     "/getstats",
		Description: "Вывести текущий результат",
	}, tgbotapi.BotCommand{
		Command:     "/study",
		Description: "Открыть учебник",
	})
	for update := range updates {
		b.handleUpdate(&update)
	}
}

func lastSend(chatID int64, messageTimes map[int64]time.Time) bool {
	if val, ok := messageTimes[chatID]; ok {
		dt := time.Now()
		return dt.After(val.Add(time.Second))
	}
	return true
}

func (b *Bot) TimeStart() {
	messageTimes := make(map[int64]time.Time)
	timer := time.NewTicker(time.Second / 30)
	defer timer.Stop()
	for range timer.C {
		b.sendQueue.Range(func(i int64, v ItemToSend) bool {
			if v.queue > 0 && lastSend(i, messageTimes) {
				err := b.Send(i)
				if err != nil {
					log.Println(err)
				}
				messageTimes[i] = time.Now()
				return false
			}
			if v.queue <= 0 {
				b.sendQueue.Delete(i)
			}
			return true
		})
	}
}
