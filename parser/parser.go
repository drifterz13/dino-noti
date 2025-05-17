package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	scraper "github.com/drifterz13/dino-noti/scraper"
)

type BuyeeParser struct{}

func NewBuyeeParser() *BuyeeParser {
	return &BuyeeParser{}
}

func (p *BuyeeParser) Parse(htmlContent string) ([]scraper.Item, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to load HTML for parsing: %w", err)
	}

	var items []scraper.Item

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

		// Only add items that have at least a name and URL
		if name != "" && exists {
			// Handle relative URLs by prepending the base URL if needed
			baseURL := "https://buyee.jp"
			if !strings.HasPrefix(url, "http") {
				url = baseURL + url
			}

			items = append(items, scraper.Item{
				Name:  name,
				Price: price,
				URL:   url,
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
