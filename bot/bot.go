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
	wordDataProvider  WordDataProvider
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

func NewBot(config *Config, wordDataProvider WordDataProvider) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to Telegram: %s ", err)
	}
	return &Bot{botAPI: bot, wordDataProvider: wordDataProvider}, nil
}

func NewBotWithProxy(config *Config, wordDataProvider WordDataProvider) (*Bot, error) {
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
	return &Bot{botAPI: bot, wordDataProvider: wordDataProvider}, nil
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
	updates.Clear()
	b.handleUpdates(updates)
}

func (b *Bot) StartWebHook() {
	updates := b.botAPI.ListenForWebhook("/" + b.botAPI.Token)
	updates.Clear()
	b.handleUpdates(updates)
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for u := range updates {
		switch {
		case u.Message != nil:
			log.Println("Message")
			word := u.Message.Text
			var expositor exegete.Expositor
			var err error
			expositor, ok := b.wordDataProvider.GetWordExpositor(word)
			if !ok {
				expositor, err = b.expositorProvider.GetWordExpositor("en", word)
				if err != nil {
					b.replyWordNotFound(&u, word)
					continue
				}
			}

			sp, _ := expositor.GetSpelling()
			d := expositor.GetSenses()[0].GetDefinitions()

			card := Card{
				Word:          word,
				Transcription: sp,
				Definitions:   d,
				Page:          1,
				Total:         len(expositor.GetSenses()),
			}
			t := template.Must(template.New("card").Parse(CardTemplate))
			var out bytes.Buffer
			t.Execute(&out, card)

			msg := tgbotapi.NewMessage(u.Message.Chat.ID, out.String())
			msg.ReplyMarkup = renderButtonsRow(card, word)
			msg.ParseMode = "Markdown"

			s, err := b.botAPI.Send(msg)
			if err != nil {
				b.replyError(u.Message.Chat.ID, u.Message.MessageID, err)
				log.Println("Error in message", s, err)
			}

		case u.CallbackQuery != nil:
			err := b.handleCallbackQuery(&u)
			if err != nil {
				b.replyError(u.CallbackQuery.Message.Chat.ID, u.CallbackQuery.Message.MessageID, err)
			}
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

func (b *Bot) replyError(chatID int64, messageID int, err error) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Some error happened! *%s*", err))
	msg.ReplyToMessageID = messageID
	msg.ParseMode = "Markdown"
	b.botAPI.Send(msg)
}

func (b *Bot) handleCallbackQuery(u *tgbotapi.Update) error {
	query := &ButtonData{}
	err := query.UnmarshalJSON([]byte(u.CallbackQuery.Data))
	if err != nil {
		return err
	}

	var expositor exegete.Expositor
	expositor, ok := b.wordDataProvider.GetWordExpositor(query.Word)
	if !ok {
		expositor, err = b.expositorProvider.GetWordExpositor("en", query.Word)
		if err != nil {
			return err
		}
		b.wordDataProvider.SetWordExpositor(query.Word, expositor)
	}
	if query.AuidoAsked {
		err = b.handleAudioQuery(expositor, u)
		return err
	}
	msg := renderMessageCard(query, expositor)
	editText := tgbotapi.NewEditMessageText(
		u.CallbackQuery.Message.Chat.ID,
		u.CallbackQuery.Message.MessageID,
		msg.Text,
	)
	editText.ParseMode = "Markdown"
	editMarkup := tgbotapi.NewEditMessageReplyMarkup(
		u.CallbackQuery.Message.Chat.ID,
		u.CallbackQuery.Message.MessageID,
		msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup),
	)
	_, err = b.botAPI.Send(editText)
	if err != nil {
		return err
	}
	_, err = b.botAPI.Send(editMarkup)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) handleAudioQuery(expositor exegete.Expositor, u *tgbotapi.Update) error {
	ID, ok := b.wordDataProvider.GetAudioID(expositor.Word())
	log.Printf("HandleAudioQuery: isFilePresent %v\n", ok)
	if !ok {
		audio, err := expositor.GetAudio()
		if err != nil {
			return err
		}
		msg := tgbotapi.NewAudioUpload(u.CallbackQuery.Message.Chat.ID, tgbotapi.FileReader{
			Name:   expositor.Word(),
			Reader: audio,
			Size:   -1,
		})
		response, err := b.botAPI.Send(msg)
		if err != nil {
			return err
		}
		b.wordDataProvider.SetAudioID(expositor.Word(), response.Audio.FileID)
		return nil
	}

	msg := tgbotapi.NewAudioShare(u.CallbackQuery.Message.Chat.ID, ID)
	_, err := b.botAPI.Send(msg)
	return err
}

func renderMessageCard(curQuery *ButtonData, expositor exegete.Expositor) tgbotapi.MessageConfig {
	senses := expositor.GetSenses()
	transc, _ := expositor.GetSpelling()
	log.Println(senses[curQuery.Next-1].GetDefinitions())
	card := Card{
		Word:          expositor.Word(),
		Transcription: transc,
		Definitions:   senses[curQuery.Next-1].GetDefinitions(),
		Page:          curQuery.Next,
		Total:         len(senses),
	}
	t := template.Must(template.New("card").Parse(CardTemplate))
	var out bytes.Buffer
	t.Execute(&out, card)
	msg := tgbotapi.NewMessage(0, out.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = renderButtonsRow(card, expositor.Word())
	return msg
}

func renderButtonsRow(card Card, word string) tgbotapi.InlineKeyboardMarkup {
	buttons := make([]tgbotapi.InlineKeyboardButton, 0, 3)
	if card.Page != 1 {
		data := ButtonData{Word: word, Next: card.Page - 1}
		bytes, _ := data.MarshalJSON()
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("⬅️", string(bytes)))
	}
	if card.Page != card.Total {
		data := ButtonData{Word: word, Next: card.Page + 1}
		bytes, _ := data.MarshalJSON()
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("➡️", string(bytes)))
	}
	data := ButtonData{Word: word, AuidoAsked: true}
	bytes, _ := data.MarshalJSON()
	buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("🔈", string(bytes)))
	return tgbotapi.NewInlineKeyboardMarkup(buttons)
}
