package reasoning

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ilyaero1911/api-aggregator/packages/contracts/adapter"
	"github.com/ilyaero1911/api-aggregator/packages/contracts/publicapi"
)

// Effort level for canonical reasoning parameter
type Effort string

const (
	EffortNone    Effort = "none"
	EffortMinimal Effort = "minimal"
	EffortLow     Effort = "low"
	EffortMedium  Effort = "medium"
	EffortHigh    Effort = "high"
)

// Config represents unified reasoning configuration (ID24_API-60 / E04-14)
type Config struct {
	Effort    Effort `json:"effort,omitempty"`
	MaxTokens int    `json:"max_tokens,omitempty"`
	Exclude   bool   `json:"exclude,omitempty"`
}

// ProviderReasoningRequest holds upstream-specific parameters
type ProviderReasoningRequest struct {
	OpenAIReasoningEffort string                 `json:"reasoning_effort,omitempty"`
	AnthropicThinking     *AnthropicThinking     `json:"thinking,omitempty"`
	GeminiThinkingConfig  *GeminiThinkingConfig  `json:"thinking_config,omitempty"`
	DeepSeekReasoner      bool                   `json:"deepseek_reasoner,omitempty"`
}

type AnthropicThinking struct {
	Type         string `json:"type"` // "enabled"
	BudgetTokens int    `json:"budget_tokens"`
}

type GeminiThinkingConfig struct {
	ThinkingBudget int `json:"thinking_budget"`
}

// MapReasoning converts canonical Config into provider-specific payload
func MapReasoning(provider string, model string, cfg *Config) (*ProviderReasoningRequest, error) {
	if cfg == nil || cfg.Effort == EffortNone {
		return &ProviderReasoningRequest{}, nil
	}

	res := &ProviderReasoningRequest{}

	switch strings.ToLower(provider) {
	case "openai", "azure-openai":
		// OpenAI supports o1/o3 reasoning_effort: low, medium, high
		effortStr := string(cfg.Effort)
		if cfg.Effort == EffortMinimal {
			effortStr = "low" // mapping minimal to lowest supported OpenAI effort
		}
		if effortStr != "low" && effortStr != "medium" && effortStr != "high" {
			effortStr = "medium"
		}
		res.OpenAIReasoningEffort = effortStr

	case "anthropic":
		// Anthropic Claude 3.7+ Messages API: thinking { type: "enabled", budget_tokens: N }
		budget := cfg.MaxTokens
		if budget <= 0 {
			switch cfg.Effort {
			case EffortMinimal:
				budget = 1024
			case EffortLow:
				budget = 2048
			case EffortHigh:
				budget = 16384
			default:
				budget = 4096 // medium
			}
		}
		res.AnthropicThinking = &AnthropicThinking{
			Type:         "enabled",
			BudgetTokens: budget,
		}

	case "gemini", "google":
		// Google Gemini thinkingConfig
		budget := cfg.MaxTokens
		if budget <= 0 {
			switch cfg.Effort {
			case EffortMinimal, EffortLow:
				budget = 2048
			case EffortHigh:
				budget = 16384
			default:
				budget = 8192
			}
		}
		res.GeminiThinkingConfig = &GeminiThinkingConfig{
			ThinkingBudget: budget,
		}

	case "deepseek":
		// DeepSeek R1 reasoning model
		if strings.Contains(strings.ToLower(model), "reasoner") || strings.Contains(strings.ToLower(model), "r1") {
			res.DeepSeekReasoner = true
		}

	default:
		// Generic / local compatible
		res.OpenAIReasoningEffort = string(cfg.Effort)
	}

	return res, nil
}

// ExtractedReasoning contains thinking text and token accounting
type ExtractedReasoning struct {
	ReasoningText   string
	ReasoningTokens int
}

// ExtractReasoningFromResponse parses reasoning content and token metrics from raw provider response
func ExtractReasoningFromResponse(provider string, rawBody []byte, exclude bool) (*ExtractedReasoning, error) {
	result := &ExtractedReasoning{}

	switch strings.ToLower(provider) {
	case "deepseek":
		var dsResp struct {
			Choices []struct {
				Message struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content"`
				} `json:"message"`
			} `json:"choices"`
			Usage struct {
				CompletionTokensDetails struct {
					ReasoningTokens int `json:"reasoning_tokens"`
				} `json:"completion_tokens_details"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(rawBody, &dsResp); err != nil {
			return nil, fmt.Errorf("unmarshal deepseek response: %w", err)
		}
		if len(dsResp.Choices) > 0 && !exclude {
			result.ReasoningText = dsResp.Choices[0].Message.ReasoningContent
		}
		result.ReasoningTokens = dsResp.Usage.CompletionTokensDetails.ReasoningTokens

	case "anthropic":
		var antResp struct {
			Content []struct {
				Type     string `json:"type"`
				Thinking string `json:"thinking"`
				Text     string `json:"text"`
			} `json:"content"`
			Usage struct {
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(rawBody, &antResp); err != nil {
			return nil, fmt.Errorf("unmarshal anthropic response: %w", err)
		}
		var thinkingParts []string
		for _, part := range antResp.Content {
			if part.Type == "thinking" && part.Thinking != "" {
				thinkingParts = append(thinkingParts, part.Thinking)
			}
		}
		if !exclude {
			result.ReasoningText = strings.Join(thinkingParts, "\n\n")
		}

	case "openai", "azure-openai":
		var oaiResp struct {
			Usage struct {
				CompletionTokensDetails struct {
					ReasoningTokens int `json:"reasoning_tokens"`
				} `json:"completion_tokens_details"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(rawBody, &oaiResp); err != nil {
			return nil, fmt.Errorf("unmarshal openai response: %w", err)
		}
		result.ReasoningTokens = oaiResp.Usage.CompletionTokensDetails.ReasoningTokens

	default:
		// Generic OpenAI compatible
		var genResp struct {
			Choices []struct {
				Message struct {
					ReasoningContent string `json:"reasoning_content"`
					Reasoning        string `json:"reasoning"`
				} `json:"message"`
			} `json:"choices"`
			Usage struct {
				CompletionTokensDetails struct {
					ReasoningTokens int `json:"reasoning_tokens"`
				} `json:"completion_tokens_details"`
				ReasoningTokens int `json:"reasoning_tokens"`
			} `json:"usage"`
		}
		_ = json.Unmarshal(rawBody, &genResp)
		if len(genResp.Choices) > 0 && !exclude {
			if genResp.Choices[0].Message.ReasoningContent != "" {
				result.ReasoningText = genResp.Choices[0].Message.ReasoningContent
			} else {
				result.ReasoningText = genResp.Choices[0].Message.Reasoning
			}
		}
		if genResp.Usage.CompletionTokensDetails.ReasoningTokens > 0 {
			result.ReasoningTokens = genResp.Usage.CompletionTokensDetails.ReasoningTokens
		} else {
			result.ReasoningTokens = genResp.Usage.ReasoningTokens
		}
	}

	return result, nil
}

// NormalizeReasoningUsage validates and updates canonical usage struct
func NormalizeReasoningUsage(usage *adapter.NormalizedUsage, reasoningTokens int) error {
	if reasoningTokens < 0 {
		return fmt.Errorf("reasoning tokens cannot be negative: %d", reasoningTokens)
	}
	if usage != nil {
		usage.ReasoningTokens = reasoningTokens
	}
	return nil
}
