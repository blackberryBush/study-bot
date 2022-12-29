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
	switch message.CommandWithAt() {
	case "start":
		databases.ClearUser(b.DB, chatID)
		b.PullText("Основные команды бота:\n"+
			"/test - запуск тестирования\n"+
			"/getstats - получить результаты тестирования\n"+
			"/study - открыть письменные материалы\n\n"+
			"Основные правила тестирования:\n"+
			"- тестирование ограничено по времени\n"+
			"- вариант ответа всегда один\n"+
			"- между вопросами нельзя переключаться и изменять уже отправленный ответ\n"+
			"- при каждом новом запуске /test, старые результаты стираются", chatID, 0)
	case "test":
		databases.ClearUser(b.DB, chatID)
		args := message.CommandArguments()
		if args == "" {
			b.PullText("Введите: /test [Группа,ФИО] (без пробелов)\nПример: /test БКС1902,ИвановАВ", chatID, message.MessageID)
			return
		}
		go b.TimerRun(user)
		user.PollRun = 0
		user.Corrects = 0
		user.Username = args
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
		b.PullText(fmt.Sprintf("ID: %v\nПользователь: %v\nСтатистика: %v%%\nОтветов: %v\nПравильных: %v%s",
			user.ID, user.Username, user.Corrects*100/user.PollRun, user.PollRun, user.Corrects, s), chatID, 0)
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
	keyboard, err := getKeyboardElems("textbook/", "l")
	if err != nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Выберите:")
	msg.ReplyMarkup = keyboard
	go b.Pull(chatID, *botBasic.NewChattable(msg))
}

func getFiles(directory string, isDirectory bool) ([]string, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Println(err)
	}
	arr := make([]string, 0)
	for _, file := range files {
		if file.IsDir() == isDirectory {
			arr = append(arr, file.Name())
		}
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("no files error")
	}
	return arr, nil
}

func getKeyboardElems(directory string, directoryShort string) (tgbotapi.InlineKeyboardMarkup, error) {
	files, err := getFiles(directory, true)
	dirShort := ""
	if err != nil {
		files, err = getFiles(directory, false)
		dirShort = "f" + directoryShort
		if err != nil {
			return tgbotapi.InlineKeyboardMarkup{}, err
		}
	} else {
		dirShort = directoryShort
	}
	k := len(files)
	buttons := make([][]tgbotapi.InlineKeyboardButton, k)
	for i := 0; i < k; i++ {
		p, _, found := strings.Cut(strings.Replace(strings.Replace(files[i], "_", " ", -1),
			" 0", " ", -1), ".pdf")
		if !found {
			p, _, _ = strings.Cut(strings.Replace(strings.Replace(files[i], "_", " ", -1),
				" 0", " ", -1), ".doc")
		}
		buttons[i] = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
			p, dirShort+"no"+strconv.Itoa(i)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...), nil
}

func getFilename(nums []int) (string, string) {
	name1, name2 := "textbook/", ""
	files, err := getFiles("textbook/", true)
	if err != nil && len(nums) == 1 {
		mFiles, _ := getFiles("textbook/", false)
		return "textbook/" + mFiles[nums[0]], mFiles[nums[0]]
	}
	if len(nums) == 1 {
		return "textbook/" + files[nums[0]] + "/", ""
	}
	for i := 1; i < len(nums); i++ {
		name1 += files[nums[i-1]] + "/"
		files, _ = getFiles(name1, true)
		if len(files) == 0 {
			files, _ = getFiles(name1, false)
		}
		name2 = files[nums[i]]
	}
	return name1 + name2, name2
}

func (b *TesterBot) showTextbookFiles(user *databases.User, chapterIDs []int) {
	chatID := user.ID
	u, _ := getFilename(chapterIDs)
	ds := "l"
	for _, v := range chapterIDs {
		ds += "no" + strconv.Itoa(v)
	}
	keyboard, _ := getKeyboardElems(u, ds)
	msg := tgbotapi.NewMessage(chatID, "Выберите: ")
	msg.ReplyMarkup = keyboard
	go b.Pull(chatID, *botBasic.NewChattable(msg))
}

func getChapterNums(k string) []int {
	cloneK := strings.Clone(k)
	nums := make([]int, 0)
	for found := true; ; {
		n := ""
		n, k, found = strings.Cut(k, "no")
		if !found {
			break
		} else {
			cloneK = k
		}
		intN, err := strconv.Atoi(n)
		if err == nil {
			nums = append(nums, intN)
		}
	}
	k = cloneK
	intN, err := strconv.Atoi(k)
	if err == nil {
		nums = append(nums, intN)
	}
	return nums
}

func (b *TesterBot) handleCallbackQuery(callback *tgbotapi.CallbackQuery, user *databases.User) {
	chatID := user.ID
	switch {
	case strings.HasPrefix(callback.Data, "fl"):
		_, k, _ := strings.Cut(callback.Data, "flno")
		nums := getChapterNums(k)
		s, s1 := getFilename(nums)
		b.PullFile(s, chatID, 0, strings.Replace(s1, "_", " ", -1))
	case strings.HasPrefix(callback.Data, "l"):
		_, k, _ := strings.Cut(callback.Data, "lno")
		nums := getChapterNums(k)
		b.showTextbookFiles(user, nums)
	}
}
