package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func PrintReceive(update *tgbotapi.Update) {
	ans := fmt.Sprintf("[@%s][%s", update.SentFrom().UserName, update.SentFrom().FirstName)
	if update.SentFrom().LastName != "" {
		ans = fmt.Sprintf("%s %s", ans, update.SentFrom().LastName)
	}
	ans = fmt.Sprintf("%s][%d]", ans, update.SentFrom().ID)
	if update.CallbackQuery != nil {
		log.Printf("NEW CALLBACK:   %s %s\n", ans, update.CallbackData())
		return
	}
	if update.Message != nil {
		ans = fmt.Sprintf("NEW MESSAGE:    %s %s", ans, update.Message.Text)
	}
	log.Println(ans)
}

func PrintSent(c *tgbotapi.Chattable) {
	switch (*c).(type) {
	case tgbotapi.MessageConfig:
		log.Printf("SENT MESSAGE:   [%d] %s", (*c).(tgbotapi.MessageConfig).ChatID, (*c).(tgbotapi.MessageConfig).Text)
	case tgbotapi.CallbackConfig:
		log.Printf("SENT CALLBACK:  [%s] %s", (*c).(tgbotapi.CallbackConfig).CallbackQueryID, (*c).(tgbotapi.CallbackConfig).Text)
	case tgbotapi.DeleteMessageConfig:
		log.Printf("DELETE MESSAGE: [%d] %d", (*c).(tgbotapi.DeleteMessageConfig).ChatID, (*c).(tgbotapi.DeleteMessageConfig).MessageID)
	case tgbotapi.StickerConfig:
		log.Printf("SENT STICKER:   [%d] %s", (*c).(tgbotapi.StickerConfig).ChatID, (*c).(tgbotapi.StickerConfig).File)
	default:
		log.Println("printing error")
	}
}
