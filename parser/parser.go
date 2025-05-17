package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/drifterz13/dino-noti/model"
)

type BuyeeParser struct{}

func NewBuyeeParser() *BuyeeParser {
	return &BuyeeParser{}
}

func (p *BuyeeParser) Parse(htmlContent string) ([]model.ScrapeItem, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to load HTML for parsing: %w", err)
	}

	var items []model.ScrapeItem

	itemSelector := ".itemCard"

	doc.Find(itemSelector).Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find(".itemCard__itemName a").Text())

		url, exists := s.Find(".itemCard__itemName a").Attr("href")
		if !exists {
			// Try another selector if the first one doesn't work
			url, exists = s.Find(".g-thumbnail__outer a").Attr("href")
		}

		// Find the current price - first price in the list
		priceText := s.Find(".g-priceDetails__item .g-price").First().Text()
		// Clean up the price (remove "yen" and trim spaces)
		price := strings.TrimSpace(strings.Replace(priceText, "yen", "", -1))

		// Find the image URL
		imageURL, exists := s.Find(".g-thumbnail__image").Attr("data-src")
		if exists {
			fileExt := ".jpg"
			fileExtIdx := strings.Index(imageURL, fileExt)

			if fileExtIdx != -1 {
				imageURL = imageURL[:fileExtIdx+len(fileExt)]
			}
		}

		// Only add items that have at least a name and URL
		if name != "" && exists {
			// Handle relative URLs by prepending the base URL if needed
			baseURL := "https://buyee.jp"
			if !strings.HasPrefix(url, "http") {
				url = baseURL + url
			}

			items = append(items, model.ScrapeItem{
				Name:     name,
				Price:    price,
				URL:      url,
				ImageURL: imageURL,
			})
		}
	})

	if len(items) == 0 {
		fmt.Println("Warning: No items found with selector:", itemSelector)
	} else {
		fmt.Printf("Found %d items on the page.\n", len(items))
	}

	return items, nil
}
