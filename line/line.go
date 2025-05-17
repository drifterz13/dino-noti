package line

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"

	"github.com/drifterz13/dino-noti/config"
	"github.com/drifterz13/dino-noti/model"
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

func (c *LineBotClient) HandleSendMessage(events []webhook.EventInterface, items []model.MatchedItem) error {
	for _, event := range events {
		switch e := event.(type) {
		case webhook.MessageEvent:
			switch message := e.Message.(type) {
			case webhook.StickerMessageContent:
				fmt.Println("Sending flex message...")

				var flexBubbles []*messaging_api.FlexBubble
				for _, item := range items {
					flexBubble := BuildFlexBubbleContainer(item.MatchedName, item.Price, item.ImageURL, item.URL)
					flexBubbles = append(flexBubbles, flexBubble)
				}
				carousel := BuildCarouselFlexMessage(flexBubbles)
				if err := c.SendFlexMessages(e.ReplyToken, *carousel); err != nil {
					return err
				}
			case webhook.TextMessageContent:
				replyMessage := generateMessage(items)
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

func (c *LineBotClient) SendFlexMessages(replyToken string, flexMessage messaging_api.FlexMessage) error {
	if _, err := c.Bot.ReplyMessage(
		&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				&messaging_api.FlexMessage{
					AltText:  "Cameras on radar 🦖",
					Contents: flexMessage.Contents,
				},
			},
		},
	); err != nil {
		return fmt.Errorf("Failed to reply message: %v", err)
	}

	return nil
}

func generateMessage(matchedItems []model.MatchedItem) string {
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("Cameras on the radar 🦖:\n"))
	for idx, item := range matchedItems {
		msg.WriteString(fmt.Sprintf("%d. (%s yen) %s - %s\n", idx+1, item.Price, item.MatchedName, item.URL))
	}
	return msg.String()
}
