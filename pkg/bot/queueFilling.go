package bot

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

// для функций, которые будут заполнять очередь сообщениями в формате Chattable

func (b *Bot) PullText(text string, chatID int, reply int, args ...any) {
	msg := tgbotapi.NewMessage(int64(chatID), text)
	if reply > 0 {
		msg.ReplyToMessageID = reply
	}
	if len(args) != 0 {
		for i := range args {
			switch v := args[i].(type) {
			case []tgbotapi.KeyboardButton:
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(v)
			case tgbotapi.ReplyKeyboardRemove:
				msg.ReplyMarkup = v
			}
		}
	}
	go b.Pull(chatID, *NewChattable(msg))
}

func (b *Bot) PullSticker(name string, chatID int, byFile bool, reply int) error {
	var stickerID tgbotapi.RequestFileData
	if byFile {
		photoBytes, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		stickerID = tgbotapi.FileBytes{
			Name:  "picture",
			Bytes: photoBytes,
		}
	} else {
		if fileID := tgbotapi.FileID(name); fileID == "" {
			return errors.New("fileID error")
		} else {
			stickerID = tgbotapi.FileID(name)
		}
	}
	msg := tgbotapi.NewSticker(int64(chatID), stickerID)
	if reply > 0 {
		msg.ReplyToMessageID = reply
	}
	go b.Pull(chatID, *NewChattable(msg))
	return nil
}

func (b *Bot) PullPoll(id int, question string, chatID int, reply int, isMultiple bool, correct int, ans ...string) error {
	msg := tgbotapi.NewPoll(int64(chatID), question, ans...)
	if reply > 0 {
		msg.ReplyToMessageID = reply
	}
	msg.IsAnonymous = false
	msg.AllowsMultipleAnswers = isMultiple
	go b.Pull(chatID, *NewChattable(msg, id, correct))
	return nil
}

func (b *Bot) PullDeleteMessage(messageID int, chatID int) error {
	msg := tgbotapi.NewDeleteMessage(int64(chatID), messageID)
	go b.Pull(chatID, *NewChattable(msg))
	return nil
}

func (b *Bot) PullPicture(filename string, chatID int, reply int) error {
	photoBytes, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
	}
	photoFileBytes := tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: photoBytes,
	}
	msg := tgbotapi.NewPhoto(int64(chatID), photoFileBytes)
	if reply > 0 {
		msg.ReplyToMessageID = reply
	}
	go b.Pull(chatID, *NewChattable(msg))
	return nil
}

// Отправляют апдейт сразу, не ожидая очереди

func (b *Bot) SendCommands(cmd ...tgbotapi.BotCommand) {
	msg := tgbotapi.NewSetMyCommands(cmd...)
	_, err := b.bot.Send(msg)
	if err != nil {
		return
	}
}
