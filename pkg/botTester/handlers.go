package botTester

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"sort"
	"strconv"
	"strings"
	"study-bot/pkg/botBasic"
	"study-bot/pkg/databases"
	"study-bot/pkg/log"
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
	messageNotAdmin
)

func (b *TesterBot) getUpdateType(update *tgbotapi.Update) (int, int64, databases.User, error) {
	if update.PollAnswer != nil {
		chatID := update.PollAnswer.User.ID
		user, err := databases.GetUser(b.DB, chatID)
		return updatePollAnswer, chatID, user, err
	}
	if update.Message != nil {
		chatID := update.Message.From.ID
		user, err := databases.GetUser(b.DB, chatID)
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
		chatID := update.CallbackQuery.From.ID
		user, err := databases.GetUser(b.DB, chatID)
		return callbackQuery, chatID, user, err
	}
	return messageUndefined, 0, databases.User{}, nil
}

func (b *TesterBot) HandleUpdate(update *tgbotapi.Update) {
	updateType, chatID, currentUser, err := b.getUpdateType(update)
	if err != nil {
		currentUser = *databases.NewUser(chatID)
		databases.InsertUser(b.DB, currentUser)
	}
	log.PrintReceive(update, updateType, chatID)
	if currentUser.Regime == 1 && updateType != messageRegimeNo && updateType != messageRegimeYes {
		if updateType != updatePollAnswer {
			currentUser.Regime = 0
			currentUser.PollRun = -1
			currentUser.Corrects = 0
			b.PullText("Счёл это за отказ...", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
			databases.UpdateUser(b.DB, currentUser)
			return
		} else {
			currentUser.Regime = 0
			b.PullText("Счёл это за продолжение...", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
		}
	}
	switch updateType {
	case updatePollAnswer:
		b.handlePollAnswer(update.PollAnswer, &currentUser)
	case messageCommand:
		b.handleCommand(update.Message, &currentUser)
	case messageSticker:
		b.handleSticker(update.Message, &currentUser)
	case messageText:
		b.handleText(update.Message, &currentUser)
	case messageRegimeNo:
		b.handleRegimeNo(&currentUser)
	case messageRegimeYes:
		b.handleRegimeYes(&currentUser)
	case callbackQuery:
		b.handleCallbackQuery(update.CallbackQuery, &currentUser)
	default:
		b.handleUnknown()
	}
	databases.UpdateUser(b.DB, currentUser)
}

func (b *TesterBot) handleCommand(message *tgbotapi.Message, user *databases.User) {
	chatID := user.ID
	switch message.Command() {
	case "start":
		databases.ClearUser(b.DB, chatID)
		b.PullText("Основные команды бота:\n"+
			"/test - запуск тестирования\n"+
			"/getstats - получить результаты тестирования\n"+
			"/study - открыть учебник\n\n"+
			"Основные правила тестирования:\n"+
			"- тестирование ограничено по времени\n"+
			"- вариант ответа всегда один\n"+
			"- между вопросами нельзя переключаться и изменять уже отправленный ответ\n"+
			"- при каждом новом запуске /test, старые результаты стираются", chatID, 0)
	case "test":
		databases.ClearUser(b.DB, chatID)
		go b.TimerRun(user)
		user.PollRun = 0
		user.Corrects = 0
		for i := range user.Worst {
			user.Worst[i] = 0
			user.Chapters[i] = 0
		}
		b.iterateTest(user)
	case "getstats":
		if user.PollRun == -1 {
			databases.GetLastStats(b.DB, user)
			b.getResult(user)
		} else if user.PollRun < len(b.Chapters) {
			keyboardDefault := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(`Да`),
				tgbotapi.NewKeyboardButton(`Нет`))
			user.Regime = 1
			b.PullText("Недостаточно информации для оценки уровня знаний. Продолжить опрос?", chatID, 0, keyboardDefault)
		} else {
			user.PollRun--
			b.getResult(user)
		}
	case "study":
		b.showTextbook(user)
	default:
		b.handleUnknown()
	}
}

func (b *TesterBot) handleSticker(message *tgbotapi.Message, user *databases.User) {
	chatID := user.ID
	b.PullSticker(message.Sticker.FileID, chatID, false, 0)
}

func (b *TesterBot) handleRegimeNo(user *databases.User) {
	chatID := user.ID
	user.Regime = 0
	user.PollRun = -1
	databases.ClearUser(b.DB, chatID)
	b.TimerStop(user)
	b.PullText("Тестирование остановлено, результат не сохранен. ", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
}

func (b *TesterBot) handleRegimeYes(user *databases.User) {
	chatID := user.ID
	user.Regime = 0
	b.PullText("Тестирование продолжается... ", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
}

func (b *TesterBot) handleText(message *tgbotapi.Message, user *databases.User) {
	chatID := user.ID
	b.PullText(message.Text, chatID, message.MessageID)
}

func (b *TesterBot) handlePollAnswer(ans *tgbotapi.PollAnswer, user *databases.User) {
	chatID := user.ID
	if user.PollRun != -1 {
		checkTask, checkChapter, checkCorrect := databases.GetTask(b.DB, ans.PollID, chatID)
		if checkTask <= 0 {
			return
		}
		if len(ans.OptionIDs) == 1 {
			databases.UpdateAnswer(b.DB, chatID, ans.PollID, ans.OptionIDs[0]+1)
			user.Chapters[checkChapter]++
			if checkCorrect == ans.OptionIDs[0]+1 {
				user.Corrects++
			} else {
				user.Worst[checkChapter]++
			}
		}
		b.iterateTest(user)
	}
}

func (b *TesterBot) iterateTest(user *databases.User) {
	chatID := user.ID
	if user.PollRun < len(b.Chapters)*b.iterations {
		user.PollRun++
		questionID := databases.GetRandomQuestionNumber(b.DB, user.PollRun, b.Chapters, user.ID)
		currentTask, err := databases.GetQuestion(b.DB, questionID)
		if err != nil {
			b.PullText("Произошла ошибка при тестировании...", chatID, 0)
			b.getResult(user)
			return
		}
		currentTask.MixQuestion()
		if currentTask.Picture > 0 {
			b.PullPicture(fmt.Sprintf("pics/%d.png", currentTask.Picture), chatID, 0)
		}
		b.PullPoll(currentTask.Number, currentTask.Problem, chatID, 0, false, currentTask.Correct, currentTask.Variants...)
	} else {
		b.getResult(user)
	}
}

func OutputSortedByKey(user *databases.User) string {
	keys := make([]int, 0, len(user.Chapters))
	for k := range user.Chapters {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	s := "\n\nДоля неправильных ответов по главам:"
	for _, k := range keys {
		if user.Chapters[k] != 0 {
			s += fmt.Sprintf("\nГлава %d: %d%%", k, user.Worst[k]*100/user.Chapters[k])
		}
	}
	return s
}

func (b *TesterBot) getResult(user *databases.User) {
	b.TimerStop(user)
	chatID := user.ID
	if user.PollRun > 0 {
		s := OutputSortedByKey(user)
		b.PullText(fmt.Sprintf("Пользователь: %v\nСтатистика: %v%%\nОтветов: %v\nПравильных: %v%s", user.ID,
			user.Corrects*100/user.PollRun, user.PollRun, user.Corrects, s), chatID, 0)
	} else {
		b.PullText("Статистика: Нет информации о пройденных тестах.", chatID, 0)
	}
	user.PollRun = -1
}

func (b *TesterBot) handleUnknown() {
	return
}

func (b *TesterBot) showTextbook(user *databases.User) {
	chatID := user.ID
	keyboard := getKeyboardChapters()
	msg := tgbotapi.NewMessage(chatID, "Выберите главу: ")
	msg.ReplyMarkup = keyboard
	go b.Pull(chatID, *botBasic.NewChattable(msg))
}

func getFiles(directory string, isDir bool) []string {
	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Println(err)
	}
	arr := make([]string, 0)
	for _, file := range files {
		if file.IsDir() == isDir {
			arr = append(arr, file.Name())
		}
	}
	return arr
}

func getKeyboardChapters() tgbotapi.InlineKeyboardMarkup {
	files := getFiles("textbook/", true)
	k := len(files)
	buttons := make([][]tgbotapi.InlineKeyboardButton, k)
	for i := 0; i < k; i++ {
		buttons[i] = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
			strings.Replace(strings.Replace(files[i], "_", " ", -1), " 0", " ", -1),
			"cH"+strconv.Itoa(i)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func getKeyboardParagraphs(dirID int) tgbotapi.InlineKeyboardMarkup {
	files1 := getFiles("textbook/", true)
	files := getFiles("textbook/"+files1[dirID]+"/", false)
	k := len(files)
	if k == 0 {
		return tgbotapi.InlineKeyboardMarkup{}
	}
	buttons := make([][]tgbotapi.InlineKeyboardButton, k)
	for i := 0; i < k; i++ {
		p, _, found := strings.Cut(strings.Replace(files[i], "_", " ", -1), ".pdf")
		if !found {
			p, _, _ = strings.Cut(strings.Replace(files[i], "_", " ", -1), ".doc")
		}
		fmt.Println(p)
		buttons[i] = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
			p, "cF"+strconv.Itoa(dirID)+"ph"+strconv.Itoa(i)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func getFilename(chapter int, paragraph int) (string, string) {
	files1 := getFiles("textbook/", true)
	files2 := getFiles("textbook/"+files1[chapter], false)
	return "textbook/" + files1[chapter] + "/" + files2[paragraph], files2[paragraph]
}

func (b *TesterBot) showTextbookFiles(user *databases.User, chapterID int) {
	chatID := user.ID
	keyboard := getKeyboardParagraphs(chapterID)
	msg := tgbotapi.NewMessage(chatID, "Выберите пункт главы: ")
	msg.ReplyMarkup = keyboard
	go b.Pull(chatID, *botBasic.NewChattable(msg))
}

func (b *TesterBot) handleCallbackQuery(callback *tgbotapi.CallbackQuery, user *databases.User) {
	chatID := user.ID
	switch {
	case strings.HasPrefix(callback.Data, "cF"):
		_, k, _ := strings.Cut(callback.Data, "cF")
		k1, k2, _ := strings.Cut(k, "ph")
		intK1, _ := strconv.Atoi(k1)
		intK2, _ := strconv.Atoi(k2)
		s, s1 := getFilename(intK1, intK2)
		b.PullFile(s, chatID, 0, strings.Replace(s1, "_", " ", -1))
	case strings.HasPrefix(callback.Data, "cH"):
		_, k, _ := strings.Cut(callback.Data, "cH")
		v, _ := strconv.Atoi(k)
		b.showTextbookFiles(user, v)
	}
}
