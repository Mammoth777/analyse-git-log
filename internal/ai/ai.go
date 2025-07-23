package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"git-log-analyzer/internal/analyzer"
)

// AIConfig contains configuration for AI analysis
type AIConfig struct {
	APIEndpoint string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
}

// AIClient handles communication with AI models
type AIClient struct {
	config AIConfig
	client *http.Client
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the request payload for chat API
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

// ChatResponse represents the response from chat API
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// NewAIClient creates a new AI client from environment variables
func NewAIClient() (*AIClient, error) {
	config := AIConfig{
		APIEndpoint: getEnv("AI_API_ENDPOINT", "https://api.openai.com/v1/chat/completions"),
		APIKey:      getEnv("AI_API_KEY", ""),
		Model:       getEnv("AI_MODEL", "gpt-3.5-turbo"),
		MaxTokens:   getEnvInt("AI_MAX_TOKENS", 2000),
		Temperature: getEnvFloat("AI_TEMPERATURE", 0.7),
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("AI_API_KEY environment variable is required")
	}

	return &AIClient{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// AnalyzeWithAI performs AI-powered analysis of git statistics
func (c *AIClient) AnalyzeWithAI(stats *analyzer.Statistics, basicReport string) (string, error) {
	prompt := c.buildAnalysisPrompt(stats, basicReport)
	
	return c.sendChatRequest(prompt)
}

// buildAnalysisPrompt creates a prompt for AI analysis
func (c *AIClient) buildAnalysisPrompt(stats *analyzer.Statistics, basicReport string) string {
	prompt := fmt.Sprintf(`Please analyze the following Git repository statistics and provide insights:

BASIC STATISTICS:
%s

DETAILED DATA:
- Total commits: %d
- Number of contributors: %d
- Active period: %s to %s
- Active days: %d

Please provide:
1. Development pattern analysis
2. Team collaboration insights
3. Code quality observations
4. Productivity trends
5. Recommendations for improvement

Focus on actionable insights that can help improve the development process.`,
		basicReport,
		stats.TotalCommits,
		len(stats.AuthorStats),
		stats.TimeStats.FirstCommit.Format("2006-01-02"),
		stats.TimeStats.LastCommit.Format("2006-01-02"),
		stats.TimeStats.ActiveDays)

	return prompt
}

// sendChatRequest sends a request to the AI chat API
func (c *AIClient) sendChatRequest(prompt string) (string, error) {
	request := ChatRequest{
		Model: c.config.Model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are an expert software development analyst. Analyze git repository data and provide actionable insights.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", c.config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as integer with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvFloat gets environment variable as float with default value
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var floatValue float64
		if _, err := fmt.Sscanf(value, "%f", &floatValue); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
