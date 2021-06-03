package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/eleboucher/covid/models/chat"
	"github.com/eleboucher/covid/sources"
	"github.com/eleboucher/covid/vaccines"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Fetcher is the type to allow fetching information for an appointment
type Fetcher interface {
	Name() string
	Fetch() ([]*vaccines.Result, error)
	ShouldSendResult(result []*vaccines.Result) bool
	ResultSentNow(result []*vaccines.Result)
}

var rootCmd = &cobra.Command{
	Use: "berlin-vaccine-alert <command>",
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func fetchAllAppointment(fetchers []Fetcher, bot *Telegram) {
	done := make(chan bool)
	errChan := make(chan error)

	for _, fetcher := range fetchers {
		fetcher := fetcher
		go func() {
			log.Infof("%s: Starting fetch", fetcher.Name())
			res, err := fetcher.Fetch()
			if err != nil {
				errChan <- err
				return
			}
			log.Infof("%s: Received %d result", fetcher.Name(), len(res))
			if len(res) > 0 && fetcher.ShouldSendResult(res) {
				fetcher.ResultSentNow(res)
				for _, r := range res {

					bot.SendMessageToAllUser(r)
					if err != nil {
						errChan <- err
						return
					}
				}
				log.Infof("%s: messages sent on telegram", fetcher.Name())
			}
			done <- true
		}()
	}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-done:
			continue
		case <-timeout:
			return
		case err := <-errChan:
			log.Errorf("%v\n", err)

		}
	}
}

func getAllDoctolibSources() ([]Fetcher, error) {
	var doctolibConfigs []*DoctolibConfig
	var ret []Fetcher
	err := viper.UnmarshalKey("doctolib", &doctolibConfigs)
	if err != nil {
		return nil, err
	}
	for _, doctolibConfig := range doctolibConfigs {
		tmp := sources.Doctolib{
			URL:           doctolibConfig.URL,
			PracticeID:    doctolibConfig.PracticeID,
			AgendaID:      doctolibConfig.AgendaID,
			VisitMotiveID: doctolibConfig.VisitMotiveID,
			VaccineName:   doctolibConfig.VaccineName,
		}
		if doctolibConfig.Delay != nil {
			tmp.Delay = time.Duration(*doctolibConfig.Delay)
		}
		if doctolibConfig.Detail != nil {
			tmp.Detail = *doctolibConfig.Detail
		}
		ret = append(ret, &tmp)
	}
	return ret, nil
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
		log.Error(err)
		return
	}
	bot, err := tgbotapi.NewBotAPI(viper.GetString("telegram-token"))
	if err != nil {
		log.Error(err)
		return
	}
	chatModel := chat.NewModel(db)
	telegram := NewBot(bot, chatModel)
	doctolibs, err := getAllDoctolibSources()
	if err != nil {
		log.Error(err)
		return
	}
	var s = []Fetcher{
		&sources.PuntoMedico{},
		&sources.MedicoLeopoldPlatz{},
		&sources.ArkonoPlatz{},
		&sources.Helios{},
	}

	s = append(s, doctolibs...)

	var runCMD = &cobra.Command{
		Use:   "run",
		Short: "run the telegram bot",
		Run: func(cmd *cobra.Command, args []string) {
			var wg sync.WaitGroup

			wg.Add(2)

			go func() {
				defer wg.Done()
				err := telegram.HandleNewUsers()
				if err != nil {
					log.Error(err)
					return
				}
			}()

			go func() {
				defer wg.Done()
				for range time.Tick(5 * time.Second) {
					go fetchAllAppointment(s, telegram)
				}
			}()

			wg.Wait()
		},
	}

	var sendCMD = &cobra.Command{
		Use:   "send",
		Short: "send message to all active user",
		RunE: func(cmd *cobra.Command, args []string) error {
			chats, err := chatModel.List(nil)
			if err != nil {
				return err
			}
			for _, chat := range chats {
				msg := tgbotapi.MessageConfig{
					BaseChat: tgbotapi.BaseChat{
						ChatID:           chat.ID,
						ReplyToMessageID: 0,
					},
					Text:                  "Hey, Thanks again for using the bot!\n\nI added a lot of Doctolib clinic and doctors and the Arkonoplatz clinic",
					DisableWebPagePreview: true,
				}
				_, err := bot.Send(msg)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	rootCmd.AddCommand(runCMD)
	rootCmd.AddCommand(sendCMD)

	rootCmd.Execute()
}
