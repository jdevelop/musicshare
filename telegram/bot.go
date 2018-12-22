package telegram

import (
	"github.com/jdevelop/musicshare/music"
	"log"
	"net/url"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramBot struct {
	key  string
	stop chan struct{}
}

func NewTelegramBot(key string) *TelegramBot {
	return &TelegramBot{
		key:  key,
		stop: make(chan struct{}),
	}
}

func (t *TelegramBot) Disconnect() error {
	close(t.stop)
	return nil
}

type servicesCallbacks struct {
	svcType   music.Service
	callbackF func(string, string) (string, *string)
}

func stringPtr(str string) *string {
	return &str
}

var serviceKbds = [...]servicesCallbacks{
	{
		svcType: music.YouTube,
		callbackF: func(svc, id string) (string, *string) {
			return "YouTube", stringPtr(svc + ":yt:" + id)
		},
	},
	{
		svcType: music.Spotify,
		callbackF: func(svc, id string) (string, *string) {
			return "Spotify", stringPtr(svc + ":sp:" + id)
		},
	},
}

func (t *TelegramBot) Connect(resolver *music.ResolverService) error {
	bot, err := tgbotapi.NewBotAPI(t.key)
	if err != nil {
		return nil
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}
	log.Println("Started message loop")
	go func() {
		for {
			select {
			case <-t.stop:
				bot.StopReceivingUpdates()
				return
			case m := <-updates:
				emptyReply := func(err error, id string) {
					if err != nil {
						log.Println("Can't process message: ", err)
					}
					if _, err := bot.AnswerInlineQuery(tgbotapi.InlineConfig{
						InlineQueryID: id,
						Results:       nil,
					}); err != nil {
						log.Printf("Error %#v\n", err)
					}
				}
				switch {
				case m.InlineQuery != nil:
					id, svc := resolver.ResolveServiceAndId(m.InlineQuery.Query)
					switch {
					case id != "":
						name := music.ServiceToHumanName(svc)
						msgs := make([]interface{}, 1)
						reply := tgbotapi.NewInlineQueryResultArticle(id, name, name)
						kbds := make([][]tgbotapi.InlineKeyboardButton, 1)
						svcStr := music.Service2String(svc)
						for _, ks := range serviceKbds {
							if ks.svcType == svc {
								continue
							}
							name, id := ks.callbackF(svcStr, id)
							kbds[0] = append(kbds[0], tgbotapi.InlineKeyboardButton{
								CallbackData: id,
								Text:         name,
							})
						}
						reply.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
							InlineKeyboard: kbds,
						}
						msgs[0] = reply
						if _, err := bot.AnswerInlineQuery(tgbotapi.InlineConfig{
							InlineQueryID: m.InlineQuery.ID,
							Results:       msgs,
						}); err != nil {
							log.Printf("Can't send message: %#v\n", err)
						}
					default:
						log.Printf("Can't understand message: %#v : %#v \n", m.InlineQuery, *m.InlineQuery.From)
						emptyReply(nil, m.InlineQuery.ID)
					}
				case m.CallbackQuery != nil:
					if _, err := bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
						CallbackQueryID: m.CallbackQuery.ID,
					}); err != nil {
						log.Println("Can't answer callback", err)
					}
					link := resolver.ResolveExternalLink(m.CallbackQuery.Data)
					v := url.Values{}
					v.Add("inline_message_id", m.CallbackQuery.InlineMessageID)
					v.Add("text", link)
					if _, err := bot.MakeRequest("editMessageText", v); err != nil {
						log.Println("Can't send response", err)
					}
				default:
					continue
				}
			}
		}
	}()
	return nil
}
