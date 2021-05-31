package main

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Result struct {
	VaccineName string
	Amount      int64
	Message     *string
}

type Fetcher interface {
	Fetch() ([]*Result, error)
	ResultSendLastAt() time.Time
	ResultSentNow()
	FormatMessage(result *Result) (string, error)
}

func fetchAllAppointment(fetchers []Fetcher, bot *Telegram) {
	done := make(chan bool)
	errChan := make(chan error)

	for _, fetcher := range fetchers {
		fetcher := fetcher
		go func() {
			fmt.Println("Starting fetch\n")
			res, err := fetcher.Fetch()
			if err != nil {
				errChan <- err
				return
			}
			fmt.Printf("Received %d result\n", len(res))
			if len(res) > 0 && fetcher.ResultSendLastAt().Before(time.Now().Add(-20*time.Second)) {
				fetcher.ResultSentNow()
				for _, r := range res {
					if r.Message == nil {
						message, err := fetcher.FormatMessage(r)
						if err != nil {
							errChan <- err
							return
						}
						bot.SendMessage(message)
					} else {
						bot.SendMessage(*r.Message)
					}
				}
				fmt.Printf("messages sent on telegram\n")
			}
			done <- true
		}()
	}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-done:
			fmt.Println("fetch done")
		case <-timeout:
			return
		case err := <-errChan:
			fmt.Printf("%v\n", err)

		}
	}
}

func init() {
	viper.SetConfigName(".config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(viper.GetString("telegram-token"))
	if err != nil {
		return
	}
	sources := []Fetcher{
		&PuntoMedico{},
		&VaccineCenter{},
	}

	telegram := NewBot(bot, viper.GetInt64("telegram-channels"))
	for range time.Tick(5 * time.Second) {
		go fetchAllAppointment(sources, telegram)
	}
}
