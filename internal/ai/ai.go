package ai

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"git-log-analyzer/internal/analyzer"
	"git-log-analyzer/internal/i18n"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

// AIConfig contains configuration for AI analysis
type AIConfig struct {
	APIEndpoint string
	APIKey      string
	Model       string
	MaxTokens   int64
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
	MaxTokens   int64         `json:"max_tokens,omitempty"`
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
		MaxTokens:   int64(getEnvInt("AI_MAX_TOKENS", 2000)),
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

// NewAIClientWithConfig creates a new AI client with custom configuration
func NewAIClientWithConfig(config AIConfig) (*AIClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Set defaults if not provided
	if config.APIEndpoint == "" {
		config.APIEndpoint = "https://api.openai.com/v1/chat/completions"
	}
	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 2000
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	return &AIClient{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// ValidateConfig validates AI configuration
func ValidateConfig(config AIConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if config.MaxTokens < 100 || config.MaxTokens > 8000 {
		return fmt.Errorf("max tokens should be between 100 and 8000")
	}
	if config.Temperature < 0.0 || config.Temperature > 2.0 {
		return fmt.Errorf("temperature should be between 0.0 and 2.0")
	}
	return nil
}

// AnalyzeWithAI performs AI-powered analysis of git statistics
func (c *AIClient) AnalyzeWithAI(stats *analyzer.Statistics, basicReport string) (string, error) {
	prompt := c.buildAnalysisPrompt(stats, basicReport)
	// log.Println("AI analysis request with prompt:", prompt)
	isKnowledgeBase := c.detectKnowledgeBaseProject(stats)
	return c.sendChatRequest(prompt, isKnowledgeBase)
}

// buildAnalysisPrompt creates a prompt for AI analysis
func (c *AIClient) buildAnalysisPrompt(stats *analyzer.Statistics, basicReport string) string {
	msg := i18n.T()
	
	// Detect if this is a knowledge base project
	isKnowledgeBase := c.detectKnowledgeBaseProject(stats)
	
	if isKnowledgeBase {
		prompt := fmt.Sprintf(msg.AIKnowledgeBasePromptTemplate,
			basicReport,
			stats.TotalCommits,
			len(stats.AuthorStats),
			stats.TimeStats.FirstCommit.Format("2006-01-02"),
			stats.TimeStats.LastCommit.Format("2006-01-02"),
			stats.TimeStats.ActiveDays)
		return prompt
	}
	
	prompt := fmt.Sprintf(msg.AIPromptTemplate,
		basicReport,
		stats.TotalCommits,
		len(stats.AuthorStats),
		stats.TimeStats.FirstCommit.Format("2006-01-02"),
		stats.TimeStats.LastCommit.Format("2006-01-02"),
		stats.TimeStats.ActiveDays)

	return prompt
}

// detectKnowledgeBaseProject detects if the repository is a knowledge base project
func (c *AIClient) detectKnowledgeBaseProject(stats *analyzer.Statistics) bool {
	if stats == nil || stats.FileStats == nil {
		return false
	}
	
	// Count file types
	mdCount := 0
	codeFileCount := 0
	totalFiles := len(stats.FileStats)
	
	codeExtensions := map[string]bool{
		".go": true, ".java": true, ".py": true, ".js": true, ".ts": true,
		".cpp": true, ".c": true, ".h": true, ".rs": true, ".rb": true,
		".php": true, ".swift": true, ".kt": true, ".cs": true, ".vb": true,
		".scala": true, ".clj": true, ".r": true, ".lua": true, ".sh": true,
	}
	
	for file := range stats.FileStats {
		if strings.HasSuffix(file, ".md") {
			mdCount++
		}
		for ext := range codeExtensions {
			if strings.HasSuffix(file, ext) {
				codeFileCount++
				break
			}
		}
	}
	
	// Knowledge base indicators:
	// - High ratio of markdown files (>50%)
	// - Few code files (<20% of total)
	// - Total files suggest documentation structure
	if totalFiles > 10 {
		mdRatio := float64(mdCount) / float64(totalFiles)
		codeRatio := float64(codeFileCount) / float64(totalFiles)
		
		// If markdown files dominate and code files are minimal, it's likely a knowledge base
		return mdRatio > 0.5 && codeRatio < 0.2
	}
	
	return false
}

// sendChatRequest sends a request to the AI chat API
func (c *AIClient) sendChatRequest(prompt string, isKnowledgeBase bool) (string, error) {
	msg := i18n.T()
	
	systemMessage := msg.AISystemMessage
	if isKnowledgeBase {
		systemMessage = msg.AIKnowledgeBaseSystemMessage
	}
	
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("AI_API_KEY")),
		option.WithBaseURL(os.Getenv("AI_API_ENDPOINT")),
	)
	chatCompletion, err := client.Chat.Completions.New(
		context.TODO(), 
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemMessage),
				openai.UserMessage(prompt),
			},
			Model: "qwen-plus",
			MaxTokens: param.Opt[int64]{Value: c.config.MaxTokens},
			Temperature: param.Opt[float64]{Value: c.config.Temperature},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %v", err)
	}
	return chatCompletion.Choices[0].Message.Content, nil
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
