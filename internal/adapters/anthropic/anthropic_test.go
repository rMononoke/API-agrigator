package anthropic

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ilyaero1911/api-aggregator/packages/contracts/adapter"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

func TestAnthropicAdapter_BuildRequest(t *testing.T) {
	adp := New()
	cfg := &adapter.EndpointConfig{
		BaseURL:         "https://api.anthropic.com",
		APIKeySecretRef: "ant-key-12345",
	}

	req := &publicapi.ChatCompletionRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []publicapi.ChatCompletionMessage{
			{Role: "system", Content: "You are a helpful assistant"},
			{Role: "user", Content: "Hello world"},
		},
	}

	httpReq, err := adp.BuildRequest(context.Background(), cfg, req)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if httpReq.URL.String() != "https://api.anthropic.com/v1/messages" {
		t.Errorf("unexpected URL: %s", httpReq.URL.String())
	}
	if httpReq.Header.Get("x-api-key") != "ant-key-12345" {
		t.Errorf("missing x-api-key header")
	}
	if httpReq.Header.Get("anthropic-version") != "2023-06-01" {
		t.Errorf("missing anthropic-version header")
	}
}

func TestAnthropicAdapter_ExtractUsage(t *testing.T) {
	adp := New()
	body := []byte(`{
		"usage": {
			"input_tokens": 120,
			"output_tokens": 80,
			"cache_read_input_tokens": 40
		}
	}`)

	u, err := adp.ExtractUsage(body)
	if err != nil {
		t.Fatalf("ExtractUsage failed: %v", err)
	}
	if u.PromptTokens != 120 || u.CompletionTokens != 80 || u.TotalTokens != 200 || u.CachedTokens != 40 {
		t.Errorf("unexpected usage counts: %+v", u)
	}
}
