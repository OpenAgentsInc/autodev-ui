package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/openagentsinc/autodev/pkg/llm"
)

type Config struct {
	GreptileApiKey  string
	GithubToken     string
	AnthropicAPIKey string
	LLM             llm.LLM
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file")
	}

	config := &Config{
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		GreptileApiKey:  os.Getenv("GREPTILE_API_KEY"),
		GithubToken:     os.Getenv("GITHUB_TOKEN"),
	}

	if config.GreptileApiKey == "" || config.GithubToken == "" {
		return nil, fmt.Errorf("GREPTILE_API_KEY and GITHUB_TOKEN must be set")
	}

	return config, nil
}
