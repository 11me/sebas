package main

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/go-telegram/bot"
)

type Bot struct {
	bot  *bot.Bot
	conf *Config
	ctx  context.Context

	isStarted atomic.Bool
}

func NewBot(conf *Config) *Bot {
	return &Bot{conf: conf}
}

func (b *Bot) Start(ctx context.Context) {
	if b.isStarted.Load() {
		return
	}

	b.ctx = ctx

	var err error
	b.bot, err = bot.New(b.conf.TgBot.Token)
	if err != nil {
		panic(err)
	}

	b.isStarted.Store(true)
	defer b.isStarted.Store(false)

	b.bot.Start(ctx)
}

func (b *Bot) SendMessage(chatID any, text string) (err error) {
	_, err = b.bot.SendMessage(b.ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
