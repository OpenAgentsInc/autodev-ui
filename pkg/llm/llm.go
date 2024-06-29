package llm

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// CompletionResponse represents the structure of a completion response
type CompletionResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice represents a single choice in the completion response
type Choice struct {
	Message Message `json:"message"`
}

// Message represents the content of a choice
type Message struct {
	Content string `json:"content"`
}

// LLM represents a Language Model instance
type LLM struct {
	ModelName    string
	APIKey       string
	BaseURL      string
	APIVersion   string
	NumRetries   int
	RetryMinWait time.Duration
	RetryMaxWait time.Duration
}

// NewLLM creates a new LLM instance with default or provided values
func NewLLM(modelName, apiKey, baseURL, apiVersion string, numRetries int, retryMinWait, retryMaxWait time.Duration) *LLM {
	if modelName == "" {
		modelName = "default-model"
	}
	if apiKey == "" {
		apiKey = "dummy-api-key"
	}
	if numRetries == 0 {
		numRetries = 3
	}
	if retryMinWait == 0 {
		retryMinWait = time.Second
	}
	if retryMaxWait == 0 {
		retryMaxWait = 5 * time.Second
	}

	return &LLM{
		ModelName:    modelName,
		APIKey:       apiKey,
		BaseURL:      baseURL,
		APIVersion:   apiVersion,
		NumRetries:   numRetries,
		RetryMinWait: retryMinWait,
		RetryMaxWait: retryMaxWait,
	}
}

// Completion simulates a completion request and returns a dummy response
func (l *LLM) Completion(messages []map[string]string, stop []string) (*CompletionResponse, error) {
	// Simulate API call delay
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	// Generate a dummy response based on the last message
	lastMessage := messages[len(messages)-1]["content"]
	dummyResponse := generateDummyResponse(lastMessage)

	return &CompletionResponse{
		Choices: []Choice{
			{
				Message: Message{
					Content: dummyResponse,
				},
			},
		},
	}, nil
}

func (l *LLM) GenerateResponse(messages []Message, maxTokens int) (string, error) {
	// Implement the logic to call the Anthropic API and generate a response
	// This is a placeholder implementation
	return "This is a placeholder response from the LLM", nil
}

// generateDummyResponse creates a simple response based on the input
func generateDummyResponse(input string) string {
	lowercaseInput := strings.ToLower(input)

	if strings.Contains(lowercaseInput, "hello") || strings.Contains(lowercaseInput, "hi") {
		return "Hello! How can I assist you today?"
	}

	if strings.Contains(lowercaseInput, "goodbye") || strings.Contains(lowercaseInput, "bye") {
		return "Goodbye! Have a great day!"
	}

	if strings.Contains(lowercaseInput, "help") {
		return "I'm here to help! What kind of assistance do you need?"
	}

	if strings.Contains(lowercaseInput, "weather") {
		return "I'm sorry, I don't have real-time weather information. You might want to check a weather website or app for accurate forecasts."
	}

	if strings.Contains(lowercaseInput, "name") {
		return "My name is AutoDev AI. It's nice to meet you!"
	}

	// Default response for any other input
	return "I understand you're saying something about " + input + ". Can you please provide more context or ask a specific question?"
}

func (l *LLM) String() string {
	if l.APIVersion != "" {
		return fmt.Sprintf("LLM(model=%s, api_version=%s, base_url=%s)", l.ModelName, l.APIVersion, l.BaseURL)
	} else if l.BaseURL != "" {
		return fmt.Sprintf("LLM(model=%s, base_url=%s)", l.ModelName, l.BaseURL)
	}
	return fmt.Sprintf("LLM(model=%s)", l.ModelName)
}
