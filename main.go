package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/drifterz13/dino-noti/config"
	"github.com/drifterz13/dino-noti/db"
	"github.com/drifterz13/dino-noti/llm"
	"github.com/drifterz13/dino-noti/parser"
	"github.com/drifterz13/dino-noti/scraper"

	// Add this if you need to listen on a port for Cloud Run execution
	// For a job-like execution triggered by Cloud Scheduler, listening might not be needed,
	// but it's standard for Cloud Run services.
	"os"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	database, err := db.InitDB(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.CloseDB(database)

	llmClient, err := llm.NewLLMClient(cfg.GeminiAPIKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing LLM client: %v\n", err)
		os.Exit(1)
	}

	// --- Main Logic for Scraping and Processing ---
	fmt.Printf("Starting scrape for %s up to page %d...\n", cfg.TargetURL, cfg.MaxPages)

	// Use your specific parser implementation
	ps := parser.NewBuyeeParser()

	// We'll collect all scraped items first, then process with LLM
	var allScrapedItems []scraper.Item
	scrapeErrors := []error{}

	for pageNum := 1; pageNum <= cfg.MaxPages; pageNum++ {
		pageURL := fmt.Sprintf("%s?page=%d", cfg.TargetURL, pageNum)

		itemsOnPage, err := scraper.ScrapePage(pageURL, ps)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scraping page %d (%s): %v\n", pageNum, pageURL, err)
			scrapeErrors = append(scrapeErrors, err)
			// Decide if you want to continue or stop on error
			continue
		}
		allScrapedItems = append(allScrapedItems, itemsOnPage...)

		// Optional: Add a delay between page fetches to be polite
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("Finished scraping. Found a total of %d items.\n", len(allScrapedItems))

	// --- LLM Matching and Saving ---
	fmt.Printf("Starting LLM matching for %d items...\n", len(allScrapedItems))

	// Process items concurrently with LLM calls
	var wg sync.WaitGroup
	itemChan := make(chan scraper.Item, len(allScrapedItems)) // Buffered channel

	// Worker pool for LLM calls
	numWorkers := 5 // Limit concurrent LLM calls to avoid hitting rate limits or overwhelming resources
	if numWorkers > len(allScrapedItems) {
		numWorkers = len(allScrapedItems)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range itemChan {
				matchedTerm, isMatch, llmErr := llmClient.CheckMatch(item.Name, cfg.MyList) // Use item.Name as description for now
				if llmErr != nil {
					fmt.Fprintf(os.Stderr, "LLM error for item '%s': %v\n", item.Name, llmErr)
					// Decide how to handle LLM errors - skip item, retry, etc.
					continue
				}

				if isMatch {
					fmt.Printf("Item '%s' matched search term(s): %s\n", item.Name, matchedTerm)
					dbItem := db.Item{
						URL:         item.URL,
						Name:        item.Name,
						Price:       item.Price,
						MatchedTerm: matchedTerm, // Store the term LLM said it matched
						Timestamp:   time.Now(),
					}
					saveErr := db.SaveItem(database, dbItem)
					if saveErr != nil {
						fmt.Fprintf(os.Stderr, "Error saving item '%s' to DB: %v\n", item.Name, saveErr)
					}
				} else {
					fmt.Printf("Item '%s' did not match.\n", item.Name)
				}

				// Add a small delay between LLM calls too
				time.Sleep(3 * time.Second)
			}
		}()
	}

	// Send items to the channel
	for _, item := range allScrapedItems {
		itemChan <- item
	}
	close(itemChan) // Close the channel to signal workers no more items are coming

	wg.Wait() // Wait for all workers to finish

	fmt.Println("LLM matching and saving complete.")

	if len(scrapeErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Completed with %d scraping errors.\n", len(scrapeErrors))
	}
	fmt.Println("Application finished.")

	// --- Cloud Run Service Listener (Optional for Job, standard for Service) ---
	// If you trigger this via HTTP (e.g., Cloud Scheduler targeting a Cloud Run service),
	// your application needs to listen on a port.
	// If triggered as a Cloud Run JOB, the app just runs and exits.
	// For a service, add this:
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for Cloud Run
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// You could re-run the logic here, or just return a success message
		// Running the whole scraping logic on every HTTP request is probably NOT
		// what you want for a scheduled job. A better pattern is:
		// 1. HTTP request arrives (e.g., from Cloud Scheduler)
		// 2. Trigger the scraping/processing logic (perhaps in a goroutine, though
		//    Cloud Run expects the request handler to finish relatively quickly).
		//    A Cloud Run JOB is a better fit for a long-running process like this.
		// For simplicity in this example, let's assume this is for a JOB or
		// the main logic runs and exits, and this handler is just for pinging.
		fmt.Fprintln(w, "Scraping and matching process started/completed.")
	})

	// If running as a long-lived service (less likely for a scheduled scrape job)
	// fmt.Printf("Listening on port %s\n", port)
	// if err := http.ListenAndServe(":"+port, nil); err != nil && err != http.ErrServerClosed {
	// 	fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
	// 	os.Exit(1)
	// }
	// For a JOB, the main function simply finishes.
}
