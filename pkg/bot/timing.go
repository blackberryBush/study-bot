package bot

import (
	"fmt"
	"github.com/spf13/viper"
	"study-bot/pkg/users"
	"time"
)

func getTime() time.Duration {
	viper.SetConfigName("options")
	viper.AddConfigPath(".")
	time1 := 10
	if err := viper.ReadInConfig(); err == nil {
		time1 = viper.GetInt("options.time_min")
	}
	return time.Duration(time1)
}

func (b *Bot) TimerRun(user *users.User) {
	duration := time.Minute * getTime()
	chatID := user.ID
	b.PullText(fmt.Sprintf("Внимание! На тестирование отведено %v минут", duration.Minutes()), chatID, 0)
	b.timers[chatID] = time.NewTimer(duration)
	<-b.timers[chatID].C
	user.PollRun--
	b.PullText("Тестирование остановлено в связи с истечением времени.", chatID, 0)
	b.getResult(user)
	users.UpdateUser(b.DB, *user)
	delete(b.timers, chatID)
}

func (b *Bot) TimerStop(user *users.User) {
	if b.timers[user.ID] != nil {
		b.timers[user.ID].Stop()
		delete(b.timers, user.ID)
	}
}
