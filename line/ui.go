package line

import (
	"fmt"

	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

func BuildCarouselFlexMessage(bubbles []*messaging_api.FlexBubble) *messaging_api.FlexMessage {
	contents := make([]messaging_api.FlexBubble, len(bubbles))
	for i, bubble := range bubbles {
		contents[i] = *bubble
	}

	carousel := &messaging_api.FlexCarousel{
		Contents: contents,
	}

	return &messaging_api.FlexMessage{
		Contents: carousel,
	}
}

func BuildFlexBubbleContainer(title, price, imageURL, productURL string) *messaging_api.FlexBubble {
	bubble := &messaging_api.FlexBubble{
		Hero: &messaging_api.FlexImage{
			Url:         imageURL,
			Size:        "full",
			AspectRatio: "20:13",
			AspectMode:  "cover",
			Action:      messaging_api.UriAction{Uri: productURL, Label: "View Product"},
		},
		Body: &messaging_api.FlexBox{
			Layout:  messaging_api.FlexBoxLAYOUT_VERTICAL,
			Spacing: "md",
			Contents: []messaging_api.FlexComponentInterface{
				&messaging_api.FlexText{
					Text: title,
					Size: string(messaging_api.FlexTextFontSize_MD),
					Wrap: true,
				},
				&messaging_api.FlexText{
					Text:   fmt.Sprintf("%s JP¥", price),
					Size:   string(messaging_api.FlexTextFontSize_LG),
					Weight: messaging_api.FlexTextWEIGHT_BOLD,
					Wrap:   true,
				},
			},
		},
		Styles: &messaging_api.FlexBubbleStyles{
			Body: &messaging_api.FlexBlockStyle{
				BackgroundColor: "#ffffff",
			},
		},
	}

	return bubble
}
