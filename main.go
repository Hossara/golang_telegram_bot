package main

import (
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"net/http"
	"os"
)

type ConfigurationType struct {
	NewUpdateOffset int
	Timeout         int
}

var Configuration = &ConfigurationType{
	NewUpdateOffset: 0,
	Timeout:         60,
}

var (
	client *http.Client
	bot    *tgbotapi.BotAPI
)

func main() {
	// Read .env file
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
		return
	}

	// Set up SOCKS5 proxy
	if os.Getenv("PROXY") != "" {
		// Create socks 5 connection
		dialer, err := proxy.SOCKS5("tcp", os.Getenv("PROXY"), &proxy.Auth{
			User:     "",
			Password: "",
		}, proxy.Direct)

		if err != nil {
			log.Fatalf("failed to connect to proxy: %v", err)
			return
		}

		// Use it in http transport layer
		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		}

		// Create HTTP client with the proxy
		client = &http.Client{Transport: transport}
		bot, err = tgbotapi.NewBotAPIWithClient(os.Getenv("TELEGRAM_TOKEN"), client)
	} else {
		// Create telegram bot instance without proxy
		bot, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	}

	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
		return
	}

	// Set true. the bot to use debug mode (verbose logging).
	bot.Debug = false

	// Set up updates configuration.
	u := tgbotapi.NewUpdate(Configuration.NewUpdateOffset)
	u.Timeout = Configuration.Timeout

	// Get updates from the Telegram API.
	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		log.Fatalf("Error getting updates: %v", err)
		return
	}

	for update := range updates {
		// Ignore any non-Message updates.
		if update.Message == nil {
			continue
		}

		log.Printf("New message from user @%s", update.Message.From.UserName)

		// Respond to the user.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Message: %s", update.Message.Text))

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
