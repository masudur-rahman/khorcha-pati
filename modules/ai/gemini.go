package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

const (
	Gemini25FlashLite   = "gemini-2.5-flash-lite"
	Gemini20FlashLite   = "gemini-2.0-flash-lite"
	Gemini25Flash       = "gemini-2.5-flash"
	Gemini3FlashPreview = "gemini-3-flash-preview"
)

func TxnSubcategoryClassifier(ctx context.Context, apiKey, userInput, taxonomyJSON string, model ...string) (*ClassificationResult, error) {
	classifier := Gemini25FlashLite
	if len(model) > 0 {
		classifier = model[0]
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`
    Analyze the User Input and map it to the most relevant Category and Subcategory from the provided Taxonomy.

    Constraints:
    1. Output must be valid JSON matching the Schema.
    2. The selected SubcategoryID **must** exist under the selected CategoryID in the Taxonomy.
    3. If the input is ambiguous, choose the most logical option.

    Taxonomy (Strict Options):
    %s

    User Input: "%s"
`, taxonomyJSON, userInput)

	resp, err := client.Models.GenerateContent(ctx, classifier, genai.Text(prompt), &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"category_id":    {Type: genai.TypeString},
				"subcategory_id": {Type: genai.TypeString},
				"confidence":     {Type: genai.TypeNumber},
			},
			Required: []string{"category_id", "subcategory_id", "confidence"},
		},
	})
	if err != nil {
		return nil, err
	}

	var result = &ClassificationResult{}
	if err = json.Unmarshal([]byte(resp.Candidates[0].Content.Parts[0].Text), &result); err != nil {
		return nil, err
	}

	return result, err
}
