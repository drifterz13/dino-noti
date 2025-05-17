package scraper

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/drifterz13/dino-noti/model"
)

func FetchPage(url string) (string, error) {
	fmt.Printf("Fetching URL: %s\n", url)
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}

type Parser interface {
	Parse(htmlContent string) ([]model.ScrapeItem, error)
}

func ScrapePage(url string, parser Parser) ([]model.ScrapeItem, error) {
	htmlContent, err := FetchPage(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch and parse %s: %w", url, err)
	}
	return parser.Parse(htmlContent)
}
