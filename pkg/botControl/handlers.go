package botControl

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"strconv"
	"study-bot/pkg/botTester"
	"study-bot/pkg/databases"
	"study-bot/pkg/log"
)

// вызов гайда для прочих обновлений

const (
	messageUndefined = -1
	messageNotAdmin  = iota
	messageCommand
)

func checkID(userID int64) bool {
	viper.SetConfigName("options")
	viper.AddConfigPath(".")
	var admins []string
	if err := viper.ReadInConfig(); err == nil {
		admins = viper.GetStringSlice("admins")
	}
	for _, v := range admins {
		if strconv.FormatInt(userID, 10) == v {
			return true
		}
	}
	return false
}

func (b *ControlBot) getUpdateType(update *tgbotapi.Update) (int, int64, error) {
	if update.Message != nil {
		chatID := update.Message.From.ID
		if !checkID(chatID) {
			return messageNotAdmin, chatID, nil
		}
		if update.Message.IsCommand() {
			return messageCommand, chatID, nil
		}
		return messageUndefined, chatID, nil
	}
	return messageUndefined, 0, nil
}

func (b *ControlBot) HandleUpdate(update *tgbotapi.Update) {
	updateType, chatID, _ := b.getUpdateType(update)
	log.PrintReceive(update, updateType, chatID)
	switch updateType {
	case messageNotAdmin:
		b.handleMessageNotAdmin(update.Message, chatID)
	case messageCommand:
		b.handleCommand(update.Message, chatID)
	default:
		b.handleUnknown()
	}
}

func (b *ControlBot) handleMessageNotAdmin(message *tgbotapi.Message, chatID int64) {
	b.PullText("Вы не админ! (Ваш ID "+strconv.FormatInt(chatID, 10)+")", chatID, message.MessageID)
}

func (b *ControlBot) handleCommand(message *tgbotapi.Message, chatID int64) {
	switch message.CommandWithAt() {
	case "start":
		b.PullText("Го чё-то вступительное сюда напишем...", chatID, message.MessageID)
	case "user":
		user := message.CommandArguments()
		if user == "" {
			b.PullText("Не введён ID пользователя", chatID, message.MessageID)
		}
		if ID, err := strconv.ParseInt(user, 10, 0); err == nil {
			userN, err := databases.GetUser(b.DB, ID)
			if err == nil {
				databases.GetLastStats(b.DB, &userN)
				b.getResult(&userN, chatID)
			} else {
				b.PullText("Пользователь не найден", chatID, message.MessageID)
			}
		}
	case "clear":
		databases.ClearNotes(b.DB)
		databases.ClearUsers(b.DB)
		b.PullText("Базы пользователей очищены", chatID, message.MessageID)
	case "clearuser":
		user := message.CommandArguments()
		if user == "" {
			b.PullText("Не введён ID пользователя", chatID, message.MessageID)
		}
		if ID, err := strconv.ParseInt(user, 10, 0); err == nil {
			_, err := databases.GetUser(b.DB, ID)
			if err == nil {
				databases.ClearUser(b.DB, ID)
				b.PullText("Тесты пользователя очищены", chatID, message.MessageID)
			} else {
				b.PullText("Пользователь не найден", chatID, message.MessageID)
			}
		}
	case "usertasks":
		user := message.CommandArguments()
		if user == "" {
			b.PullText("Не введён ID пользователя", chatID, message.MessageID)
		}
		if ID, err := strconv.ParseInt(user, 10, 0); err == nil {
			_, err := databases.GetUser(b.DB, ID)
			if err == nil {
				s := databases.GetUserNotes(b.DB, ID)
				b.PullText("Тесты пользователя:\n"+s, chatID, message.MessageID)
			} else {
				b.PullText("Пользователь не найден", chatID, message.MessageID)
			}
		}
	case "task":
		task := message.CommandArguments()
		if task == "" {
			b.PullText("Не введён ID вопроса", chatID, message.MessageID)
		}
		if questionID, err := strconv.ParseInt(task, 10, 0); err == nil {
			currentTask, err := databases.GetQuestion(b.DB, int(questionID))
			if err != nil {
				b.PullText("Произошла ошибка при запросе вопроса...", chatID, 0)
			}
			currentTask.MixQuestion()
			if currentTask.Picture > 0 {
				b.PullPicture(fmt.Sprintf("pics/%d.png", currentTask.Picture), chatID, 0)
			}
			b.PullPoll(currentTask.Number, currentTask.Problem, chatID, 0, false, currentTask.Correct, currentTask.Variants...)
		}
	case "getdb":
		s := databases.GetAllNotes(b.DB)
		b.PullFileBytes([]byte(s), chatID, message.MessageID, "notes.txt")
		s = databases.GetAllTasks(b.DB)
		b.PullFileBytes([]byte(s), chatID, message.MessageID, "tasks.txt")
		s = databases.GetAllUsers(b.DB)
		b.PullFileBytes([]byte(s), chatID, message.MessageID, "users.txt")
	case "update":
		b.PullText("Инструкция: текст", chatID, message.MessageID)
	default:
		b.handleUnknown()
	}
}

func (b *ControlBot) getResult(user *databases.User, chatID int64) {
	if user.PollRun > 0 {
		s := botTester.OutputSortedByKey(user)
		b.PullText(fmt.Sprintf("Пользователь: %v\nСтатистика: %v%%\nОтветов: %v\nПравильных: %v%s",
			user.ID, user.Corrects*100/user.PollRun, user.PollRun, user.Corrects, s), chatID, 0)
	} else {
		b.PullText("Статистика: Нет информации о пройденных тестах.", chatID, 0)
	}
}

func (b *ControlBot) handleUnknown() {
	// =)
}
