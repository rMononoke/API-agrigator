package generic

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ilyaero1911/api-aggregator/internal/adapters/openai"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/adapter"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

// ProviderDefinition задает декларативную конфигурацию OpenAI-совместимого провайдера
type ProviderDefinition struct {
	ID           string               `json:"id" yaml:"id"`
	Name         string               `json:"name" yaml:"name"`
	BaseURL      string               `json:"base_url" yaml:"base_url"`
	AuthHeader   string               `json:"auth_header" yaml:"auth_header"`
	AuthPrefix   string               `json:"auth_prefix" yaml:"auth_prefix"`
	SecretRef    string               `json:"secret_ref" yaml:"secret_ref"`
	Capabilities adapter.Capabilities `json:"capabilities" yaml:"capabilities"`
}

// ConfigurableAdapter реализует ProviderAdapter для любого совместимого провайдера без написания Go-кода
type ConfigurableAdapter struct {
	def         ProviderDefinition
	baseAdapter *openai.Adapter
}

func NewConfigurableAdapter(def ProviderDefinition) (*ConfigurableAdapter, error) {
	if def.ID == "" || def.BaseURL == "" {
		return nil, fmt.Errorf("provider definition must include id and base_url")
	}
	if def.AuthHeader == "" {
		def.AuthHeader = "Authorization"
	}
	if def.AuthPrefix == "" {
		def.AuthPrefix = "Bearer "
	}

	return &ConfigurableAdapter{
		def:         def,
		baseAdapter: openai.New(),
	}, nil
}

func (c *ConfigurableAdapter) ID() string {
	return c.def.ID
}

func (c *ConfigurableAdapter) Capabilities(model string) adapter.Capabilities {
	return c.def.Capabilities
}

func (c *ConfigurableAdapter) BuildRequest(ctx context.Context, cfg *adapter.EndpointConfig, req *publicapi.ChatCompletionRequest) (*http.Request, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = c.def.BaseURL
	}
	httpReq, err := c.baseAdapter.BuildRequest(ctx, cfg, req)
	if err != nil {
		return nil, err
	}

	// Кастомный заголовок авторизации если требуется
	if c.def.AuthHeader != "Authorization" || c.def.AuthPrefix != "Bearer " {
		httpReq.Header.Del("Authorization")
		httpReq.Header.Set(c.def.AuthHeader, c.def.AuthPrefix+cfg.APIKeySecretRef)
	}

	return httpReq, nil
}

func (c *ConfigurableAdapter) ParseResponse(resp *http.Response) (*publicapi.ChatCompletionResponse, error) {
	return c.baseAdapter.ParseResponse(resp)
}

func (c *ConfigurableAdapter) ParseStreamEvent(line []byte) (*publicapi.ChatCompletionChunk, bool, error) {
	return c.baseAdapter.ParseStreamEvent(line)
}

func (c *ConfigurableAdapter) ExtractUsage(rawBody []byte) (*adapter.NormalizedUsage, error) {
	return c.baseAdapter.ExtractUsage(rawBody)
}

func (c *ConfigurableAdapter) MapError(statusCode int, rawBody []byte) (int, publicapi.Error) {
	return c.baseAdapter.MapError(statusCode, rawBody)
}
