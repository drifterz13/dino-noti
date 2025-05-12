package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

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

// CheckMatch uses Gemini to check if the item description matches any term in the list.
// It returns the matched term (if any) and a boolean indicating if a match occurred.
func (c *LLMClient) CheckMatch(itemDescription string, searchTerms []string) (string, bool, error) {
	prompt := buildMatchPrompt(itemDescription, searchTerms)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Give LLM some time
	defer cancel()

	fmt.Printf("Sending prompt to LLM for item: %s...\n", itemDescription)
	resp, err := c.client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text(prompt), nil)
	if err != nil {
		return "", false, fmt.Errorf("failed to generate content from LLM: %w", err)
	}

	// Parse the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		fmt.Println("LLM returned no candidates or parts.")
		return "", false, nil // No usable response
	}

	responseText := fmt.Sprint(resp.Candidates[0].Content.Parts[0])
	fmt.Printf("LLM Response: %s\n", responseText) // Log the raw response

	// Simple parsing: Look for a specific phrase like "MATCH: [Term]" or "NO MATCH"
	// This is highly dependent on your prompt design and expected output format.
	// A more robust approach would be to ask the LLM for JSON output.
	// Let's parse for "MATCH:" followed by a potential term.
	if strings.Contains(strings.ToUpper(responseText), "NO MATCH") {
		return "", false, nil
	}

	matchPrefix := "MATCH:"
	if strings.Contains(strings.ToUpper(responseText), matchPrefix) {
		// Attempt to extract the term after "MATCH:"
		parts := strings.SplitN(responseText, matchPrefix, 2)
		if len(parts) == 2 {
			matchedTerm := strings.TrimSpace(parts[1])
			// Optional: Validate if the extracted term is one of the search terms
			// for better reliability, but LLM might return a variation.
			// Simple check for non-empty term:
			if matchedTerm != "" {
				// Further refinement: Sometimes LLM just says "MATCH: Yes".
				// If it says "MATCH: Yes", let's return the original search terms joined as the matched term for logging clarity.
				if strings.EqualFold(matchedTerm, "Yes") {
					matchedTerm = strings.Join(searchTerms, ", ") // Indicate it matched one of these
				}
				return matchedTerm, true, nil
			}
		}
	}

	// If we reach here, the LLM didn't clearly indicate a match in the expected format
	fmt.Println("LLM response did not indicate a clear match.")
	return "", false, nil
}
