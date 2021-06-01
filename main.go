package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Result holds the information for a vaccine appointment
type Result struct {
	VaccineName string
	Amount      int64
	Message     string
}

// Fetcher is the type to allow fetching information for an appointment
type Fetcher interface {
	Fetch() ([]*Result, error)
	ResultSendLastAt() time.Time
	ResultSentNow()
}

func fetchAllAppointment(fetchers []Fetcher, bot *Telegram) {
	done := make(chan bool)
	errChan := make(chan error)

	for _, fetcher := range fetchers {
		fetcher := fetcher
		go func() {
			fmt.Println("Starting fetch")
			res, err := fetcher.Fetch()
			if err != nil {
				errChan <- err
				return
			}
			fmt.Printf("Received %d result\n", len(res))
			if len(res) > 0 && fetcher.ResultSendLastAt().Before(time.Now().Add(-1*time.Minute)) {
				fetcher.ResultSentNow()
				for _, r := range res {

					bot.SendMessageToAllUser(r.Message)
					if err != nil {
						errChan <- err
						return
					}
				}
				fmt.Println("messages sent on telegram")
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
	db, err := NewDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	bot, err := tgbotapi.NewBotAPI(viper.GetString("telegram-token"))
	if err != nil {
		fmt.Println(err)
		return
	}
	telegram := NewBot(bot, db)

	sources := []Fetcher{
		&PuntoMedico{},
		&VaccineCenter{},
	}

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		err := telegram.HandleNewUsers()
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		for range time.Tick(5 * time.Second) {
			go fetchAllAppointment(sources, telegram)
		}
	}()

	wg.Wait()
}
