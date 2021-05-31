package main

import (
	"database/sql"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
)

type Telegram struct {
	db      *sql.DB
	bot     *tgbotapi.BotAPI
	channel int64
}

func NewBot(bot *tgbotapi.BotAPI, db *sql.DB) *Telegram {
	return &Telegram{
		bot: bot,
		db:  db,
	}
}

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

func (t *Telegram) SendMessageToAllUser(message string) error {
	rows, err := t.db.Query("SELECT id FROM chats")
	if err != nil {
		return err
	}
	ids, err := scanRows(rows)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(len(ids) + 1)

	for _, id := range ids {
		id := id
		go func() {
			defer wg.Done()
			err := t.SendMessage(message, id)
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

func scanRows(rows *sql.Rows) ([]int64, error) {
	entries := make([]int64, 0)

	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		entries = append(entries, id)
	}

	return entries, nil
}

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

			statement, err := t.db.Prepare("INSERT INTO chats (id) VALUES (?)")
			if err != nil {
				fmt.Println(err)
				continue
			}
			_, err = statement.Exec(update.Message.Chat.ID)
			if err != nil {
				if err.Error() == "UNIQUE constraint failed: chats.id" {
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

			statement, err := t.db.Prepare("DELETE FROM chats WHERE id = (?)")
			if err != nil {
				fmt.Println(err)
				continue
			}
			_, err = statement.Exec(update.Message.Chat.ID)
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
