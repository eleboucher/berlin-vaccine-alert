package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/eleboucher/berlin-vaccine-alert/models/chat"
	"github.com/eleboucher/berlin-vaccine-alert/vaccines"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const (
	startButton      = "Start"
	stopButton       = "Stop"
	filterButton     = "Add filters (multiple choices available)"
	azButton         = "Look for AstraZeneca"
	jjButton         = "Look for Johnson & Johnson"
	vcButton         = "Look for MRNA vaccine (clinics and vaccination centers)"
	everythingButton = "Look for everything"
	contributeButton = "Contribute and support"
	infoFilterButton = "Info about filters"
	backButton       = "Back"
)

var keyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(startButton),
		tgbotapi.NewKeyboardButton(stopButton),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(filterButton),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(contributeButton),
	),
)

var filtersKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(azButton),
		tgbotapi.NewKeyboardButton(jjButton),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(vcButton),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(everythingButton),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(infoFilterButton),
		tgbotapi.NewKeyboardButton(backButton),
	),
)

// Telegram Holds the structure for the telegram bot
type Telegram struct {
	bot       *tgbotapi.BotAPI
	limiter   *rate.Limiter
	channel   int64
	chatModel *chat.Model
}

// NewBot return a new Telegram Bot
func NewBot(bot *tgbotapi.BotAPI, chatModel *chat.Model) *Telegram {
	return &Telegram{
		bot:       bot,
		chatModel: chatModel,
		limiter:   rate.NewLimiter(rate.Every(time.Second/30), 1),
	}
}

// SendMessage send a message in string to a channel id
func (t *Telegram) SendMessage(message string, channel int64) error {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           channel,
			ReplyToMessageID: 0,
		},
		Text:                  message,
		DisableWebPagePreview: true,
	}
	ctx := context.Background()
	err := t.limiter.Wait(ctx)
	if err != nil {
		return err
	}
	_, err = t.bot.Send(msg)
	if err != nil {
		if strings.Contains(err.Error(), "Forbidden:") {
			_, err := t.chatModel.Delete(channel)
			if err != nil {
				return err
			}
		}
		return err
	}
	return nil
}

// SendMessageToAllUser send a message to all the enabled users
func (t *Telegram) SendMessageToAllUser(result *vaccines.Result) error {
	chats, err := t.chatModel.List(&result.VaccineName)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(len(chats))
	log.Infof("sending message %s for %d users\n", result.Message, len(chats))

	for _, chat := range chats {
		chat := chat
		go func() {
			defer wg.Done()
			chat := chat
			err := t.SendMessage(result.Message, chat.ID)
			if err != nil {
				log.Error(err)
			}
		}()
	}
	wg.Wait()
	return nil
}

// HandleNewUsers handle the commands from telegrams
func (t *Telegram) HandleNewUsers() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)
	for update := range updates {
		update := update
		go func() {
			if update.Message == nil { // ignore any non-Message Updates
				return
			}
			logrus.Infof("Receiving new message: %#s", update.Message)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			switch update.Message.Text {
			case "open", backButton:
				msg.ReplyMarkup = keyboard
				_, err := t.bot.Send(msg)
				if err != nil {
					log.Error(err)
				}
			case "close":
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				_, err := t.bot.Send(msg)
				if err != nil {
					log.Error(err)
				}
			case contributeButton:
				err := t.SendMessage("Hey you 🚀,\nThanks a lot for using the bot,\n\n\nFeel free to contribute on Github: https://github.com/eleboucher/berlin-vaccine-alert\n\n\nOr feel free to contribute on Paypal https://paypal.me/ELeboucher or Buy me a beer https://www.buymeacoffee.com/eleboucher", update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			case filterButton:
				msg.ReplyMarkup = filtersKeyboard
				_, err := t.bot.Send(msg)
				if err != nil {
					log.Error(err)
				}
			case stopButton:
				err := t.stopChat(update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			case startButton:
				err := t.startChat(update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			case azButton:
				_, err := t.chatModel.UpdateFilters(update.Message.Chat.ID, vaccines.AstraZeneca)
				if err != nil {
					log.Error(err)
				}
				err = t.SendMessage("subscribed to AstraZeneca updates", update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			case jjButton:
				_, err := t.chatModel.UpdateFilters(update.Message.Chat.ID, vaccines.JohnsonAndJohnson)
				if err != nil {
					log.Error(err)
				}
				err = t.SendMessage("subscribed to Johnson And Johnson updates", update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			case vcButton:
				_, err := t.chatModel.UpdateFilters(update.Message.Chat.ID, vaccines.MRNA)
				if err != nil {
					log.Error(err)
				}
				err = t.SendMessage("subscribed to MRNA vaccines updates", update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			case everythingButton:
				_, err := t.chatModel.UpdateFilters(update.Message.Chat.ID, "")
				if err != nil {
					log.Error(err)
				}
				err = t.SendMessage("subscribed to every updates", update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			case infoFilterButton:
				chat, err := t.chatModel.Find(update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
				var filters string
				if len(chat.Filters) == 0 {
					filters = "unfiltered"
				} else {
					filters = strings.Join(chat.Filters, "\n")
				}
				msg := fmt.Sprintf("your current filters are :\n%s\n\nSelect %s to reset them", filters, everythingButton)
				err = t.SendMessage(msg, update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			}

			switch update.Message.Command() {
			case "start":
				err := t.startChat(update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			case "stop":
				err := t.stopChat(update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			case "open":
				msg.ReplyMarkup = filtersKeyboard
				t.bot.Send(msg)
			case "contribute":
				err := t.SendMessage("Hey you 🚀,\nThanks a lot for using the bot,\n\n\nFeel free to contribute on Github: https://github.com/eleboucher/berlin-vaccine-alert\n\n\nOr feel free to contribute on Paypal https://paypal.me/ELeboucher or Buy me a beer https://www.buymeacoffee.com/eleboucher", update.Message.Chat.ID)
				if err != nil {
					log.Error(err)
				}
			}
		}()
	}

	log.Info("done with telegram handler")
	return nil
}

func (t *Telegram) startChat(chatID int64) error {
	log.Infof("adding chat %d\n", chatID)

	_, err := t.chatModel.Create(chatID)
	if err != nil {
		if errors.Is(err, chat.ErrChatAlreadyExist) {
			_, err := t.chatModel.Enable(chatID)
			if err != nil {
				return err
			}
			err = t.SendMessage(`
Hey Again!
You are already added to the subscription list, you will receive appointments shortly when they will be available!

I hope this bot helps you in your research to get the vaccine!

Provide feedback 📢 on Reddit: https://www.reddit.com/r/berlinvaccination/comments/np81h5/telegram_bot_to_get_a_vaccine_appointment/

Feel free to help me with the cost or with the code, via:
💸 Donate via PayPal: https://paypal.me/ELeboucher
🍻 Buy me a beer: https://www.buymeacoffee.com/eleboucher
🧑‍💻 Contribute to the code:  https://github.com/eleboucher/berlin-vaccine-alert

I really hope it can help you to find your appointment!

Stay Safe, and thanks for your support! ❤️`, chatID)
			if err != nil {
				return err
			}
			return err
		}
		return err
	}
	err = t.SendMessage(`
Welcome 👋🏼!
You are now added to the subscription list, you will receive appointments shortly when they will be available
I hope this bot helps you in your research to get the vaccine!

Provide feedback 📢 on Reddit: https://www.reddit.com/r/berlinvaccination/comments/np81h5/telegram_bot_to_get_a_vaccine_appointment/

Feel free to help me with the cost of the bot or with the code, via:
💸 Donate via PayPal: https://paypal.me/ELeboucher
🍻 Buy me a beer: https://www.buymeacoffee.com/eleboucher
🧑‍💻 Contribute to the code:  https://github.com/eleboucher/berlin-vaccine-alert

I really hope it can help you to find your appointment!

Stay Safe, and thanks for your support! ❤️`, chatID)
	if err != nil {
		return err
	}
	return nil
}

func (t *Telegram) stopChat(chatID int64) error {
	log.Infof("removing chat %d\n", chatID)

	_, err := t.chatModel.Delete(chatID)
	if err != nil {
		return err
	}
	err = t.SendMessage(`
Hey!

You are removed from the list. If you want to receive messages again type /start.

I hope you had book an appointment and you are getting vaccinated soon!

If you have any feedback feel free to post something on Reddit: https://www.reddit.com/r/berlinvaccination/comments/np81h5/telegram_bot_to_get_a_vaccine_appointment/

Feel free to help me with the cost of the bot or with the code, via:
💸 Donate via PayPal: https://paypal.me/ELeboucher
🍻 Buy me a beer: https://www.buymeacoffee.com/eleboucher
🧑‍💻 Contribute to the code:  https://github.com/eleboucher/berlin-vaccine-alert

Stay Safe, and thanks for your support! ❤️`, chatID)
	if err != nil {
		return err
	}
	return nil
}
