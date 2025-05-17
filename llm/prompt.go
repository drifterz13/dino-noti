package llm

import (
	"fmt"
	"strings"
)

func buildProductNames(itemDescriptions []string) string {
	formattedDescriptions := formatItemDescriptions(itemDescriptions)

	prompt := fmt.Sprintf(`
You are an assistant designed to extract the brand and model specifically focusing on digital compact cameras.
Analyze the following product descriptions and output the brand and model of the digital compact camera along with 
the provided item description.
The item descriptions might include extra details, specifications, or marketing text. Focus on identifying brand and model for Canon, Nikon, Sony, Fuji, Casio, and Panasonic digital compact cameras.

Instructions:
1. For each product description that's a digital compact camera, respond with the format: "[Item description]: [Brand] [Model]".
2. If a product description cannot be identified as a digital compact camera, skip it.
3. Be concise. Do not include explanations or conversational text, just the required format.
4. Ensure that you extract and return the brand and model name of the digital compact camera

Examples:
Product Descriptions:
1. Canon キヤノン PowerShot A4000 IS コンパクトデジ
2. 動作確認済み】ACアダプター CASIO カシオ デジカ
3. VANGUARD◆デジタルカメラその他/VEO3T+234A
4. 205 ★稼働品★Canon キャノン IXY 110F コンパク


Response:
Canon キヤノン PowerShot A4000 IS コンパクトデジ: Canon PowerShot A4000
205 ★稼働品★Canon キャノン IXY 110F コンパク: Canon IXY 110F

Now, analyze the following:
Product Descriptions:
%s

Your Response:`,
		formattedDescriptions,
	)

	return strings.TrimSpace(prompt)
}

func formatItemDescriptions(descriptions []string) string {
	var formatted []string
	for i, desc := range descriptions {
		formatted = append(formatted, fmt.Sprintf("%d. %s", i+1, desc))
	}
	return strings.Join(formatted, "\n")
}
