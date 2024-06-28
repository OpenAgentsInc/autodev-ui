package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	AnthropicAPIURL     = "https://api.anthropic.com/v1/messages"
	AnthropicAPIVersion = "2023-06-01"
	DefaultModel        = "claude-3-5-sonnet-20240620"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type AnthropicResponse struct {
	Content string `json:"content"`
}

type LLM struct {
	APIKey string
	Model  string
}

func NewLLM(apiKey string, model string) *LLM {
	if model == "" {
		model = DefaultModel
	}
	return &LLM{
		APIKey: apiKey,
		Model:  model,
	}
}

func (l *LLM) GenerateResponse(messages []Message, maxTokens int) (string, error) {
	if l.APIKey == "" {
		l.APIKey = os.Getenv("ANTHROPIC_API_KEY")
		if l.APIKey == "" {
			return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
		}
	}

	requestBody := AnthropicRequest{
		Model:     l.Model,
		MaxTokens: maxTokens,
		Messages:  messages,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %v", err)
	}

	req, err := http.NewRequest("POST", AnthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("x-api-key", l.APIKey)
	req.Header.Set("anthropic-version", AnthropicAPIVersion)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var anthropicResp AnthropicResponse
	err = json.Unmarshal(body, &anthropicResp)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}

	return anthropicResp.Content, nil
}
