package service

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/drifterz13/dino-noti/config"
	"github.com/drifterz13/dino-noti/line"
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

	batchSize := 40
	numGoroutines := (len(scrapedItems) + batchSize - 1) / batchSize

	var wg sync.WaitGroup
	resultChan := make(chan []MatchedItem, numGoroutines)
	errorChan := make(chan error, numGoroutines)

	for i := 0; i < len(scrapedItems); i += batchSize {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()

			end := start + batchSize
			if end > len(scrapedItems) {
				end = len(scrapedItems)
			}

			var chunk []string
			for _, item := range scrapedItems[start:end] {
				chunk = append(chunk, item.Name)
			}

			matches, err := llmClient.CheckMatches(chunk, srv.cfg.MyList)
			if err != nil {
				errorChan <- err
				return
			}

			var chunkMatchedItems []MatchedItem
			for _, matchedItem := range matches {
				scrapedItem := findScrapedItemByName(scrapedItems[start:end], matchedItem.OriginalName)
				if scrapedItem != nil {
					chunkMatchedItems = append(chunkMatchedItems, MatchedItem{
						URL:          scrapedItem.URL,
						Price:        scrapedItem.Price,
						OriginalName: matchedItem.OriginalName,
						MatchedName:  matchedItem.MatchedName,
					})
				}
			}

			resultChan <- chunkMatchedItems
		}(i)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	var allMatchedItems []MatchedItem
	for chunkMatchedItems := range resultChan {
		allMatchedItems = append(allMatchedItems, chunkMatchedItems...)
	}

	if len(errorChan) > 0 {
		return nil, <-errorChan // Return the first error encountered
	}

	return allMatchedItems, nil
}

func (srv *Service) HandleLineMessageReq(w http.ResponseWriter, req *http.Request) {
	lineBotClient, err := line.NewLineBotClient(srv.cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating LINE Bot client: %v\n", err)
		return
	}
	events, err := lineBotClient.ParseEvents(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing LINE events: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)

	go func() {
		allScrapedItems, scrapeErrors := srv.ScrapeItems()
		if len(scrapeErrors) > 0 {
			fmt.Fprintf(os.Stderr, "Completed with %d scraping errors.\n", len(scrapeErrors))
		}

		matchedItems, err := srv.FindMatchItems(allScrapedItems)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding matched items: %v\n", err)
			return
		}

		fmt.Printf("Parsed events: %v\n", events)

		replyMessage := generateMessage(matchedItems)
		err = lineBotClient.HandleSendMessage(events, replyMessage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending message: %v\n", err)
		} else {
			fmt.Println("Message sent successfully")
		}
	}()
}

func generateMessage(matchedItems []MatchedItem) string {
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("Cameras on the radar 🦖:\n"))
	for idx, item := range matchedItems {
		msg.WriteString(fmt.Sprintf("%d. (%s yen) %s - %s\n", idx+1, item.Price, item.MatchedName, item.URL))
	}
	return msg.String()
}

func findScrapedItemByName(scrapedItems []scraper.Item, name string) *scraper.Item {
	for _, item := range scrapedItems {
		if item.Name == name {
			return &item
		}
	}
	return nil
}
