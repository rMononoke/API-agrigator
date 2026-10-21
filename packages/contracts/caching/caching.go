package caching

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ilyaero1911/api-aggregator/packages/contracts/adapter"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

// CacheControl holds directive for ephemeral prompt caching (Anthropic, Gemini)
type CacheControl struct {
	Type string `json:"type"` // "ephemeral"
}

// PrefixHash represents the deterministic signature used for sticky routing
type PrefixHash string

// ComputePrefixHash calculates SHA-256 of the prompt prefix (system + initial messages)
// for deterministic routing to the same upstream worker node.
func ComputePrefixHash(messages []publicapi.ChatCompletionMessage) PrefixHash {
	h := sha256.New()
	var prefixParts []string

	for i, m := range messages {
		// Include system prompts and first turn messages
		if m.Role == "system" || m.Role == "developer" || i < 2 {
			prefixParts = append(prefixParts, fmt.Sprintf("%s:%v", m.Role, m.Content))
		}
	}

	h.Write([]byte(strings.Join(prefixParts, "||")))
	return PrefixHash(hex.EncodeToString(h.Sum(nil))[:16])
}

// CachePricing defines token costs in USD per 1M tokens
type CachePricing struct {
	BaseInputPer1M      float64 `json:"base_input_per_1m"`
	CacheReadPer1M      float64 `json:"cache_read_per_1m"`       // typically 50-90% discount
	CacheWritePer1M     float64 `json:"cache_write_per_1m"`      // typically 25% surcharge for TTL
	BaseOutputPer1M     float64 `json:"base_output_per_1m"`
}

// Default pricing presets for major providers (as of 2026)
var (
	AnthropicClaude35Sonnet = CachePricing{
		BaseInputPer1M:  3.00,
		CacheReadPer1M:  0.30,  // 90% discount ($0.30 / 1M)
		CacheWritePer1M: 3.75,  // 25% write surcharge ($3.75 / 1M)
		BaseOutputPer1M: 15.00,
	}
	OpenAIGpt4o = CachePricing{
		BaseInputPer1M:  2.50,
		CacheReadPer1M:  1.25,  // 50% discount ($1.25 / 1M)
		CacheWritePer1M: 2.50,  // automatic write included in base
		BaseOutputPer1M: 10.00,
	}
	DeepSeekV3 = CachePricing{
		BaseInputPer1M:  0.14,
		CacheReadPer1M:  0.014, // 90% discount ($0.014 / 1M)
		CacheWritePer1M: 0.14,
		BaseOutputPer1M: 0.28,
	}
)

// BillingCalculation details cost savings from prompt caching
type BillingCalculation struct {
	BaseCostUSD      float64 `json:"base_cost_usd"`
	ActualCostUSD    float64 `json:"actual_cost_usd"`
	SavedUSD         float64 `json:"saved_usd"`
	DiscountPercent  float64 `json:"discount_percent"`
	CachedTokens     int     `json:"cached_tokens"`
	CacheWriteTokens int     `json:"cache_write_tokens"`
}

// CalculateSavings evaluates financial benefit of cached tokens against baseline
func CalculateSavings(p CachePricing, u *adapter.NormalizedUsage) (*BillingCalculation, error) {
	if u == nil {
		return nil, fmt.Errorf("usage cannot be nil")
	}

	nonCachedPrompt := u.PromptTokens - u.CachedTokens
	if nonCachedPrompt < 0 {
		nonCachedPrompt = 0
	}

	// 1. Base cost if no prompt caching was used
	basePromptCost := (float64(u.PromptTokens) / 1_000_000.0) * p.BaseInputPer1M
	completionCost := (float64(u.CompletionTokens) / 1_000_000.0) * p.BaseOutputPer1M
	baseCost := basePromptCost + completionCost

	// 2. Actual cost applying cache read discounts and cache write surcharges
	actualPromptCost := (float64(nonCachedPrompt) / 1_000_000.0) * p.BaseInputPer1M
	cacheReadCost := (float64(u.CachedTokens) / 1_000_000.0) * p.CacheReadPer1M
	cacheWriteCost := (float64(u.CacheWriteTokens) / 1_000_000.0) * p.CacheWritePer1M

	actualCost := actualPromptCost + cacheReadCost + cacheWriteCost + completionCost
	savedUSD := baseCost - actualCost
	if savedUSD < 0 {
		savedUSD = 0
	}

	var discountPct float64
	if baseCost > 0 {
		discountPct = (savedUSD / baseCost) * 100.0
	}

	return &BillingCalculation{
		BaseCostUSD:      baseCost,
		ActualCostUSD:    actualCost,
		SavedUSD:         savedUSD,
		DiscountPercent:  discountPct,
		CachedTokens:     u.CachedTokens,
		CacheWriteTokens: u.CacheWriteTokens,
	}, nil
}
