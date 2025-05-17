package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drifterz13/dino-noti/matcher"
	"github.com/drifterz13/dino-noti/model"
	"google.golang.org/genai"
)

type LLMClient struct {
	client *genai.Client
}

func NewLLMClient(apiKey string) (*LLMClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &LLMClient{client: client}, nil
}

func (c *LLMClient) CheckMatches(itemDescriptions []string, searchTerms []string) ([]model.MatchedItem, error) {
	var matchedItems []model.MatchedItem

	prompt := buildProductNames(itemDescriptions)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text(prompt), nil)
	if err != nil {
		return matchedItems, fmt.Errorf("failed to generate content from LLM: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		fmt.Println("LLM returned no candidates or parts.")
		return matchedItems, nil
	}

	responseText := strings.TrimSpace(resp.Candidates[0].Content.Parts[0].Text)

	for i, line := range strings.Split(responseText, "\n") {
		splittedStr := strings.Split(line, ":")
		originalName := strings.TrimSpace(splittedStr[0])
		responseName := strings.TrimSpace(splittedStr[1])

		matched, itemName := matcher.MatchItem(responseName, searchTerms)
		if matched {
			item := model.MatchedItem{
				Index:        i + 1,
				OriginalName: originalName,
				MatchedName:  itemName,
			}
			matchedItems = append(matchedItems, item)
			fmt.Printf("Matched Item: %s -> %s\n", originalName, itemName)
		}
	}

	return matchedItems, nil
}
