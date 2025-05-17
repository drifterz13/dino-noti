package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/drifterz13/dino-noti/config"
	"github.com/drifterz13/dino-noti/db"
	"github.com/drifterz13/dino-noti/service"

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

	srv := service.NewService(cfg)
	allScrapedItems, scrapeErrors := srv.ScrapeItems()

	matchedItems, err := srv.FindMatchItems(allScrapedItems)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding matched items: %v\n", err)
	}

	for _, item := range matchedItems {
		matchedItem := db.Item{
			URL:         item.URL,
			Name:        item.OriginalName,
			Price:       item.Price,
			MatchedTerm: item.MatchedName,
			Timestamp:   time.Now(),
		}
		fmt.Printf("Saving matched item: %+v\n", matchedItem)

		err := db.SaveItem(database, matchedItem)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving item to database: %v\n", err)
		}
	}

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
