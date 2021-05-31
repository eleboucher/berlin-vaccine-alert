package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Telegram struct {
	bot     *tgbotapi.BotAPI
	channel int64
}

func NewBot(bot *tgbotapi.BotAPI, channel int64) *Telegram {
	return &Telegram{
		bot:     bot,
		channel: channel,
	}
}

func (t *Telegram) SendMessage(message string) error {
	msg := tgbotapi.NewMessage(t.channel, message)
	_, err := t.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
