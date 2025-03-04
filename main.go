package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	conf, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	client := NewExchangeClient(conf)
	bot := NewBot(conf)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		bot.Start(ctx)
		return nil
	})
	eg.Go(func() error {
		watchDelistings(ctx, conf, client, bot)
		return nil
	})

	done := make(chan struct{}, 1)
	go func() {
		eg.Wait()
		done <- struct{}{}
	}()

	<-ctx.Done()
	slog.Info("Stopping application")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
	case <-done:
	}
}

func watchDelistings(
	ctx context.Context,
	conf *Config,
	client *ExchangeClient,
	bot *Bot,
) {
	slog.Info("start watching delisting on Binance")
	notified := make(map[string]bool)
	interval := 10 * time.Second
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for {
		func() (err error) {
			defer func() {
				if err != nil {
					slog.Error("delisting watcher failed", "err", err)
				} else if r := recover(); r != nil {
					slog.Warn("delistingWatcher recovered from panic", "panic", r, "stack", string(debug.Stack()))
				}

				ticker.Reset(interval)
			}()

			// Get new delistings.
			var delistings []DelistSchedule

			delistings, err = client.GetDelistings()
			if err != nil {
				return fmt.Errorf("failed to get delistings: %w", err)
			}

			// Prepare message.
			var message string

		outer:
			for _, d := range delistings {
				var symbols string
				for i, symbol := range d.Symbols {
					if _, ok := notified[symbol]; ok {
						continue outer
					}

					if i == len(d.Symbols)-1 {
						symbols += fmt.Sprintf("%s\n\n", symbol)
					} else {
						symbols += fmt.Sprintf("%s\n", symbol)
					}

					notified[symbol] = true
				}
				message += fmt.Sprintf(
					"New delisting scheduled on %s\n%s", d.DelistTime.ToTime().UTC().String(), symbols)
			}

			// Send to channel.
			if message != "" {
				err = bot.SendMessage(conf.TgBot.DelistingChannelID, message)
				if err != nil {
					return fmt.Errorf("failed to send message: %w", err)
				}
			}

			return
		}()

		select {
		case <-ticker.C:
		case <-ctx.Done():
			slog.Info("Stopping delisting watcher")
			return
		}
	}
}
