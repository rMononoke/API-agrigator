package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ilyaero1911/api-aggregator/packages/contracts/adapter"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ID() string {
	return "anthropic"
}

func (a *Adapter) Capabilities(model string) adapter.Capabilities {
	return adapter.Capabilities{
		SupportsStreaming:      true,
		SupportsTools:          true,
		SupportsVision:         true,
		SupportsAudio:          false,
		SupportsReasoning:      true, // Extended thinking
		SupportsStructuredJSON: true,
		SupportsPromptCaching:  true,
		MaxInputTokens:         200000,
		MaxOutputTokens:        8192,
	}
}

// AnthropicMessageRequest — структура запроса к Anthropic Messages API
type AnthropicMessageRequest struct {
	Model       string                   `json:"model"`
	Messages    []AnthropicMessage       `json:"messages"`
	System      string                   `json:"system,omitempty"`
	MaxTokens   int                      `json:"max_tokens"`
	Temperature *float32                 `json:"temperature,omitempty"`
	TopP        *float32                 `json:"top_p,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
	Thinking    *AnthropicThinkingConfig `json:"thinking,omitempty"`
	Tools       []AnthropicTool          `json:"tools,omitempty"`
}

type AnthropicThinkingConfig struct {
	Type         string `json:"type"` // "enabled"
	BudgetTokens int    `json:"budget_tokens"`
}

type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type AnthropicMessage struct {
	Role    string      `json:"role"` // "user" или "assistant"
	Content interface{} `json:"content"`
}

type AnthropicUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

type AnthropicResponse struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Role       string                 `json:"role"`
	Content    []AnthropicContentPart `json:"content"`
	StopReason string                 `json:"stop_reason"`
	Usage      AnthropicUsage         `json:"usage"`
}

type AnthropicContentPart struct {
	Type     string                 `json:"type"` // "text", "tool_use", "thinking"
	Text     string                 `json:"text,omitempty"`
	Thinking string                 `json:"thinking,omitempty"`
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Input    map[string]interface{} `json:"input,omitempty"`
}

func (a *Adapter) BuildRequest(ctx context.Context, cfg *adapter.EndpointConfig, req *publicapi.ChatCompletionRequest) (*http.Request, error) {
	url := strings.TrimRight(cfg.BaseURL, "/") + "/v1/messages"

	anthropicReq := AnthropicMessageRequest{
		Model:       req.Model,
		MaxTokens:   4096,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream != nil && *req.Stream,
	}

	if req.MaxTokens != nil {
		anthropicReq.MaxTokens = *req.MaxTokens
	}

	// 1. Вынос system / developer промптов в top-level поле system
	var systemParts []string
	var convMessages []AnthropicMessage

	for _, m := range req.Messages {
		if m.Role == "system" || m.Role == "developer" {
			systemParts = append(systemParts, fmt.Sprintf("%v", m.Content))
		} else {
			role := m.Role
			if role == "tool" {
				role = "user" // Anthropic принимает tool_result в контексте user
			}
			convMessages = append(convMessages, AnthropicMessage{
				Role:    role,
				Content: m.Content,
			})
		}
	}

	if len(systemParts) > 0 {
		anthropicReq.System = strings.Join(systemParts, "\n\n")
	}
	anthropicReq.Messages = convMessages

	bodyBytes, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", cfg.APIKeySecretRef)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	for k, v := range cfg.Headers {
		httpReq.Header.Set(k, v)
	}

	return httpReq, nil
}

func (a *Adapter) ParseResponse(resp *http.Response) (*publicapi.ChatCompletionResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, errObj := a.MapError(resp.StatusCode, body)
		return nil, fmt.Errorf("anthropic error (%d): %s", resp.StatusCode, errObj.Message)
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("unmarshal anthropic response: %w", err)
	}

	// Преобразование контента Anthropic в канонический формат
	var contentText strings.Builder
	for _, part := range anthropicResp.Content {
		if part.Type == "text" {
			contentText.WriteString(part.Text)
		}
	}

	// Маппинг stop_reason
	finishReason := "stop"
	switch anthropicResp.StopReason {
	case "max_tokens":
		finishReason = "length"
	case "tool_use":
		finishReason = "tool_calls"
	case "stop_sequence", "end_turn":
		finishReason = "stop"
	}

	cached := anthropicResp.Usage.CacheReadInputTokens
	canonicalResp := &publicapi.ChatCompletionResponse{
		Id:     anthropicResp.ID,
		Object: "chat.completion",
		Model:  "claude-3-5-sonnet",
		Choices: []publicapi.ChatCompletionChoice{
			{
				Index: 0,
				Message: publicapi.ChatCompletionResponseMessage{
					Role:    "assistant",
					Content: contentText.String(),
				},
				FinishReason: finishReason,
			},
		},
		Usage: publicapi.ChatCompletionUsage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
			PromptTokensDetails: &publicapi.PromptTokensDetails{
				CachedTokens: &cached,
			},
		},
	}

	return canonicalResp, nil
}

func (a *Adapter) ParseStreamEvent(line []byte) (*publicapi.ChatCompletionChunk, bool, error) {
	str := strings.TrimSpace(string(line))
	if str == "" || strings.HasPrefix(str, "event: ") {
		return nil, false, nil
	}
	if !strings.HasPrefix(str, "data: ") {
		return nil, false, nil
	}

	data := strings.TrimPrefix(str, "data: ")
	var event struct {
		Type  string `json:"type"`
		Delta struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"delta"`
	}

	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return nil, false, nil
	}

	if event.Type == "message_stop" {
		return nil, true, nil
	}

	if event.Delta.Text != "" {
		chunk := &publicapi.ChatCompletionChunk{
			Object: "chat.completion.chunk",
			Choices: []publicapi.ChatCompletionChunkChoice{
				{
					Index: 0,
					Delta: publicapi.ChatCompletionChunkDelta{
						Content: &event.Delta.Text,
					},
				},
			},
		}
		return chunk, false, nil
	}

	return nil, false, nil
}

func (a *Adapter) ExtractUsage(rawBody []byte) (*adapter.NormalizedUsage, error) {
	var resp struct {
		Usage *AnthropicUsage `json:"usage"`
	}
	if err := json.Unmarshal(rawBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal usage: %w", err)
	}
	if resp.Usage == nil {
		return nil, fmt.Errorf("usage block is missing")
	}

	if resp.Usage.InputTokens < 0 || resp.Usage.OutputTokens < 0 {
		return nil, fmt.Errorf("negative tokens rejected (R06)")
	}

	return &adapter.NormalizedUsage{
		PromptTokens:     resp.Usage.InputTokens,
		CompletionTokens: resp.Usage.OutputTokens,
		TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		CachedTokens:     resp.Usage.CacheReadInputTokens,
	}, nil
}

func (a *Adapter) MapError(statusCode int, rawBody []byte) (int, publicapi.Error) {
	var errResp struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rawBody, &errResp)

	msg := errResp.Error.Message
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d from Anthropic upstream", statusCode)
	}

	return statusCode, publicapi.Error{
		Code:    errResp.Error.Type,
		Message: msg,
		Type:    errResp.Error.Type,
	}
}
