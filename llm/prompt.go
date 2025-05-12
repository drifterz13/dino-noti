package llm

import (
	"fmt"
	"strings"
)

// buildMatchPrompt constructs the prompt for the LLM.
// Experiment heavily with this prompt for best results!
func buildMatchPrompt(itemDescription string, searchTerms []string) string {
	// Example prompt structure:
	// - Role: Assistant
	// - Task: Compare item description to a list of desired items.
	// - Input: Item description, list of desired items.
	// - Output: Clear YES/NO answer, potentially indicating which term matched.
	// - Format instruction: Use specific keywords like "MATCH:" or "NO MATCH".

	prompt := fmt.Sprintf(`
You are an assistant designed to match product descriptions to a list of desired items.
Analyze the following product description and determine if it matches any item in the provided list.
The item description might include extra details, specifications, or marketing text. Focus on identifying the core product.

Product Description:
"%s"

List of Desired Items:
- %s

Instructions:
1. If the product description clearly matches one or more items in the "List of Desired Items", respond *only* with the format: "MATCH: [The term from the list that matched or a summary if multiple]".
2. If the product description does NOT match any item in the "List of Desired Items", respond *only* with the format: "NO MATCH".
3. Be concise. Do not include explanations or conversational text, just the required format.

Examples:
Product Description: "Canon キヤノン PowerShot A4000 IS コンパクトデジ"
List of Desired Items: - Canon Powershot A4000
Response: MATCH: Canon Powershot A4000

Product Description: "動作確認済み】ACアダプター CASIO カシオ デジカ"
List of Desired Items: - Casio Exilim EX-ZR100
Response: NO MATCH

Product Description: "205 ★稼働品★Canon キャノン IXY 110F コンパク"
List of Desired Items: - Canon IXY 110F
Response: MATCH: Canon IXY 110F

Now, analyze the following:
Product Description: "%s"
List of Desired Items:
- %s

Your Response:`,
		itemDescription,
		strings.Join(searchTerms, ",\n- "), // Format list nicely
		itemDescription,                    // Repeat for structure
		strings.Join(searchTerms, ",\n- "), // Repeat for structure
	)

	return strings.TrimSpace(prompt) // Trim leading/trailing whitespace
}
