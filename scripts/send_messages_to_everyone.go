package main

import (
	"database/sql"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	db, err := NewDB()
	if err != nil {
		return
	}
	bot, err := tgbotapi.NewBotAPI("todo")
	if err != nil {
		return
	}
	rows, err := db.Query("SELECT id FROM chats")
	if err != nil {
		return
	}
	ids, err := scanRows(rows)
	if err != nil {
		return
	}
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(ids))
	fmt.Printf("sending to %d chats\n", len(ids))
	for _, id := range ids {
		id := id
		go func() {
			defer wg.Done()
			msg := tgbotapi.MessageConfig{
				BaseChat: tgbotapi.BaseChat{
					ChatID:           id,
					ReplyToMessageID: 0,
				},
				Text:                  "Hey, Thanks again for using the bot!\nYou can now filter which vaccine you want!\ntype `open` or `/open` and use the button to add a filter (or multiple)\n\nLet's get out of this mess together!",
				DisableWebPagePreview: true,
			}
			_, err := bot.Send(msg)
			if err != nil {
				fmt.Println("abh", err)
				return
			}
		}()
	}
	wg.Wait()
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
