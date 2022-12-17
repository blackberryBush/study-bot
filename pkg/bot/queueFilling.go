package bot

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

// для функций, которые будут заполнять очередь сообщениями в формате Chattable

func (b *BotGeneral) PullText(text string, chatID int64, reply int, args ...any) {
	msg := tgbotapi.NewMessage(chatID, text)
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

func (b *BotGeneral) PullSticker(name string, chatID int64, byFile bool, reply int) error {
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
	msg := tgbotapi.NewSticker(chatID, stickerID)
	if reply > 0 {
		msg.ReplyToMessageID = reply
	}
	go b.Pull(chatID, *NewChattable(msg))
	return nil
}

func cutString(text string, length int) string {
	runes := []rune(text)
	if len(runes) >= length {
		return string(runes[:length])
	}
	return text
}

func (b *BotGeneral) PullPoll(id int, question string, chatID int64, reply int, isMultiple bool, correct int, ans ...string) error {
	for i := range ans {
		ans[i] = cutString(ans[i], 100)
	}
	question = cutString(question, 255)
	msg := tgbotapi.NewPoll(chatID, question, ans...)
	if reply > 0 {
		msg.ReplyToMessageID = reply
	}
	msg.IsAnonymous = false
	msg.AllowsMultipleAnswers = isMultiple
	go b.Pull(chatID, *NewChattable(msg, id, correct))
	return nil
}

func (b *BotGeneral) PullDeleteMessage(messageID int, chatID int64) error {
	msg := tgbotapi.NewDeleteMessage(chatID, messageID)
	go b.Pull(chatID, *NewChattable(msg))
	return nil
}

func (b *BotGeneral) PullPicture(filename string, chatID int64, reply int) error {
	photoBytes, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
	}
	photoFileBytes := tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: photoBytes,
	}
	msg := tgbotapi.NewPhoto(chatID, photoFileBytes)
	if reply > 0 {
		msg.ReplyToMessageID = reply
	}
	go b.Pull(chatID, *NewChattable(msg))
	return nil
}

func (b *BotGeneral) PullFile(filename string, chatID int64, reply int, newFilename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
	}
	fileBytes := tgbotapi.FileBytes{
		Name:  newFilename,
		Bytes: file,
	}
	msg := tgbotapi.NewDocument(chatID, fileBytes)
	if reply > 0 {
		msg.ReplyToMessageID = reply
	}
	go b.Pull(chatID, *NewChattable(msg))
	return nil
}

// Отправляют апдейт сразу, не ожидая очереди

func (b *BotGeneral) SendCommands(cmd ...tgbotapi.BotCommand) {
	msg := tgbotapi.NewSetMyCommands(cmd...)
	_, err := b.Bot.Send(msg)
	if err != nil {
		return
	}
}
