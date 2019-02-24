package bot

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kramerdust/indefibot/exegete"
	"golang.org/x/net/proxy"
)

type Bot struct {
	botAPI            *tgbotapi.BotAPI
	expositorProvider exegete.ExpositorProvider
}

type Config struct {
	Token    string      `yaml:"botToken"`
	ProxyURL string      `yaml:"proxyURL"`
	Source   ExternalAPI `yaml:"source"`
}

type ExternalAPI struct {
	AppID  string `yaml:"appID"`
	AppKey string `yaml:"appKey"`
}

func NewBot(config *Config) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to Telegram: %s ", err)
	}
	return &Bot{botAPI: bot}, nil
}

func NewBotWithProxy(config *Config) (*Bot, error) {
	dialer, err := proxy.SOCKS5("tcp", config.ProxyURL, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("Proxy error: %s ", err)
	}
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	bot, err := tgbotapi.NewBotAPIWithClient(config.Token, httpClient)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to Telegram: %s ", err)
	}
	return &Bot{botAPI: bot}, nil
}

func (b *Bot) SetExpositorProvider(provider exegete.ExpositorProvider) {
	b.expositorProvider = provider
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.botAPI.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}
	b.handleUpdates(updates)
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for u := range updates {
		if u.Message == nil {
			continue
		}
		if u.Message.IsCommand() {
			switch u.Message.Command() {
			case "pronounce":
				word := u.Message.CommandArguments()
				expositor, err := b.expositorProvider.GetWordExpositor("en", word)
				if err != nil {
					msg := tgbotapi.NewMessage(u.Message.Chat.ID, fmt.Sprintf("Error! %s", err))
					msg.ReplyToMessageID = u.Message.MessageID
					continue
				}
				audio, err := expositor.GetAudio()
				msg := tgbotapi.NewAudioUpload(u.Message.Chat.ID, tgbotapi.FileReader{
					word,
					audio,
					-1,
				})
				_, err = b.botAPI.Send(msg)
				if err != nil {
					log.Printf("Error while answering to user! %s\n", err)
				}

			}
		} else {
			msg := tgbotapi.NewMessage(u.Message.Chat.ID, "I don't undestand you!")
			msg.ReplyToMessageID = u.Message.MessageID

			b.botAPI.Send(msg)
		}
	}
}
