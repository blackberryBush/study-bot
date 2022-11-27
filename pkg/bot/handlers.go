package bot

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"study-bot/pkg/task"
	"study-bot/pkg/users"
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

func (b *Bot) getUpdateType(update *tgbotapi.Update) (int, int, users.User, error) {
	// Закрыть дыру с неизвестными видами содержимого!!!!!

	/*if update.Poll != nil {
		return updatePoll, 0
	}*/
	if update.PollAnswer != nil {
		chatID := int(update.PollAnswer.User.ID)
		user, err := users.GetUser(b.DB, chatID)
		return updatePollAnswer, chatID, user, err
	}
	if update.Message != nil {
		chatID := int(update.Message.From.ID)
		user, err := users.GetUser(b.DB, chatID)
		if update.Message.IsCommand() {
			return messageCommand, chatID, user, err
		}
		if user.Regime == 1 && update.Message.Text == "Нет" {
			return messageRegimeNo, chatID, user, err
		}
		if user.Regime == 1 && update.Message.Text == "Да" {
			return messageRegimeYes, chatID, user, err
		}
		if update.Message.Text == "" && update.Message.Sticker != nil {
			return messageSticker, chatID, user, err
		}
		if update.Message.Text != "" {
			return messageText, chatID, user, err
		}
		return messageUndefined, chatID, user, err
	}
	if update.CallbackQuery != nil {
		chatID := int(update.CallbackQuery.From.ID)
		user, err := users.GetUser(b.DB, chatID)
		return callbackQuery, chatID, user, err
	}
	return messageUndefined, 0, users.User{}, nil
}

func (b *Bot) handleUpdate(update *tgbotapi.Update) {
	updateType, chatID, currentUser, err := b.getUpdateType(update)
	if err != nil {
		currentUser = *users.NewUser(chatID)
		users.InsertUser(b.DB, currentUser)
	}
	if currentUser.Regime == 1 && updateType != messageRegimeNo && updateType != messageRegimeYes {
		if updateType != updatePollAnswer {
			currentUser.Regime = 0
			currentUser.PollRun = -1
			currentUser.Corrects = 0
			b.PullText("Счёл это за отказ...", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
			return
		} else {
			currentUser.Regime = 0
			b.PullText("Счёл это за продолжение...", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
		}
	}
	switch updateType {
	/*case updatePoll:
	b.handlePoll(update.Poll, chatID)*/
	case updatePollAnswer:
		b.handlePollAnswer(update.PollAnswer, &currentUser)
	case messageCommand:
		b.handleCommand(update.Message, &currentUser)
	case messageSticker:
		b.handleSticker(update.Message, &currentUser)
	case messageText:
		b.handleText(update.Message, &currentUser)
	case messageRegimeNo:
		b.handleRegimeNo(update.Message, &currentUser)
	case messageRegimeYes:
		b.handleRegimeYes(update.Message, &currentUser)
	case callbackQuery:
	default:
		b.handleUnknown()
	}
	users.UpdateUser(b.DB, currentUser)
}

func (b *Bot) handleCommand(message *tgbotapi.Message, user *users.User) error {
	chatID := user.ID
	switch message.Command() {
	case "start":
		user.PollRun = -1
		user.Corrects = 0
	case "test":
		user.PollRun = 0
		questionID := user.PollRun + 1 // номер вопроса
		currentTask := task.GetQuestion(b.DB, questionID)
		if currentTask.Picture > 0 {
			b.PullPicture(fmt.Sprintf("pics\\%d.png", currentTask.Picture), chatID, 0)
		}
		b.PullPoll(questionID, currentTask.Problem, chatID, 0, false, currentTask.Variants...)
	case "getstats":
		if user.PollRun < 10 {
			keyboardDefault := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(`Да`),
				tgbotapi.NewKeyboardButton(`Нет`))
			user.Regime = 1
			b.PullText("Недостаточно информации для оценки уровня знаний. Продолжить опрос?", chatID, 0, keyboardDefault)
		} else {
			b.getResult(user)
		}
	default:
		b.handleUnknown()
	}
	return nil
}

func (b *Bot) handleSticker(message *tgbotapi.Message, user *users.User) error {
	chatID := user.ID
	b.PullSticker(message.Sticker.FileID, chatID, false, 0)
	return nil
}

func (b *Bot) handleRegimeNo(message *tgbotapi.Message, user *users.User) {
	chatID := user.ID
	user.Regime = 1
	user.PollRun = -1
	user.Corrects = 0
	b.PullText("Тестирование остановлено, результат не сохранен. ", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
}

func (b *Bot) handleRegimeYes(message *tgbotapi.Message, user *users.User) {
	chatID := user.ID
	user.Regime = 1
	b.PullText("Тестирование продолжается... ", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
}

func (b *Bot) handleText(message *tgbotapi.Message, user *users.User) {
	chatID := user.ID
	b.PullText(message.Text, chatID, message.MessageID)
}

/* На удаление
func (b *Bot) handlePoll(message *tgbotapi.Poll, chatID int) {
	b.oprosRun++
	b.stat++
	a := task.GenerateRandomQuestion(b.oprosRun/3, b.oprosRun%3)
	b.PullPoll(a.Problem, chatID, 0, false, a.Variants...)
}*/

func (b *Bot) handlePollAnswer(ans *tgbotapi.PollAnswer, user *users.User) error {
	chatID := user.ID
	if user.PollRun != -1 {
		// Проверить время
		check := users.GetTask(b.DB, ans.PollID, chatID)
		if check <= 0 {
			return fmt.Errorf("check error")
		}
		if len(ans.OptionIDs) == 1 {
			if task.CheckQuestion(b.DB, check, ans.OptionIDs[0]+1) {
				user.Corrects++
			}
		}
		b.iterateTest(user)
	}
	return nil
}

func (b *Bot) iterateTest(user *users.User) {
	chatID := user.ID
	if user.PollRun < 30 {
		user.PollRun++
		// Получить из бд нужный вопрос
		currentTask := task.GetQuestion(b.DB, user.PollRun+1)
		if currentTask.Picture > 0 {
			b.PullPicture(fmt.Sprintf("pics\\%d.png", currentTask.Picture), chatID, 0)
		}
		b.PullPoll(currentTask.Number, currentTask.Problem, chatID, 0, false, currentTask.Variants...)
	} else {
		b.getResult(user)
	}
}

func (b *Bot) getResult(user *users.User) {
	chatID := user.ID
	if user.PollRun > 0 {
		b.PullText(fmt.Sprintf("Статистика: %v%%\nОтветов: %v\nПравильных: %v", user.Corrects*100/user.PollRun, user.PollRun, user.Corrects), chatID, 0)
	} else {
		b.PullText("Статистика: Нет информации о пройденных тестах.", chatID, 0)
	}
	user.PollRun = -1
	user.Corrects = 0
	users.ClearUser(b.DB, chatID)
}

func (b *Bot) handleUnknown() error {
	return errors.New("unknown item was received")
}

// КАКАЯ-ТО ШЛЯПА НЕ ТРОГАТЬ (вызывает колбеки)
func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery, user users.User) error {
	chatID := user.ID
	switch callback.Data {
	case "testContinueNo":
		b.PullText("Советую попробовать еще раз...", chatID, 0)
	case "testContinueYes":
		// Получить из бд нужный вопрос
		a := task.GetQuestion(b.DB, user.PollRun)
		//a = task.GenerateRandomQuestion(b.oprosRun/3, b.oprosRun%3)
		log.Println(a)
		b.PullPoll(a.Number, a.Problem, chatID, 0, false, a.Variants...)
	default:
		return b.handleUnknown()
	}
	return nil
}
