package botControl

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"log"
	bb "study-bot/pkg/botBasic"
	"time"
)

type ControlBot struct {
	bb.BotGeneral
	Chapters   []int
	iterations int
	timers     map[int64]*time.Timer
}

func NewControlBot(bot *tgbotapi.BotAPI, dbTasks *sql.DB) *ControlBot {
	var _ bb.BotInterface = &ControlBot{}
	viper.SetConfigName("options")
	viper.AddConfigPath(".")
	iterations := 10
	if err := viper.ReadInConfig(); err == nil {
		iterations = viper.GetInt("options.iterations")
	}
	return &ControlBot{
		BotGeneral: *bb.NewBot(bot, dbTasks),
		Chapters:   nil,
		iterations: iterations,
		timers:     make(map[int64]*time.Timer),
	}
}

func (b *ControlBot) Run() {
	log.Printf("Authorized on account %s", b.Bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.Bot.GetUpdatesChan(u)
	b.Bot.Self.CanJoinGroups = false
	b.SendCommands(tgbotapi.BotCommand{
		Command:     "/user",
		Description: "Получить информацию о тесте пользователя",
	}, tgbotapi.BotCommand{
		Command:     "/clearuser",
		Description: "Удалить информацию о пользователе",
	}, tgbotapi.BotCommand{
		Command:     "/clear",
		Description: "Очистить базу результатов тестов",
	}, tgbotapi.BotCommand{
		Command:     "/usertasks",
		Description: "Получить информацию о вопросах для пользователя",
	}, tgbotapi.BotCommand{
		Command:     "/task",
		Description: "Вывести вопрос по его номеру",
	}, tgbotapi.BotCommand{
		Command:     "/getdb",
		Description: "Получить полную базу",
	}, tgbotapi.BotCommand{
		Command:     "/update",
		Description: "Инструкция по обновлению базы вопросов и прочим настройкам",
	})
	for update := range updates {
		b.HandleUpdate(&update)
	}
}

func lastSend(chatID int64, messageTimes map[int64]time.Time) bool {
	if val, ok := messageTimes[chatID]; ok {
		dt := time.Now()
		return dt.After(val.Add(time.Second))
	}
	return true
}

func (b *ControlBot) TimeStart() {
	messageTimes := make(map[int64]time.Time)
	timer := time.NewTicker(time.Second / 30)
	defer timer.Stop()
	for range timer.C {
		b.SendQueue.Range(func(i int64, v bb.ItemToSend) bool {
			if v.Queue > 0 && lastSend(i, messageTimes) {
				err := b.Send(i)
				if err != nil {
					log.Println(err)
				}
				messageTimes[i] = time.Now()
				return false
			}
			if v.Queue <= 0 {
				b.SendQueue.Delete(i)
			}
			return true
		})
	}
}
