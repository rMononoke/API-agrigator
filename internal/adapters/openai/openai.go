package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ilyaero1911/api-aggregator/packages/contracts/adapter"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ID() string {
	return "openai"
}

func (a *Adapter) Capabilities(model string) adapter.Capabilities {
	caps := adapter.Capabilities{
		SupportsStreaming:      true,
		SupportsTools:          true,
		SupportsVision:         true,
		SupportsAudio:          false,
		SupportsReasoning:      false,
		SupportsStructuredJSON: true,
		SupportsPromptCaching:  true,
		MaxInputTokens:         128000,
		MaxOutputTokens:        16384,
	}

	if strings.HasPrefix(model, "o1") || strings.HasPrefix(model, "o3") {
		caps.SupportsReasoning = true
	}
	if strings.Contains(model, "audio") {
		caps.SupportsAudio = true
	}
	return caps
}

type OpenAIUsageDetails struct {
	CachedTokens    int `json:"cached_tokens"`
	ReasoningTokens int `json:"reasoning_tokens"`
}

type OpenAIUsage struct {
	PromptTokens            int                 `json:"prompt_tokens"`
	CompletionTokens        int                 `json:"completion_tokens"`
	TotalTokens             int                 `json:"total_tokens"`
	PromptTokensDetails     *OpenAIUsageDetails `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails *OpenAIUsageDetails `json:"completion_tokens_details,omitempty"`
}

type OpenAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (a *Adapter) BuildRequest(ctx context.Context, cfg *adapter.EndpointConfig, req *publicapi.ChatCompletionRequest) (*http.Request, error) {
	url := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Поддержка стандартного OpenAI и Azure OpenAI
	if strings.Contains(cfg.BaseURL, "openai.azure.com") {
		httpReq.Header.Set("api-key", cfg.APIKeySecretRef)
	} else {
		httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKeySecretRef)
	}

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
		return nil, fmt.Errorf("openai error (%d): %s", resp.StatusCode, errObj.Message)
	}

	var canonical publicapi.ChatCompletionResponse
	if err := json.Unmarshal(body, &canonical); err != nil {
		return nil, fmt.Errorf("unmarshal canonical response: %w", err)
	}

	return &canonical, nil
}

func (a *Adapter) ParseStreamEvent(line []byte) (*publicapi.ChatCompletionChunk, bool, error) {
	str := strings.TrimSpace(string(line))
	if str == "" || strings.HasPrefix(str, ":") {
		return nil, false, nil
	}
	if !strings.HasPrefix(str, "data: ") {
		return nil, false, nil
	}

	data := strings.TrimPrefix(str, "data: ")
	if data == "[DONE]" {
		return nil, true, nil
	}

	var chunk publicapi.ChatCompletionChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return nil, false, fmt.Errorf("unmarshal sse chunk: %w", err)
	}

	return &chunk, false, nil
}

func (a *Adapter) ExtractUsage(rawBody []byte) (*adapter.NormalizedUsage, error) {
	var resp struct {
		Usage *OpenAIUsage `json:"usage"`
	}
	if err := json.Unmarshal(rawBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal usage: %w", err)
	}
	if resp.Usage == nil {
		return nil, fmt.Errorf("usage block is missing")
	}

	if resp.Usage.PromptTokens < 0 || resp.Usage.CompletionTokens < 0 {
		return nil, fmt.Errorf("negative token count rejected (R06)")
	}

	u := &adapter.NormalizedUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}

	if resp.Usage.PromptTokensDetails != nil {
		u.CachedTokens = resp.Usage.PromptTokensDetails.CachedTokens
	}
	if resp.Usage.CompletionTokensDetails != nil {
		u.ReasoningTokens = resp.Usage.CompletionTokensDetails.ReasoningTokens
	}

	return u, nil
}

func (a *Adapter) MapError(statusCode int, rawBody []byte) (int, publicapi.Error) {
	var errResp OpenAIErrorResponse
	_ = json.Unmarshal(rawBody, &errResp)

	msg := errResp.Error.Message
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d from OpenAI upstream", statusCode)
	}

	return statusCode, publicapi.Error{
		Code:    errResp.Error.Code,
		Message: msg,
		Type:    errResp.Error.Type,
	}
}

// ParseRateLimits извлекает лимиты из HTTP-заголовков OpenAI
func ParseRateLimits(h http.Header) (remReq, remTokens int, resetReq time.Duration) {
	remReq, _ = strconv.Atoi(h.Get("x-ratelimit-remaining-requests"))
	remTokens, _ = strconv.Atoi(h.Get("x-ratelimit-remaining-tokens"))
	if s := h.Get("x-ratelimit-reset-requests"); s != "" {
		resetReq, _ = time.ParseDuration(s)
	}
	return
}
