package botBasic

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"study-bot/pkg/databases"
	"study-bot/pkg/log"
)

func (b *BotGeneral) Send(chatID int64) (err error) {
	b.SendQueue.QueueDec(chatID)
	item, ok := b.SendQueue.Load(chatID)
	if !ok {
		return fmt.Errorf("reading error")
	}
	data := <-item.data
	if item.Queue > 5 {
		data = *NewChattable(tgbotapi.NewMessage(int64(chatID), "Не флуди!"))
	}
	log.PrintSent(&data.data)
	switch data.data.(type) {
	case tgbotapi.MessageConfig, tgbotapi.StickerConfig, tgbotapi.PhotoConfig, tgbotapi.DocumentConfig:
		_, err = b.Bot.Send(data.data)
	case tgbotapi.CallbackConfig, tgbotapi.DeleteMessageConfig:
		_, err = b.Bot.Request(data.data)
	case tgbotapi.SendPollConfig:
		msg, err := b.Bot.Send(data.data.(tgbotapi.SendPollConfig))
		databases.InputNote(b.DB, *databases.NewNote(chatID, msg.Poll.ID, data.option.taskID, 0, data.option.correct))
		return err
	default:
		err = fmt.Errorf("undefined type")
	}
	return err
}

func (b *BotGeneral) Pull(chatID int64, c Chattable) {
	if v, ok := b.SendQueue.Load(chatID); ok {
		if v.Queue > 7 {
			return
		}
		b.SendQueue.QueueInc(chatID)
		v.data <- c
		b.SendQueue.StoreData(chatID, v.data)
	} else {
		item := *NewItemToSend()
		item.data <- c
		item.Queue++
		b.SendQueue.Store(chatID, item)
	}
}
