package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/eleboucher/covid/models/chat"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
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
func (t *Telegram) SendMessageToAllUser(message string) error {
	chats, err := t.chatModel.List()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(len(chats) + 1)

	for _, chat := range chats {
		chat := chat
		go func() {
			defer wg.Done()
			err := t.SendMessage(message, chat.ID)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	go func() {
		wg.Done()
		err := t.SendMessage(message, viper.GetInt64("telegram-channel"))
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
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		switch update.Message.Command() {
		case "start":
			fmt.Printf("adding chat %d\n", update.Message.Chat.ID)

			_, err := t.chatModel.Create(update.Message.Chat.ID)
			if err != nil {
				if errors.Is(err, chat.ErrChatAlreadyExist) {
					_, err := t.chatModel.Enable(update.Message.Chat.ID)
					if err != nil {
						fmt.Println(err)
						continue
					}
					err = t.SendMessage("Hey Again! You are already added to the subscription list, you will receive appointments shortly when they will be available", update.Message.Chat.ID)
					if err != nil {
						fmt.Println(err)
					}
					continue
				}
				fmt.Println(err)
				continue
			}
			err = t.SendMessage("Welcome 👋🏼! You are now added to the subscription list, you will receive appointments shortly when they will be available", update.Message.Chat.ID)
			if err != nil {
				fmt.Println(err)
			}
		case "stop":
			fmt.Printf("removing chat %d\n", update.Message.Chat.ID)

			_, err := t.chatModel.Delete(update.Message.Chat.ID)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = t.SendMessage("Removed from the list successfully. if you want to receive messages again type /start", update.Message.Chat.ID)
			if err != nil {
				fmt.Println(err)
			}
		case "contribute":
			err = t.SendMessage("Hey you 🚀,\n Thanks a lot for using the bot,\n\n\nFeel free to contribute on Github: https://github.com/eleboucher/berlin-vaccine-alert\n\n\nOr feel free to contribute on Paypal https://paypal.me/ELeboucher or Buy me a beer https://www.buymeacoffee.com/eleboucher", update.Message.Chat.ID)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	fmt.Println("done with telegram handler")
	return nil
}
