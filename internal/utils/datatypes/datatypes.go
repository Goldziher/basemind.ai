package datatypes

import (
	"encoding/json"
	"github.com/basemind-ai/monorepo/shared/go/db/models"
	"github.com/shopspring/decimal"
	"time"
)

// OpenAIPromptMessageDTO - DTO for serializing and storing an OpenAI prompt message.
// Note- this struct represents what we store in the DB as part of a JSON array.
type OpenAIPromptMessageDTO struct { // skipcq: TCV-001
	Role              string    `json:"role"                        validate:"oneof=system user function assistant"`
	Content           *string   `json:"content,omitempty"           validate:"omitempty,required"`
	Name              *string   `json:"name,omitempty"`
	FunctionArguments *[]string `json:"functionArguments,omitempty"`
	TemplateVariables *[]string `json:"templateVariables,omitempty"`
}

// OpenAIModelParametersDTO - DTO for serializing and storing OpenAI model parameters.
// Note- this struct represents what we store in the DB as JSON.
type OpenAIModelParametersDTO struct { // skipcq: TCV-001
	MaxTokens        *int32   `json:"maxTokens,omitempty"`
	Temperature      *float32 `json:"temperature,omitempty"`
	TopP             *float32 `json:"topP,omitempty"`
	FrequencyPenalty *float32 `json:"frequencyPenalty,omitempty"`
	PresencePenalty  *float32 `json:"presencePenalty,omitempty"`
}

// CoherePromptMessageDTO - DTO for serializing and storing a Cohere prompt message.
type CoherePromptMessageDTO struct { // skipcq: TCV-001
	Message           string    `json:"message"                     validate:"required"`
	TemplateVariables *[]string `json:"templateVariables,omitempty"`
}

// CohereModelParametersDTO - DTO for serializing and storing Cohere model parameters.
type CohereModelParametersDTO struct { // skipcq: TCV-001
	Temperature      *float32 `json:"temperature,omitempty"`
	K                *uint32  `json:"k,omitempty"`
	P                *float32 `json:"p,omitempty"`
	FrequencyPenalty *float32 `json:"frequencyPenalty,omitempty"`
	PresencePenalty  *float32 `json:"presencePenalty,omitempty"`
	MaxTokens        *int32   `json:"maxTokens,omitempty"`
}

// PromptConfigDTO - DTO for serializing a prompt config.
type PromptConfigDTO struct { // skipcq: TCV-001
	ID                        string             `json:"id"`
	Name                      string             `json:"name"                      validate:"required"`
	ModelParameters           *json.RawMessage   `json:"modelParameters"           validate:"required"`
	ModelType                 models.ModelType   `json:"modelType"                 validate:"required"`
	ModelVendor               models.ModelVendor `json:"modelVendor"               validate:"oneof=OPEN_AI COHERE"`
	ProviderPromptMessages    *json.RawMessage   `json:"providerPromptMessages"    validate:"required"`
	ExpectedTemplateVariables []string           `json:"expectedTemplateVariables"`
	IsDefault                 bool               `json:"isDefault,omitempty"`
	CreatedAt                 time.Time          `json:"createdAt,omitempty"`
	UpdatedAt                 time.Time          `json:"updatedAt,omitempty"`
}

// ProviderModelPricingDTO is a data type used to encapsulate the pricing information for a model / type.
type ProviderModelPricingDTO struct { // skipcq: TCV-001
	ID               string          `json:"id"`
	InputTokenPrice  decimal.Decimal `json:"inputTokenPrice"`
	OutputTokenPrice decimal.Decimal `json:"outputTokenPrice"`
	TokenUnitSize    int32           `json:"tokenUnitSize"`
	ActiveFromDate   time.Time       `json:"activeFromDate"`
}
