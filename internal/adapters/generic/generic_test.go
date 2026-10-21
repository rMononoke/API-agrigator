package generic

import (
	"context"
	"testing"

	"github.com/ilyaero1911/api-aggregator/packages/contracts/adapter"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

func TestConfigurableAdapter_EightProviders(t *testing.T) {
	definitions := []ProviderDefinition{
		{ID: "mistral", Name: "Mistral AI", BaseURL: "https://api.mistral.ai/v1"},
		{ID: "groq", Name: "Groq Cloud", BaseURL: "https://api.groq.com/openai/v1"},
		{ID: "xai", Name: "xAI (Grok)", BaseURL: "https://api.x.ai/v1"},
		{ID: "together", Name: "Together AI", BaseURL: "https://api.together.xyz/v1"},
		{ID: "fireworks", Name: "Fireworks AI", BaseURL: "https://api.fireworks.ai/inference/v1"},
		{ID: "deepseek", Name: "DeepSeek Platform", BaseURL: "https://api.deepseek.com/v1"},
		{ID: "qwen", Name: "Alibaba DashScope", BaseURL: "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"},
		{ID: "vllm-local", Name: "vLLM", BaseURL: "http://localhost:8000/v1"},
	}

	if len(definitions) < 6 {
		t.Fatalf("expected at least 6 providers for DoD, got %d", len(definitions))
	}

	for _, def := range definitions {
		adp, err := NewConfigurableAdapter(def)
		if err != nil {
			t.Fatalf("failed initializing adapter %s: %v", def.ID, err)
		}
		if adp.ID() != def.ID {
			t.Errorf("ID mismatch: %s != %s", adp.ID(), def.ID)
		}

		cfg := &adapter.EndpointConfig{
			BaseURL:         def.BaseURL,
			APIKeySecretRef: "secret-token",
		}

		httpReq, err := adp.BuildRequest(context.Background(), cfg, &publicapi.ChatCompletionRequest{
			Model: "default-model",
			Messages: []publicapi.ChatCompletionMessage{
				{Role: "user", Content: "ping"},
			},
		})
		if err != nil {
			t.Fatalf("BuildRequest failed for %s: %v", def.ID, err)
		}
		if httpReq.URL.String() != def.BaseURL+"/chat/completions" {
			t.Errorf("unexpected URL for %s: %s", def.ID, httpReq.URL.String())
		}
	}
}
