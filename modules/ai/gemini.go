package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/masudur-rahman/expense-tracker-bot/configs"

	"google.golang.org/genai"
)

const (
	Gemini25FlashLite   = "gemini-2.5-flash-lite"
	Gemini20FlashLite   = "gemini-2.0-flash-lite"
	Gemini25Flash       = "gemini-2.5-flash"
	Gemini3FlashPreview = "gemini-3-flash-preview"
	Gemini35Flash       = "gemini-3.5-flash"
	Gemini31FlashLite   = "gemini-3.1-flash-lite"
)

func TxnSubcategoryClassifier(ctx context.Context, apiKey, userInput, taxonomyJSON string, model ...string) (*ClassificationResult, error) {
	classifier := Gemini31FlashLite
	if configured := configs.TrackerConfig.System.GeminiModel; configured != "" {
		classifier = configured
	}
	if len(model) > 0 {
		classifier = model[0]
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`You are a personal expense classification system for a Bangladeshi user.
Each subcategory has a "Hint" field (detailed examples) and a "Keywords" field (short common terms) — use both to match the input.
Pick the subcategory whose Hint/Keywords best match the input. Only fall back to misc-misc if nothing else fits.

Constraints:
1. Output must be valid JSON matching the Schema.
2. The selected SubcategoryID must exist under the selected CategoryID in the Taxonomy.
3. Match against Hint and Keywords first, then Name, then general reasoning.
4. Use ONLY the exact subcategory IDs from the Taxonomy. Never invent new IDs.
5. Identify the "intent" of the transaction using these rules:
   - "transfer": money moving between YOUR OWN accounts/wallets — ATM withdrawal (bank→cash), bank deposit (cash→bank), mobile banking transfer (bkash/nagad/rocket), send money between own accounts, cash out, cash in.
   - "income": money entering your possession — salary, bonus, interest received, money received from others, loan received.
   - "expense": money leaving for goods/services/bills — food, transport, shopping, utilities, rent, loan repayment, fees.
   When in doubt: if the user is moving money between their own accounts, it is "transfer".
6. Person-to-person debt has four subcategories — pick by WHO acts and WHICH WAY money flows:
   - fin-lend (expense): YOU give someone a new loan — "lent/gave/handed to X".
   - fin-recover (income): someone returns money YOU lent — "X returned/paid me back", "got back from X", "collected from X".
   - fin-borrow (income): YOU take a new loan — "borrowed/took from X".
   - fin-return (expense): YOU pay back money you borrowed — "returned/repaid to X", "paid back", "dhar shodh".
   The SUBJECT decides direction: "I returned" is fin-return, but "John returned" / "friend paid me back" is fin-recover (money comes to you).

Taxonomy:
%s

User Input: "%s"
`, taxonomyJSON, userInput)

	resp, err := client.Models.GenerateContent(ctx, classifier, genai.Text(prompt), &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"intent":         {Type: genai.TypeString, Enum: []string{"income", "expense", "transfer"}},
				"category_id":    {Type: genai.TypeString},
				"subcategory_id": {Type: genai.TypeString},
				"confidence":     {Type: genai.TypeNumber},
			},
			Required: []string{"intent", "category_id", "subcategory_id", "confidence"},
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
