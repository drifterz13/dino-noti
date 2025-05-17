package service

import (
	"fmt"
	"os"
	"time"

	"github.com/drifterz13/dino-noti/config"
	"github.com/drifterz13/dino-noti/llm"
	"github.com/drifterz13/dino-noti/parser"
	"github.com/drifterz13/dino-noti/scraper"
)

type Service struct {
	cfg *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (srv *Service) ScrapeItems() ([]scraper.Item, []error) {
	fmt.Printf("Starting scrape for %s up to page %d...\n", srv.cfg.TargetURL, srv.cfg.MaxPages)

	ps := parser.NewBuyeeParser()

	var allScrapedItems []scraper.Item
	scrapeErrors := []error{}

	for pageNum := 1; pageNum <= srv.cfg.MaxPages; pageNum++ {
		pageURL := fmt.Sprintf("%s&page=%d", srv.cfg.TargetURL, pageNum)

		itemsOnPage, err := scraper.ScrapePage(pageURL, ps)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scraping page %d (%s): %v\n", pageNum, pageURL, err)
			scrapeErrors = append(scrapeErrors, err)
			continue
		}
		allScrapedItems = append(allScrapedItems, itemsOnPage...)
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("Finished scraping. Found a total of %d items.\n", len(allScrapedItems))

	return allScrapedItems, scrapeErrors
}

type MatchedItem struct {
	URL          string
	OriginalName string
	MatchedName  string
	Price        string
}

func (srv *Service) FindMatchItems(scrapedItems []scraper.Item) ([]MatchedItem, error) {
	llmClient, err := llm.NewLLMClient(srv.cfg.GeminiAPIKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing LLM client: %v\n", err)
		os.Exit(1)
	}

	var matchedItems []MatchedItem
	batchSize := 40

	for i := 0; i < len(scrapedItems); i += batchSize {
		end := i + batchSize
		if end > len(scrapedItems) {
			end = len(scrapedItems)
		}

		var chunk []string
		for _, item := range scrapedItems[i:end] {
			chunk = append(chunk, item.Name)
		}

		matches, err := llmClient.CheckMatches(chunk, srv.cfg.MyList)

		if err != nil {
			return matchedItems, err
		}

		if len(matches) > 0 {
			for _, matchedItem := range matches {
				scrapedItem := findScrapedItemByName(scrapedItems, matchedItem.OriginalName)
				if scrapedItem != nil {
					matchedItems = append(matchedItems, MatchedItem{
						URL:          scrapedItem.URL,
						Price:        scrapedItem.Price,
						OriginalName: matchedItem.OriginalName,
						MatchedName:  matchedItem.MatchedName,
					})
				}
			}
		}
	}

	return matchedItems, nil
}

func findScrapedItemByName(scrapedItems []scraper.Item, name string) *scraper.Item {
	for _, item := range scrapedItems {
		if item.Name == name {
			return &item
		}
	}
	return nil
}
