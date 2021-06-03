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

var s = []Fetcher{
	&sources.PuntoMedico{},
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
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158431",
		VaccineName:   vaccines.VaccinationCenter,
		PraticeID:     "158431",
		AgendaID:      "457686-457680-457681-457684-457685-457688-457689-457691-457693-457687-457690",
		VisitMotiveID: "2495719",
		Detail:        "Arena",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158434",
		VaccineName:   vaccines.VaccinationCenter,
		PraticeID:     "158434",
		AgendaID:      "457479-457450-457475-457455-457459-457454-457447-457446-457458-457456-457472-457476-457452-457480-457461-457451-457468-457473",
		VisitMotiveID: "2495719",
		Detail:        "Messe",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158437",
		VaccineName:   vaccines.VaccinationCenter,
		PraticeID:     "158437",
		AgendaID:      "457976-457928-457927-457930-457917-457939-457975-457964-457970-457907-457924-457971-457912-457916-457922-457967-457933-457940-457968-457963-457973-457931-457915-457918-457938-457935-457979-457966-457926-457941-457937-457951-457952-457954-457947-457977-457923",
		VisitMotiveID: "2537716",
		Detail:        "Erika-Heß-Eisstadion",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158435",
		VaccineName:   vaccines.VaccinationCenter,
		PraticeID:     "158435",
		AgendaID:      "457195-457211-457201-457991-457205-457193",
		VisitMotiveID: "2495719",
		Detail:        "Velodrom",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-158436",
		VaccineName:   vaccines.VaccinationCenter,
		PraticeID:     "158436",
		AgendaID:      "457388-457320-457313-457379-457374-457302-457307-457314-457355-457308-457392-457395-457305-457377-457396-457316-457390-457382-457385-457318",
		VisitMotiveID: "2495719",
		Detail:        "Flughafen Berlin-Tegel Pfizer",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-191611",
		VaccineName:   vaccines.VaccinationCenter,
		PraticeID:     "191611",
		AgendaID:      "467906-481913-481915-481920-481917-467934-467937-467938-467939-467910-467908-467903-467907-467935-467936-467893-467895-467896-467900-467901-467905-467911-467897-467898-467912-467940-481914-481916-481919-481921-467894-467933-467899",
		VisitMotiveID: "2537716",
		Detail:        "Flughafen Tempelhof Moderna",
	},
	&sources.Doctolib{
		URL:           "https://www.doctolib.de/institut/berlin/ciz-berlin-berlin?pid=practice-191612",
		VaccineName:   vaccines.VaccinationCenter,
		PraticeID:     "191612",
		AgendaID:      "466152-466154-466155-466156-466158-466159-466160-466161-466153-466157",
		VisitMotiveID: "2537716",
		Detail:        "Flughafen Berlin-Tegel Moderna",
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
