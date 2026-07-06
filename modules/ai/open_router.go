package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/configs"
)

// Client for OpenRouter API
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
	referer string
	appName string
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// Model constants

type OpenRouterModel string

const (
	DeepSeekChatFree              OpenRouterModel = "deepseek/deepseek-chat:free"
	DeepSeekR1Free                OpenRouterModel = "deepseek/deepseek-r1-distill-qwen-1.5b:free"
	DeepSeekR10525Free            OpenRouterModel = "deepseek/deepseek-r1-0528:free"
	Gemini20FlashFree             OpenRouterModel = "google/gemini-2.0-flash-exp:free"
	Gemma34bITFree                OpenRouterModel = "google/gemma-3-4b-it:free"
	NVDIANemotron30bFree          OpenRouterModel = "nvidia/nemotron-3-nano-30b-a3b:free" // Working model
	NVDIANemotron9bFree           OpenRouterModel = "nvidia/nemotron-nano-9b-v2:free"
	NVDIANemotron12bFree          OpenRouterModel = "nvidia/nemotron-nano-12b-v2-vl:free"
	DeepSeekV31NexN1Free          OpenRouterModel = "nex-agi/deepseek-v3.1-nex-n1:free"
	StepFun35FlashFree            OpenRouterModel = "stepfun/step-3.5-flash:free"
	NVDIANemotronUltra55bFree     OpenRouterModel = "nvidia/nemotron-3-ultra-550b-a55b:free"
	NVDIANLLamaNemotronRerankFree OpenRouterModel = "nvidia/llama-nemotron-rerank-vl-1b-v2:free"
	NVIDIANemotronSuper120bFree   OpenRouterModel = "nvidia/nemotron-3-super-120b-a12b:free"
	PoolsideLagunaM1Free          OpenRouterModel = "poolside/laguna-m.1:free"
	OpenAIGPTOSS120bFree          OpenRouterModel = "openai/gpt-oss-120b:free"
	CohereNorthMiniCodeFree       OpenRouterModel = "cohere/north-mini-code:free"
	//ByteDanceSeedream             OpenRouterModel = "bytedance-seed/seedream-4.5"
)

// NewClient creates an OpenRouter client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		client:  &http.Client{Timeout: 30 * time.Second},
		referer: "https://t.me/",
		appName: "TelegramBot",
	}
}

// buildPrompt constructs the classification prompt with hint-aware instructions.
func (c *Client) buildPrompt(taxonomyJSON, userInput string) []map[string]any {
	return []map[string]any{
		{"role": "system", "content": `You are a personal expense classification system for a Bangladeshi user.
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

Always respond with valid JSON matching this exact schema:
{
	"intent": "income" | "expense" | "transfer",
	"category_id": "string",
	"subcategory_id": "string",
	"confidence": number
}
Only return the JSON object, no other text.`},
		{"role": "user", "content": fmt.Sprintf("Taxonomy:\n%s\n\nUser Input: \"%s\"\n\nClassify this into the best matching subcategory using the Hint and Keywords, and determine the intent.", taxonomyJSON, userInput)},
	}
}

// Helper: Clean JSON response from markdown
func (c *Client) cleanJSON(resp string) string {
	clean := strings.TrimSpace(resp)
	if strings.HasPrefix(clean, "```json") {
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimSuffix(clean, "```")
	}
	return strings.TrimSpace(clean)
}

// Main function to query the API
func (c *Client) query(ctx context.Context, model OpenRouterModel, messages []map[string]any) (string, error) {
	reqBody := map[string]any{
		"model":    model,
		"messages": messages,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", c.referer)
	req.Header.Set("X-Title", c.appName)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// TxnSubcategoryClassifier Classify user input into taxonomy
func (c *Client) TxnSubcategoryClassifier(ctx context.Context, userInput string, taxonomyJSON string) (*ClassificationResult, error) {
	messages := c.buildPrompt(taxonomyJSON, userInput)

	model := NVDIANemotron30bFree
	if configured := configs.TrackerConfig.System.OpenRouterModel; configured != "" {
		model = OpenRouterModel(configured)
	}

	result, err := c.query(ctx, model, messages)
	if err != nil {
		return nil, err
	}

	cleanResult := c.cleanJSON(result)
	var classification ClassificationResult
	if err := json.Unmarshal([]byte(cleanResult), &classification); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w\nResponse: %s", err, cleanResult)
	}

	return &classification, nil
}
