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
	Fetch() ([]*vaccines.Result, error)
	ShouldSendResult(result []*vaccines.Result) bool
	ResultSentNow(result []*vaccines.Result)
}

var s = []Fetcher{
	&sources.PuntoMedico{},
	&sources.VaccineCenter{},
	&sources.MedicoLeopoldPlatz{},
	&sources.ArkonoPlatz{},
	&sources.Helios{},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/praxis/brandenburg-an-der-havel/corona-schutzimpfung-gzb",
		VaccineName:   vaccines.JohnsonAndJohnson,
		PraticeID:     "186461",
		AgendaID:      "472530",
		VisitMotiveID: "2877045",
		Detail:        "(for 40+)",
		Delay:         10,
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/praxis/brandenburg-an-der-havel/corona-schutzimpfung-gzb",
		VaccineName:   vaccines.AstraZeneca,
		PraticeID:     "186461",
		AgendaID:      "472530",
		VisitMotiveID: "2741487",
		Detail:        "(for 40+)",
		Delay:         10,
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/allgemeinmedizin/berlin/sophie-ruggeberg",
		VaccineName:   vaccines.JohnsonAndJohnson,
		PraticeID:     "114976",
		AgendaID:      "190434",
		VisitMotiveID: "2886231",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/allgemeinmedizin/berlin/sophie-ruggeberg",
		VaccineName:   vaccines.AstraZeneca,
		PraticeID:     "114976",
		AgendaID:      "190434",
		VisitMotiveID: "2764198",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/facharzt-fur-hno/berlin/babak-mayelzadeh",
		VaccineName:   vaccines.AstraZeneca,
		PraticeID:     "120549",
		AgendaID:      "305777",
		VisitMotiveID: "2862419",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/facharzt-fur-hno/berlin/babak-mayelzadeh",
		VaccineName:   vaccines.JohnsonAndJohnson,
		PraticeID:     "120549",
		AgendaID:      "305777",
		VisitMotiveID: "2879179",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/facharzt-fur-hno/berlin/rafael-hardy",
		VaccineName:   vaccines.Pfizer,
		PraticeID:     "22563",
		AgendaID:      "56915",
		VisitMotiveID: "2733996",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/innere-und-allgemeinmediziner/berlin/oliver-staeck",
		VaccineName:   vaccines.AstraZeneca,
		PraticeID:     "178663",
		AgendaID:      "268801",
		VisitMotiveID: "2784656",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/innere-und-allgemeinmediziner/berlin/oliver-staeck",
		VaccineName:   vaccines.JohnsonAndJohnson,
		PraticeID:     "178663",
		AgendaID:      "268801",
		VisitMotiveID: "2885945",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/praxis/berlin/praxis-fuer-orthopaedie-und-unfallchirurgie-neukoelln",
		VaccineName:   vaccines.AstraZeneca,
		PraticeID:     "28436",
		AgendaID:      "464751",
		VisitMotiveID: "2811460",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/praxis/berlin/praxis-fuer-orthopaedie-und-unfallchirurgie-neukoelln",
		VaccineName:   vaccines.AstraZeneca,
		PraticeID:     "28436",
		AgendaID:      "464751",
		VisitMotiveID: "2811530",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/medizinisches-versorgungszentrum-mvz/berlin/ambulantes-gynaekologisches-operationszentrum",
		VaccineName:   vaccines.Pfizer,
		PraticeID:     "107774",
		AgendaID:      "439400",
		VisitMotiveID: "2757216",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/medizinisches-versorgungszentrum-mvz/berlin/ambulantes-gynaekologisches-operationszentrum",
		VaccineName:   vaccines.AstraZeneca,
		PraticeID:     "107774",
		AgendaID:      "439400",
		VisitMotiveID: "2885841",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/medizinisches-versorgungszentrum-mvz/berlin/ambulantes-gynaekologisches-operationszentrum",
		VaccineName:   vaccines.JohnsonAndJohnson,
		PraticeID:     "107774",
		AgendaID:      "439400",
		VisitMotiveID: "2880391",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/krankenhaus/berlin/gkh-havelhoehe-impfzentrum",
		VaccineName:   vaccines.AstraZeneca,
		PraticeID:     "162056",
		AgendaID:      "469719",
		VisitMotiveID: "2836657",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/krankenhaus/berlin/gkh-havelhoehe-impfzentrum",
		VaccineName:   vaccines.JohnsonAndJohnson,
		PraticeID:     "162056",
		AgendaID:      "469719",
		VisitMotiveID: "2898162",
	},
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
			log.Info("Starting fetch")
			res, err := fetcher.Fetch()
			if err != nil {
				errChan <- err
				return
			}
			log.Infof("Received %d result\n", len(res))
			if len(res) > 0 && fetcher.ShouldSendResult(res) {
				fetcher.ResultSentNow(res)
				for _, r := range res {

					bot.SendMessageToAllUser(r)
					if err != nil {
						errChan <- err
						return
					}
				}
				log.Info("messages sent on telegram")
			}
			done <- true
		}()
	}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-done:
			log.Info("fetch done")
		case <-timeout:
			return
		case err := <-errChan:
			log.Errorf("%v\n", err)

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
