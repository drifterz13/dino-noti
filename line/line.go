package line

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"

	"github.com/drifterz13/dino-noti/config"
)

type LineBotClient struct {
	Bot *messaging_api.MessagingApiAPI
	Cfg *config.Config
}

func NewLineBotClient(cfg *config.Config) (*LineBotClient, error) {
	bot, err := messaging_api.NewMessagingApiAPI(
		cfg.LineChannelToken,
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to create LINE Bot client: %v", err)
	}

	return &LineBotClient{
		Bot: bot,
		Cfg: cfg,
	}, nil
}

func (c *LineBotClient) ParseEvents(req *http.Request) ([]webhook.EventInterface, error) {
	if req.Method != http.MethodPost {
		return nil, errors.New("Invalid request method")
	}

	cb, err := webhook.ParseRequest(c.Cfg.LineChannelSecret, req)

	if err != nil {
		if errors.Is(err, linebot.ErrInvalidSignature) {
			return nil, fmt.Errorf("Invalid signature: %v", err)
		}

		return nil, fmt.Errorf("Failed to parse request: %v", err)
	}

	return cb.Events, nil
}

func (c *LineBotClient) HandleSendMessage(events []webhook.EventInterface, replyMessage string) error {
	for _, event := range events {
		switch e := event.(type) {
		case webhook.MessageEvent:
			switch message := e.Message.(type) {
			case webhook.StickerMessageContent:
			case webhook.TextMessageContent:
				c.SendMessage(e.ReplyToken, replyMessage)
			default:
				return fmt.Errorf("Unsupported message type: %T\n", message)
			}
		default:
			return fmt.Errorf("Unsupported event type: %T\n", e)
		}
	}

	return nil
}

func (c *LineBotClient) SendMessage(replyToken string, replyMessage string) error {
	if _, err := c.Bot.ReplyMessage(
		&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				messaging_api.TextMessage{
					Text: replyMessage,
				},
			},
		},
	); err != nil {
		return fmt.Errorf("Failed to reply message: %v", err)
	}

	return nil
}
