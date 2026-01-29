package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

// QuestionResult represents the AI classification result
type QuestionResult struct {
	IsQuestion   bool    `json:"is_question"`
	QuestionType string  `json:"question_type"`
	Confidence   float64 `json:"confidence"`
	OriginalText string  `json:"original_text"`
}

// Init initializes the Gemini client
func Init(geminiCfg *config.GeminiConfig) error {
	if geminiCfg.APIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is not set")
	}

	var initErr error
	once.Do(func() {
		cfg = geminiCfg
		ctx := context.Background()
		client, initErr = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  cfg.APIKey,
			Backend: genai.BackendGeminiAPI,
		})
	})
	return initErr
}

// IsInitialized returns true if the Gemini client is initialized
func IsInitialized() bool {
	return client != nil
}

// ClassifyQuestion uses Gemini to classify if a message is a question and its type
func ClassifyQuestion(ctx context.Context, message string) (*QuestionResult, error) {
	if client == nil {
		return nil, fmt.Errorf("Gemini client not initialized")
	}

	prompt := buildClassificationPrompt(message)

	result, err := client.Models.GenerateContent(ctx, cfg.Model, []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{{Text: prompt}},
		},
	}, &genai.GenerateContentConfig{
		MaxOutputTokens:  genai.Ptr[int32](2048),
		Temperature:      genai.Ptr(float32(0.1)),
		ResponseMIMEType: "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	responseText := result.Text()
	log.Printf("Raw AI response: %s", responseText)
	return parseQuestionResult(responseText, message)
}

func buildClassificationPrompt(message string) string {
	return fmt.Sprintf(`你是一個問題分類助手。請分析以下訊息，判斷是否在詢問問題。

訊息：「%s」

請以 JSON 格式回覆，格式如下：
{
  "is_question": true或false,
  "question_type": "bubble" 或 "daily_town" 或 "location" 或 "general" 或 "none",
  "confidence": 0.0到1.0之間的數字
}

問題類型說明：
- bubble: 關於粉紅泡泡(粉泡)、金色泡泡(金泡)相關的問題
- daily_town: 關於螢石、溜溜木、每日任務、小鎮報、事件等每日活動的問題
- location: 問「XX在哪裡」「XX要什麼天氣才有」「XX在哪一條河」「XX在哪裡摘」「XX要幾等才有」等位置相關問題
- weather: 詢問隕石,朵朵在哪裡,特殊天氣,彩虹,雨雪,流星等等特殊天氣問題
- general: 其他一般性問題
- none: 不是問題，只是聊天

注意：如果只是在聊天討論而非提問，is_question 應為 false，question_type 為 "none"。
只回覆 JSON，不要有其他文字。`, message)
}

func parseQuestionResult(responseText, originalMessage string) (*QuestionResult, error) {
	// Clean up the response - remove markdown code blocks if present
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var result QuestionResult
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w, response: %s", err, responseText)
	}

	result.OriginalText = originalMessage
	return &result, nil
}
