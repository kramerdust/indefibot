package bot

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kramerdust/indefibot/exegete"
	"golang.org/x/net/proxy"
)

type Bot struct {
	botAPI            *tgbotapi.BotAPI
	expositorProvider exegete.ExpositorProvider
	userDataProvider  UserDataProvider
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

func NewBot(config *Config, userDataProvider UserDataProvider) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to Telegram: %s ", err)
	}
	return &Bot{botAPI: bot, userDataProvider: userDataProvider}, nil
}

func NewBotWithProxy(config *Config, userDataProvider UserDataProvider) (*Bot, error) {
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
	return &Bot{botAPI: bot, userDataProvider: userDataProvider}, nil
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

		switch {
		case u.Message != nil:
			log.Println("Message")
			word := u.Message.Text
			expositor, err := b.expositorProvider.GetWordExpositor("en", word)

			if err != nil {
				b.replyWordNotFound(&u, word)
				continue
			}

			b.userDataProvider.SetUserExpositor(u.Message.Chat.ID, expositor)

			sp, _ := expositor.GetSpelling()
			d := expositor.GetSenses()[0].GetDefinitions()[0]

			card := Card{
				Word:          word,
				Transcription: sp,
				Definition:    d,
				Page:          1,
			}
			t := template.Must(template.New("card").Parse(CardTemplate))
			var out bytes.Buffer
			t.Execute(&out, card)

			msg := tgbotapi.NewMessage(u.Message.Chat.ID, out.String())
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(StartButtonRow())
			msg.ParseMode = "Markdown"

			s, err := b.botAPI.Send(msg)
			if err != nil {
				log.Println("Error in message", s, err)
			}

		case u.CallbackQuery != nil:
			log.Println("CallbackQuery")
		}
	}
}

func (b *Bot) replyWordNotFound(u *tgbotapi.Update, word string) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, fmt.Sprintf("Can't find *%s*", word))
	msg.ReplyToMessageID = u.Message.MessageID
	msg.ParseMode = "Markdown"
	b.botAPI.Send(msg)
}

// 		audio, err := expositor.GetAudio()
// 		msg := tgbotapi.NewAudioUpload(u.Message.Chat.ID, tgbotapi.FileReader{
// 			word,
// 			audio,
// 			-1,
// 		})
// 		_, err = b.botAPI.Send(msg)
// 		if err != nil {
// 			log.Printf("Error while answering to user! %s\n", err)
// 		}

func StartButtonRow() []tgbotapi.InlineKeyboardButton {
	row := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️", "dummy"),
		tgbotapi.NewInlineKeyboardButtonData("➡️", "dummy"),
		tgbotapi.NewInlineKeyboardButtonData("🔈", "dummy"),
	)
	return row
}
