package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"google.golang.org/genai"
	"purrtopia/internal/config"
)

var (
	client *genai.Client
	cfg    *config.GeminiConfig
	once   sync.Once
)

// Init initializes the Gemini client
func Init(geminiCfg *config.GeminiConfig) error {
	cfg = geminiCfg

	// At least one API must be configured
	if geminiCfg.APIKey == "" && geminiCfg.FallbackEndpoint == "" {
		return fmt.Errorf("either GEMINI_API_KEY or GEMINI_FALLBACK_ENDPOINT must be set")
	}

	// Initialize primary client if API key is provided
	if geminiCfg.APIKey != "" {
		var initErr error
		once.Do(func() {
			ctx := context.Background()
			client, initErr = genai.NewClient(ctx, &genai.ClientConfig{
				APIKey:  cfg.APIKey,
				Backend: genai.BackendGeminiAPI,
			})
		})
		if initErr != nil {
			log.Printf("Primary Gemini client init failed: %v, will use fallback only", initErr)
		}
	}

	if geminiCfg.FallbackEndpoint != "" {
		log.Printf("Fallback API configured: %s", geminiCfg.FallbackEndpoint)
	}

	return nil
}

// IsInitialized returns true if the Gemini client is initialized
func IsInitialized() bool {
	return client != nil
}

// GenerateContent calls the Gemini API (primary + fallback) with the given prompt
func GenerateContent(ctx context.Context, prompt string, maxTokens int32, temperature float32) (string, error) {
	// Try primary API first
	if client != nil {
		result, err := client.Models.GenerateContent(ctx, cfg.Model, []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: prompt}},
			},
		}, &genai.GenerateContentConfig{
			MaxOutputTokens: genai.Ptr(maxTokens),
			Temperature:     genai.Ptr(temperature),
		})
		if err == nil {
			return strings.TrimSpace(result.Text()), nil
		}

		if isRateLimitError(err) {
			log.Printf("Primary API rate limited, trying fallback...")
		} else {
			return "", fmt.Errorf("failed to generate content: %w", err)
		}
	}

	// Try fallback API
	responseText, err := callFallbackAPI(ctx, prompt, int(maxTokens), temperature)
	if err != nil {
		return "", fmt.Errorf("fallback API failed: %w", err)
	}

	return responseText, nil
}

// ============================================
// Fallback API (OpenAI-compatible endpoint)
// ============================================

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float32         `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func callFallbackAPI(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	if cfg.FallbackEndpoint == "" {
		return "", fmt.Errorf("fallback endpoint not configured")
	}

	reqBody := openAIRequest{
		Model: cfg.FallbackModel,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimSuffix(cfg.FallbackEndpoint, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if cfg.FallbackAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.FallbackAPIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "429") || strings.Contains(errStr, "RESOURCE_EXHAUSTED") || strings.Contains(errStr, "quota")
}
