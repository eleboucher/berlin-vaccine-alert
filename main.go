package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/eleboucher/covid/models/chat"
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
	ShouldSendResult(result []*Result) bool
	ResultSentNow(result []*Result)
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
			if len(res) > 0 && fetcher.ShouldSendResult(res) {
				fetcher.ResultSentNow(res)
				for _, r := range res {

					bot.SendMessageToAllUser(r)
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
	chatModel := chat.NewModel(db)
	telegram := NewBot(bot, chatModel)

	sources := []Fetcher{
		// &PuntoMedico{},
		// &VaccineCenter{},
		// &MedicoLeopoldPlatz{},
		&Doctolib{
			URL:           "https://www.doctolib.de/praxis/brandenburg-an-der-havel/corona-schutzimpfung-gzb",
			VaccineName:   JohnsonAndJohnson,
			PraticeID:     "186461",
			AgendaID:      "472530",
			VisitMotiveID: "2877045",
			Detail:        "(for 40+)",
		},
		&Doctolib{
			URL:           "https://www.doctolib.de/praxis/brandenburg-an-der-havel/corona-schutzimpfung-gzb",
			VaccineName:   AstraZeneca,
			PraticeID:     "186461",
			AgendaID:      "472530",
			VisitMotiveID: "2741487",
			Detail:        "(for 40+)",
		},
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
