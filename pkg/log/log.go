package log

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

const (
	messageUndefined = -1
	updatePollAnswer = iota
	messageCommand
	messageSticker
	messageText
	messageRegimeNo
	messageRegimeYes
	callbackQuery
)

func PrintReceive(update *tgbotapi.Update, updateType int, chatID int64) {
	user := update.SentFrom()
	if user == nil && updateType == updatePollAnswer {
		user = &update.PollAnswer.User
	}
	ans := "[idk who is this]"
	if user != nil {
		ans = fmt.Sprintf("[@%s][%s", user.UserName, user.FirstName)
		if user.LastName != "" {
			ans = fmt.Sprintf("%s %s", ans, user.LastName)
		}
		ans = fmt.Sprintf("%s][%d]", ans, chatID)
	}
	switch updateType {
	case updatePollAnswer:
		log.Printf("NEW POLL ANSWER:	%s %v\n", ans, update.PollAnswer.OptionIDs)
	case messageCommand:
		log.Printf("NEW COMMAND:	%s %s\n", ans, update.Message.Text)
	case messageSticker:
		log.Printf("NEW STICKER:	%s %s\n", ans, update.Message.Sticker.Emoji)
	case messageText:
		log.Printf("NEW MESSAGE:	%s %s\n", ans, update.Message.Text)
	case messageRegimeNo:
		log.Printf("NEW CHANGE REGIME:	%s [NO]\n", ans)
	case messageRegimeYes:
		log.Printf("NEW CHANGE REGIME:	%s [YES]\n", ans)
	case callbackQuery:
		log.Printf("NEW CALLBACK:	%s %s\n", ans, update.CallbackData())
	default:
		log.Printf("NEW INDEFINITE:	%s\n", ans)
	}
}

func PrintSent(c *tgbotapi.Chattable) {
	switch (*c).(type) {
	case tgbotapi.MessageConfig:
		text := strings.Replace((*c).(tgbotapi.MessageConfig).Text, "\n", " -> ", -1)
		log.Printf("SENT MESSAGE:	[%d] %s", (*c).(tgbotapi.MessageConfig).ChatID, text)
	case tgbotapi.SendPollConfig:
		log.Printf("SENT POLL:		[%d] %s", (*c).(tgbotapi.SendPollConfig).ChatID, (*c).(tgbotapi.SendPollConfig).Question)
	case tgbotapi.CallbackConfig:
		log.Printf("SENT CALLBACK:	[%s] %s", (*c).(tgbotapi.CallbackConfig).CallbackQueryID, (*c).(tgbotapi.CallbackConfig).Text)
	case tgbotapi.DeleteMessageConfig:
		log.Printf("DELETE MESSAGE:	[%d] %d", (*c).(tgbotapi.DeleteMessageConfig).ChatID, (*c).(tgbotapi.DeleteMessageConfig).MessageID)
	case tgbotapi.StickerConfig:
		log.Printf("SENT STICKER:	[%d] %s", (*c).(tgbotapi.StickerConfig).ChatID, (*c).(tgbotapi.StickerConfig).File)
	case tgbotapi.PhotoConfig:
		log.Printf("SENT PHOTO:		[%d] %s", (*c).(tgbotapi.PhotoConfig).ChatID, "*some photo*")
	default:
		log.Println("SENT, but idk what is this")
	}
}
