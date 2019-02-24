package bot

import (
	"fmt"
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
	for _ = range updates {

	}
}
