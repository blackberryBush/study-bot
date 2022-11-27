package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"study-bot/pkg/users"
)

func (b *Bot) Send(chatID int) (err error) {
	b.sendQueue.QueueDec(chatID)
	item, ok := b.sendQueue.Load(chatID)
	if !ok {
		return fmt.Errorf("reading error")
	}
	data := <-item.data
	if item.queue > 5 {
		data = *NewChattable(tgbotapi.NewMessage(int64(chatID), "Не флуди!"))
	}
	PrintSent(&data.data)
	switch data.data.(type) {
	case tgbotapi.MessageConfig, tgbotapi.StickerConfig, tgbotapi.PhotoConfig:
		_, err = b.bot.Send(data.data)
	case tgbotapi.CallbackConfig, tgbotapi.DeleteMessageConfig:
		_, err = b.bot.Request(data.data)
	case tgbotapi.SendPollConfig:
		msg, err := b.bot.Send(data.data)
		users.InputNote(b.DB, *users.NewNote(chatID, msg.Poll.ID, data.option, 0))
		return err
	default:
		err = fmt.Errorf("undefined type")
	}
	return err
}

func (b *Bot) Pull(chatID int, c Chattable) {
	if v, ok := b.sendQueue.Load(chatID); ok {
		if v.queue > 7 {
			return
		}
		b.sendQueue.QueueInc(chatID)
		v.data <- c
		b.sendQueue.StoreData(chatID, v.data)
	} else {
		item := *NewItemToSend()
		item.data <- c
		item.queue++
		b.sendQueue.Store(chatID, item)
	}
}
