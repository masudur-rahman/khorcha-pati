package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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
	DeepSeekChatFree     OpenRouterModel = "deepseek/deepseek-chat:free"
	DeepSeekR1Free       OpenRouterModel = "deepseek/deepseek-r1-distill-qwen-1.5b:free"
	DeepSeekR10525Free   OpenRouterModel = "deepseek/deepseek-r1-0528:free"
	Gemini20FlashFree    OpenRouterModel = "google/gemini-2.0-flash-exp:free"
	Gemma34bITFree       OpenRouterModel = "google/gemma-3-4b-it:free"
	NVDIANemotron30bFree OpenRouterModel = "nvidia/nemotron-3-nano-30b-a3b:free" // Working model
	NVDIANemotron9bFree  OpenRouterModel = "nvidia/nemotron-nano-9b-v2:free"
	NVDIANemotron12bFree OpenRouterModel = "nvidia/nemotron-nano-12b-v2-vl:free"
	DeepSeekV31NexN1Free OpenRouterModel = "nex-agi/deepseek-v3.1-nex-n1:free"
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

// Helper: Build the classification prompt
func (c *Client) buildPrompt(taxonomyJSON, userInput string) []map[string]any {
	return []map[string]any{
		{"role": "system", "content": `You are a classification system. Always respond with valid JSON matching this exact schema:
{
	"category_id": "string", # required, must be parent of subcategory
	"subcategory_id": "string", # required, must match one subcategory id from the Taxonomy
	"confidence": number
}
Only return the JSON object, no other text.`},
		{"role": "user", "content": fmt.Sprintf("Taxonomy:\n%s\n\nUser Input: \"%s\"\n\nClassify the input into the taxonomy.", taxonomyJSON, userInput)},
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

	result, err := c.query(ctx, NVDIANemotron30bFree, messages)
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
