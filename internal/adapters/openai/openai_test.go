package openai

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ilyaero1911/api-aggregator/packages/contracts/adapter"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

func TestOpenAIAdapter_BuildRequest(t *testing.T) {
	adp := New()
	cfg := &adapter.EndpointConfig{
		BaseURL:         "https://api.openai.com/v1",
		APIKeySecretRef: "sk-test-12345",
		Headers:         map[string]string{"OpenAI-Organization": "org-id"},
	}

	temp := float32(0.7)
	req := &publicapi.ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []publicapi.ChatCompletionMessage{
			{Role: "user", Content: "hello"},
		},
		Temperature: &temp,
	}

	httpReq, err := adp.BuildRequest(context.Background(), cfg, req)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if httpReq.URL.String() != "https://api.openai.com/v1/chat/completions" {
		t.Errorf("unexpected URL: %s", httpReq.URL.String())
	}
	if httpReq.Header.Get("Authorization") != "Bearer sk-test-12345" {
		t.Errorf("missing Authorization header")
	}
	if httpReq.Header.Get("OpenAI-Organization") != "org-id" {
		t.Errorf("missing custom header")
	}
}

func TestOpenAIAdapter_AzureOpenAI(t *testing.T) {
	adp := New()
	cfg := &adapter.EndpointConfig{
		BaseURL:         "https://my-resource.openai.azure.com/openai/deployments/gpt-4",
		APIKeySecretRef: "azure-key-xyz",
	}

	httpReq, err := adp.BuildRequest(context.Background(), cfg, &publicapi.ChatCompletionRequest{Model: "gpt-4"})
	if err != nil {
		t.Fatalf("Azure request failed: %v", err)
	}
	if httpReq.Header.Get("api-key") != "azure-key-xyz" {
		t.Errorf("missing Azure api-key header")
	}
	if httpReq.Header.Get("Authorization") != "" {
		t.Errorf("should not set Authorization header for Azure")
	}
}

func TestOpenAIAdapter_ExtractUsage(t *testing.T) {
	adp := New()
	validJSON := []byte(`{
		"usage": {
			"prompt_tokens": 100,
			"completion_tokens": 50,
			"total_tokens": 150,
			"prompt_tokens_details": {"cached_tokens": 30},
			"completion_tokens_details": {"reasoning_tokens": 15}
		}
	}`)

	u, err := adp.ExtractUsage(validJSON)
	if err != nil {
		t.Fatalf("ExtractUsage failed: %v", err)
	}
	if u.PromptTokens != 100 || u.CompletionTokens != 50 || u.CachedTokens != 30 || u.ReasoningTokens != 15 {
		t.Errorf("unexpected usage counts: %+v", u)
	}

	badJSON := []byte(`{"usage":{"prompt_tokens": -10, "completion_tokens": 20}}`)
	if _, err := adp.ExtractUsage(badJSON); err == nil {
		t.Errorf("expected error on negative tokens (R06)")
	}
}

func TestOpenAIAdapter_ParseRateLimits(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-remaining-requests", "4999")
	h.Set("x-ratelimit-remaining-tokens", "999800")
	h.Set("x-ratelimit-reset-requests", "1s")

	remReq, remTokens, reset := ParseRateLimits(h)
	if remReq != 4999 || remTokens != 999800 || reset != time.Second {
		t.Errorf("failed parsing rate limits: %d, %d, %v", remReq, remTokens, reset)
	}
}
