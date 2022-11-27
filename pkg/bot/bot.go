package bot

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

type Bot struct {
	bot       *tgbotapi.BotAPI
	sendQueue *KitToSend
	DB        *sql.DB
}

func NewBot(bot *tgbotapi.BotAPI, dbTasks *sql.DB) *Bot {
	return &Bot{
		bot:       bot,
		sendQueue: NewKitToSend(),
		DB:        dbTasks,
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
		Command:     "/start",
		Description: "Сброс статистики",
	})
	for update := range updates {
		b.handleUpdate(&update)
		/*var err error
		chatID := update.SentFrom().ID
		if update.Poll != nil {
			fmt.Println(update.Poll.Options)
		} else {
			PrintReceive(&update)
		}
		if b.param, b.changePassLength, err = GetData("UsersDB", "Options", int(chatID)); err != nil {
			if err := InsertData("UsersDB", "Options", int(chatID), b.param, b.changePassLength); err != nil {
				log.Println(err)
			}
		}
		if update.Message != nil {
			err = b.handleMessage(update.Message, int(update.Message.Chat.ID))
		} else if update.CallbackQuery != nil {
			err = b.handleCallbackQuery(update.CallbackQuery, int(chatID))
		}
		if err != nil {
			log.Println(err)
		}*/
	}
}

func lastSend(chatID int, messageTimes map[int]time.Time) bool {
	if val, ok := messageTimes[chatID]; ok {
		dt := time.Now()
		return dt.After(val.Add(time.Second))
	}
	return true
}

func (b *Bot) TimeStart() {
	messageTimes := make(map[int]time.Time)
	timer := time.NewTicker(time.Second / 30)
	for range timer.C {
		b.sendQueue.Range(func(i int, v ItemToSend) bool {
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
