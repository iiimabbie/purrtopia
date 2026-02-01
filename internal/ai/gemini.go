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
	client          *genai.Client
	cfg             *config.GeminiConfig
	once            sync.Once
	gameInformation string
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

// SetGameInformation sets the game information for chat responses
func SetGameInformation(info string) {
	gameInformation = info
}

// ChatResponse generates a natural chat response when the bot is mentioned
func ChatResponse(ctx context.Context, message string) (string, error) {
	prompt := buildChatPrompt(message)

	// Try primary API first
	if client != nil {
		result, err := client.Models.GenerateContent(ctx, cfg.Model, []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: prompt}},
			},
		}, &genai.GenerateContentConfig{
			MaxOutputTokens: genai.Ptr[int32](2048),
			Temperature:     genai.Ptr(float32(0.7)),
		})
		if err == nil {
			responseText := strings.TrimSpace(result.Text())
			log.Printf("Chat AI response (primary): %s", responseText)
			return responseText, nil
		}

		// If rate limited, try fallback
		if isRateLimitError(err) {
			log.Printf("Primary API rate limited, trying fallback...")
		} else {
			return "", fmt.Errorf("failed to generate content: %w", err)
		}
	}

	// Try fallback API
	responseText, err := callFallbackAPI(ctx, prompt, 2048, 0.7)
	if err != nil {
		return "", fmt.Errorf("fallback API failed: %w", err)
	}

	log.Printf("Chat AI response (fallback): %s", responseText)
	return responseText, nil
}

func buildChatPrompt(message string) string {
	return fmt.Sprintf(`妳是一個正在玩心動小鎮遊戲的玩家，名叫「Purrtopia」，綽號是「prr」。你熱情友善，喜歡幫助其他玩家，說話輕鬆自然，偶爾會用一些可愛的語氣詞，但是你不是貓娘，不會說喵。雖然你是玩家但是假如真的有人問你你的id或是小鎮id，妳不可以真的就隨便回答了，因為事實上妳確實沒有。

以下是你知道的遊戲資訊：
%s

---

現在有玩家問你：「%s」

請用輕鬆自然的口吻回答，像是一個一起玩遊戲的朋友在聊天。
- 如果問題跟遊戲有關，根據你知道的資訊回答
- 如果你不知道答案，就誠實說不確定，可以建議他問問其他玩家或查看左側頻道列表
- 回答要簡潔，不要太長（最多 2-3 句話）
- 不要使用 markdown 格式
- 可以適當使用表情符號
- 如果是關於粉紅泡泡(粉泡)、金色泡泡(金泡)相關的問題建議他去: 金色泡泡: https://discord.com/channels/1438429535975641120/1438441906001416292，粉紅泡泡: https://discord.com/channels/1438429535975641120/1438442017414975550
- 關於螢石、溜溜木、每日任務、小鎮報、事件等每日活動的問題建議他去: https://discord.com/channels/1438429535975641120/1438441735347769445
- 問「XX在哪裡」「XX要什麼天氣才有」「XX在哪一條河」「XX在哪裡摘」「XX要幾等才有」等位置相關問題建議他去: https://discord.com/channels/1438429535975641120/1452258561156710462
- weather: 詢問隕石,朵朵在哪裡,特殊天氣,彩虹,雨雪,流星等等特殊天氣問題建議他去: https://discord.com/channels/1438429535975641120/1438747864963747890`, gameInformation, message)
}

// ClassifyQuestion uses Gemini to classify if a message is a question and its type
func ClassifyQuestion(ctx context.Context, message string) (*QuestionResult, error) {
	prompt := buildClassificationPrompt(message)

	// Try primary API first
	if client != nil {
		result, err := client.Models.GenerateContent(ctx, cfg.Model, []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: prompt}},
			},
		}, &genai.GenerateContentConfig{
			MaxOutputTokens:  genai.Ptr[int32](4096),
			Temperature:      genai.Ptr(float32(0.1)),
			ResponseMIMEType: "application/json",
		})
		if err == nil {
			responseText := result.Text()
			log.Printf("Raw AI response (primary): %s", responseText)
			return parseQuestionResult(responseText, message)
		}

		// If rate limited, try fallback
		if isRateLimitError(err) {
			log.Printf("Primary API rate limited, trying fallback...")
		} else {
			return nil, fmt.Errorf("failed to generate content: %w", err)
		}
	}

	// Try fallback API
	responseText, err := callFallbackAPI(ctx, prompt, 4096, 0.1)
	if err != nil {
		return nil, fmt.Errorf("fallback API failed: %w", err)
	}

	log.Printf("Raw AI response (fallback): %s", responseText)
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
- bubble: 必須同時滿足以下條件才能回傳 bubble：(1) 訊息中明確提到「粉紅泡泡」「粉泡」「金色泡泡」「金泡」「黃色泡泡」其中之一 (2) 是在問這些泡泡的「位置」。不符合的情況包括：問「是什麼顏色」、問「任務泡泡」、問「怎麼拿」、問其他種類的泡泡、或沒有明確指出顏色，這些都應該回傳 general。
- daily_town: 關於螢石、溜溜木、事件時間等問題
- location: 「只有」詢問魚類、昆蟲、鳥類等「動物」的位置時才回傳 location。必須包含明確的動物名稱如「鱸魚」「蝴蝶」「白鷺鷥」等。注意：「阿嚕」是遊戲角色不是動物、「寶藏」「泡泡」「NPC」「任務」「植物」都不是動物，這些都應回傳 general。
- weather: 詢問隕石日,特殊天氣,彩虹日,雨雪日,流星日等等特殊天氣的"時間"問題, 以及朵朵在"哪裡"的問題
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
