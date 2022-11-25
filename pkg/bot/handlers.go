package bot

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"study-bot/pkg/task"
)

const (
	messageUndefined = -1
	//updatePoll       = iota
	updatePollAnswer = iota
	messageCommand
	messageSticker
	messageText
	messageRegimeNo
	messageRegimeYes
	callbackQuery
)

func (b *Bot) getUpdateType(update *tgbotapi.Update) (int, int) {
	// Закрыть дыру с неизвестными видами содержимого!!!!!

	/*if update.Poll != nil {
		return updatePoll, 0
	}*/
	if update.PollAnswer != nil {
		return updatePollAnswer, int(update.PollAnswer.User.ID)
	}
	if update.Message != nil {
		chatID := int(update.Message.From.ID)
		if update.Message.IsCommand() {
			return messageCommand, chatID
		}
		if b.regime && update.Message.Text == "Нет" {
			return messageRegimeNo, chatID
		}
		if b.regime && update.Message.Text == "Да" {
			return messageRegimeYes, chatID
		}
		if update.Message.Text == "" && update.Message.Sticker != nil {
			return messageSticker, chatID
		}
		if update.Message.Text != "" {
			return messageText, chatID
		}
		return messageUndefined, chatID
	}
	if update.CallbackQuery != nil {
		return callbackQuery, int(update.CallbackQuery.From.ID)
	}
	return messageUndefined, 0
}

func (b *Bot) handleUpdate(update *tgbotapi.Update) {
	updateType, chatID := b.getUpdateType(update)
	if b.regime && updateType != messageRegimeNo && updateType != messageRegimeYes {
		if updateType != updatePollAnswer {
			b.regime = false
			b.oprosRun = -1
			b.stat = 0
			b.PullText("Счёл это за отказ...", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
			return
		} else {
			b.regime = false
			b.PullText("Счёл это за продолжение...", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
		}
	}
	switch updateType {
	/*case updatePoll:
	b.handlePoll(update.Poll, chatID)*/
	case updatePollAnswer:
		b.handlePollAnswer(update.PollAnswer, chatID)
	case messageCommand:
		b.handleCommand(update.Message, chatID)
	case messageSticker:
		b.handleSticker(update.Message, chatID)
	case messageText:
		b.handleText(update.Message, chatID)
	case messageRegimeNo:
		b.handleRegimeNo(update.Message, chatID)
	case messageRegimeYes:
		b.handleRegimeYes(update.Message, chatID)
	case callbackQuery:
	default:
		b.handleUnknown()
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message, chatID int) error {
	switch message.Command() {
	case "start":
		b.oprosRun = -1
		b.stat = 0
	case "test":
		b.oprosRun = 0
		questionID := b.oprosRun + 1 // номер вопроса
		task := task.GetQuestion(b.DB, questionID)
		if task.Picture > 0 {
			b.PullPicture(fmt.Sprintf("pics\\%d.png", task.Picture), chatID, 0)
		}
		b.PullPoll(questionID, task.Problem, chatID, 0, false, task.Variants...)
	case "getstats":
		if b.oprosRun < 10 {
			keyboardDefault := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(`Да`),
				tgbotapi.NewKeyboardButton(`Нет`))
			b.regime = true
			b.PullText("Недостаточно информации для оценки уровня знаний. Продолжить опрос?", chatID, 0, keyboardDefault)
		} else {
			b.getResult(chatID)
		}
	default:
		b.handleUnknown()
	}
	return nil
}

func (b *Bot) handleSticker(message *tgbotapi.Message, chatID int) error {
	b.PullSticker(message.Sticker.FileID, chatID, false, 0)
	return nil
}

func (b *Bot) handleRegimeNo(message *tgbotapi.Message, chatID int) {
	b.regime = false
	b.oprosRun = -1
	b.stat = 0
	b.PullText("Тестирование остановлено, результат не сохранен. ", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
}

func (b *Bot) handleRegimeYes(message *tgbotapi.Message, chatID int) {
	b.regime = false
	b.PullText("Тестирование продолжается... ", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
}

func (b *Bot) handleText(message *tgbotapi.Message, chatID int) {
	b.PullText(message.Text, chatID, message.MessageID)
}

/* На удаление
func (b *Bot) handlePoll(message *tgbotapi.Poll, chatID int) {
	b.oprosRun++
	b.stat++
	a := task.GenerateRandomQuestion(b.oprosRun/3, b.oprosRun%3)
	b.PullPoll(a.Problem, chatID, 0, false, a.Variants...)
}*/

func (b *Bot) handlePollAnswer(ans *tgbotapi.PollAnswer, chatID int) {
	if b.oprosRun != -1 {
		// Получить номер вопроса из базы
		// Проверить время
		// Проверить принадлежность опроса к пользователю
		if len(ans.OptionIDs) == 1 {
			if task.CheckQuestion(b.DB, 2, ans.OptionIDs[0]) {
				b.stat++
			}
		}
		b.iterateTest(chatID)
	}
}

func (b *Bot) iterateTest(chatID int) {
	if b.oprosRun < 30 {
		b.oprosRun++
		// Получить из бд нужный вопрос
		task := task.GetQuestion(b.DB, b.oprosRun+1)
		if task.Picture > 0 {
			b.PullPicture(fmt.Sprintf("pics\\%d.png", task.Picture), chatID, 0)
		}
		b.PullPoll(task.Number, task.Problem, chatID, 0, false, task.Variants...)
	} else {
		b.getResult(chatID)
	}
}

func (b *Bot) getResult(chatID int) {
	if b.oprosRun > 0 {
		b.PullText(fmt.Sprintf("Статистика: %v%%\nОтветов: %v\nПравильных: %v", b.stat*100/b.oprosRun, b.oprosRun, b.stat), chatID, 0)
	} else {
		b.PullText("Статистика: Нет информации о пройденных тестах.", chatID, 0)
	}
	b.oprosRun = -1
	b.stat = 0
}

func (b *Bot) handleUnknown() error {
	return errors.New("unknown item was received")
}

// КАКАЯ-ТО ШЛЯПА НЕ ТРОГАТЬ (вызывает колбеки)
func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery, chatID int) error {
	switch callback.Data {
	case "testContinueNo":
		b.PullText("Советую попробовать еще раз...", chatID, 0)
	case "testContinueYes":
		// Получить из бд нужный вопрос
		a := task.GetQuestion(b.DB, b.oprosRun)
		//a = task.GenerateRandomQuestion(b.oprosRun/3, b.oprosRun%3)
		log.Println(a)
		b.PullPoll(a.Number, a.Problem, chatID, 0, false, a.Variants...)
	default:
		return b.handleUnknown()
	}
	return nil
}
