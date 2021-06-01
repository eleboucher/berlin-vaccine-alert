package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/eleboucher/covid/models/chat"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
)

const (
	startButton      = "Start"
	stopButton       = "Stop"
	filterButton     = "Add filters (multiple choices available)"
	azButton         = "Look for AstraZeneca"
	jjButton         = "Look for Johnson & Johnson"
	biontechButton   = "Look for Biontech/Pfizer (not in vaccination centers)"
	vcButton         = "Look for Vaccination Centers"
	everythingButton = "Look for everything"
	contributeButton = "Contribute"
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
		tgbotapi.NewKeyboardButton(biontechButton),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(vcButton),
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
	channel   int64
	chatModel *chat.Model
}

// NewBot return a new Telegram Bot
func NewBot(bot *tgbotapi.BotAPI, chatModel *chat.Model) *Telegram {
	return &Telegram{
		bot:       bot,
		chatModel: chatModel,
	}
}

// SendMessage send a message in string to a channel id
func (t *Telegram) SendMessage(message string, channel int64) error {
	fmt.Printf("sending message %s on channel %d", message, channel)
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           channel,
			ReplyToMessageID: 0,
		},
		Text:                  message,
		DisableWebPagePreview: true,
	}
	_, err := t.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

// SendMessageToAllUser send a message to all the enabled users
func (t *Telegram) SendMessageToAllUser(result *Result) error {
	chats, err := t.chatModel.List(result.VaccineName)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(len(chats) + 1)

	for _, chat := range chats {
		chat := chat
		go func() {
			defer wg.Done()
			err := t.SendMessage(result.Message, chat.ID)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	go func() {
		wg.Done()
		err := t.SendMessage(result.Message, viper.GetInt64("telegram-channel"))
		if err != nil {
			fmt.Println(err)
		}
	}()
	wg.Wait()
	return nil
}

// HandleNewUsers handle the commands from telegrams
func (t *Telegram) HandleNewUsers() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := t.bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		update := update
		go func() {
			if update.Message == nil { // ignore any non-Message Updates
				return
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			switch update.Message.Text {
			case "open", backButton:
				msg.ReplyMarkup = keyboard
				t.bot.Send(msg)
			case "close":
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				t.bot.Send(msg)
			case contributeButton:
				err = t.SendMessage("Hey you 🚀,\nThanks a lot for using the bot,\n\n\nFeel free to contribute on Github: https://github.com/eleboucher/berlin-vaccine-alert\n\n\nOr feel free to contribute on Paypal https://paypal.me/ELeboucher or Buy me a beer https://www.buymeacoffee.com/eleboucher", update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
				}
			case filterButton:
				msg.ReplyMarkup = filtersKeyboard
				t.bot.Send(msg)
			case stopButton:
				err := t.stopChat(update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
				}
			case startButton:
				err := t.startChat(update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
				}
			case azButton:
				_, err := t.chatModel.UpdateFilters(update.Message.Chat.ID, AstraZeneca)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = t.SendMessage("subscribed to AstraZeneca updates", update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
					return
				}
			case jjButton:
				_, err := t.chatModel.UpdateFilters(update.Message.Chat.ID, JohnsonAndJohnson)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = t.SendMessage("subscribed to Johnson And Johnson updates", update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
					return
				}
			case vcButton:
				_, err := t.chatModel.UpdateFilters(update.Message.Chat.ID, VaccinationCenter)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = t.SendMessage("subscribed to Vaccination centers updates", update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
					return
				}
			case biontechButton:
				_, err := t.chatModel.UpdateFilters(update.Message.Chat.ID, Pfizer)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = t.SendMessage("subscribed to Vaccination centers updates", update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
					return
				}
			case everythingButton:
				_, err := t.chatModel.UpdateFilters(update.Message.Chat.ID, "")
				if err != nil {
					fmt.Println(err)
					return
				}
				err = t.SendMessage("subscribed to every updates", update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
					return
				}
			case infoFilterButton:
				chat, err := t.chatModel.Find(update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
					return
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
					fmt.Println(err)
					return
				}
			}

			switch update.Message.Command() {
			case "start":
				err := t.startChat(update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
				}
			case "stop":
				err := t.stopChat(update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
				}
			case "open":
				msg.ReplyMarkup = filtersKeyboard
				t.bot.Send(msg)
			case "contribute":
				err = t.SendMessage("Hey you 🚀,\nThanks a lot for using the bot,\n\n\nFeel free to contribute on Github: https://github.com/eleboucher/berlin-vaccine-alert\n\n\nOr feel free to contribute on Paypal https://paypal.me/ELeboucher or Buy me a beer https://www.buymeacoffee.com/eleboucher", update.Message.Chat.ID)
				if err != nil {
					fmt.Println(err)
				}
			}
		}()
	}
	fmt.Println("done with telegram handler")
	return nil
}

func (t *Telegram) startChat(chatID int64) error {
	fmt.Printf("adding chat %d\n", chatID)

	_, err := t.chatModel.Create(chatID)
	if err != nil {
		if errors.Is(err, chat.ErrChatAlreadyExist) {
			_, err := t.chatModel.Enable(chatID)
			if err != nil {
				return err
			}
			err = t.SendMessage("Hey Again! You are already added to the subscription list, you will receive appointments shortly when they will be available", chatID)
			if err != nil {
				return err
			}
			return err
		}
		fmt.Println(err)
		return err
	}
	err = t.SendMessage("Welcome 👋🏼! You are now added to the subscription list, you will receive appointments shortly when they will be available", chatID)
	if err != nil {
		return err
	}
	return nil
}

func (t *Telegram) stopChat(chatID int64) error {
	fmt.Printf("removing chat %d\n", chatID)

	_, err := t.chatModel.Delete(chatID)
	if err != nil {
		return err
	}
	err = t.SendMessage("Removed from the list successfully. if you want to receive messages again type /start", chatID)
	if err != nil {
		return err
	}
	return nil
}
