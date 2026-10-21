package adapter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

// Capabilities описывает поддерживаемые фичи конкретного провайдера/модели
type Capabilities struct {
	SupportsStreaming       bool `json:"supports_streaming"`
	SupportsTools           bool `json:"supports_tools"`
	SupportsVision          bool `json:"supports_vision"`
	SupportsAudio           bool `json:"supports_audio"`
	SupportsReasoning       bool `json:"supports_reasoning"`
	SupportsStructuredJSON  bool `json:"supports_structured_json"`
	SupportsPromptCaching   bool `json:"supports_prompt_caching"`
	MaxInputTokens          int  `json:"max_input_tokens"`
	MaxOutputTokens         int  `json:"max_output_tokens"`
}

// EndpointConfig содержит параметры подключения к конкретному API эндпоинту
type EndpointConfig struct {
	ProviderID      string        `json:"provider_id"`
	BaseURL         string        `json:"base_url"`
	APIKeySecretRef string        `json:"api_key_secret_ref"`
	DefaultTimeout  time.Duration `json:"default_timeout"`
	MaxRetries      int           `json:"max_retries"`
	Region          string        `json:"region"`
	Headers         map[string]string `json:"headers"`
}

// NormalizedUsage содержит очищенные и проверенные счетчики токенов
type NormalizedUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CachedTokens     int `json:"cached_tokens,omitempty"`
	ReasoningTokens  int `json:"reasoning_tokens,omitempty"`
}

// ProviderAdapter — ключевой контракт межкомандного взаимодействия (Семинар 7)
type ProviderAdapter interface {
	// ID возвращает уникальный идентификатор адаптера (например, "openai", "anthropic", "mock")
	ID() string

	// Capabilities возвращает матрицу возможностей для указанной модели
	Capabilities(model string) Capabilities

	// BuildRequest транслирует канонический запрос в HTTP-запрос к внешнему провайдеру
	BuildRequest(ctx context.Context, cfg *EndpointConfig, req *publicapi.ChatCompletionRequest) (*http.Request, error)

	// ParseResponse парсит сырой HTTP-ответ провайдера в каноническую структуру
	ParseResponse(resp *http.Response) (*publicapi.ChatCompletionResponse, error)

	// ParseStreamEvent парсит единичный SSE-чанк в канонический формат события
	ParseStreamEvent(line []byte) (*publicapi.ChatCompletionChunk, bool, error)

	// ExtractUsage извлекает и валидирует токены из ответа провайдера (R06/R07)
	ExtractUsage(rawBody []byte) (*NormalizedUsage, error)

	// MapError транслирует код/тело ошибки внешнего API в канонический ErrorObject (R05)
	MapError(statusCode int, rawBody []byte) (canonicalCode int, errObject publicapi.Error)
}

// Registry потокобезопасно регистрирует и предоставляет доступ к адаптерам
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]ProviderAdapter
}

var globalRegistry = &Registry{
	adapters: make(map[string]ProviderAdapter),
}

// Register регистрирует адаптер провайдера
func Register(a ProviderAdapter) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.adapters[a.ID()] = a
}

// Get возвращает адаптер по ID
func Get(id string) (ProviderAdapter, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	a, ok := globalRegistry.adapters[id]
	if !ok {
		return nil, fmt.Errorf("provider adapter %q not registered", id)
	}
	return a, nil
}
