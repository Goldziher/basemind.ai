// Package dto provides data transfer object definitions for the application.
package dto

import (
	"github.com/google/uuid"
)

// MessageRole represents the role of a message.
type MessageRole string

// Predefined message roles.
const (
	SystemRole MessageRole = "system"
	UserRole   MessageRole = "user"
)

// ProviderName represents the name of an AI provider.
type ProviderName string

// Predefined AI provider names.
const (
	OpenAIProvider      ProviderName = "openai"
	AzureOpenAIProvider ProviderName = "azure-openai"
	AnthropicProvider   ProviderName = "anthropic"
	GroqProvider        ProviderName = "groq"
)

// PromptMessageDefinition represents a single message in a prompt exchange.
type PromptMessageDefinition struct {
	// Type indicates if the message is from the user or the system.
	Type MessageRole `json:"type" validate:"required,oneof=system user"`
	// Content is the actual message content.
	Content string `json:"content" validate:"required"`
	// Constraints is a mapping of prompt variable keys to JSON schema objects.
	Constraints map[string]any `json:"constraints,omitempty"`
}

// ToolDefinition represents a tool that can be used by the LLM client.
type ToolDefinition struct {
	// Name is the identifier of the tool.
	Name string `json:"name" validate:"required"`
	// Description provides details about the tool's functionality.
	Description *string `json:"description,omitempty"`
	// Parameters define the input structure for the tool.
	Parameters map[string]interface{} `json:"parameters" validate:"required"`
}

// ModelConfig represents the configuration for an AI model.
type ModelConfig struct {
	// Name is the unique identifier for the model config.
	Name string `json:"name" validate:"required"`
	// ProviderName specifies the AI provider for this model.
	ProviderName ProviderName `json:"provider_name" validate:"required,oneof=openai azure-openai anthropic groq"`
	// APIKey is the authentication key for the AI service.
	APIKey string `json:"api_key" validate:"required"`
	// EndpointURL is the base URL for the API.
	EndpointURL string `json:"endpoint_url,omitempty" validate:"omitempty,url"`
	// MaxRetries specifies the maximum number of retry attempts for API calls.
	MaxRetries int `json:"max_retries,omitempty" validate:"omitempty,min=0"`
	// ExponentialBackoff indicates whether to use exponential backoff for retries.
	ExponentialBackoff bool `json:"exponential_backoff,omitempty"`
	// ClientParameters contains additional parameters for the AI client.
	ClientParameters map[string]any `json:"client_parameters,omitempty"`
}

// PromptConfig represents the configuration for a prompt.
type PromptConfig struct {
	// Name is the unique identifier for the prompt config.
	Name string `json:"name" validate:"required"`
	// ModelConfig specifies the AI model configuration to use.
	ModelConfig ModelConfig `json:"model_config" validate:"required"`
	// Description provides additional details about the prompt config.
	Description string `json:"description,omitempty"`
	// Messages is the template of prompt messages.
	Messages []PromptMessageDefinition `json:"messages" validate:"required,min=1,dive"`
	// ToolDefinition is a tool definition object.
	ToolDefinition ToolDefinition `json:"tool_definition,omitempty" validate:"omitempty"`
	// EnforceJSON specifies if the AI response must be a valid JSON object.
	EnforceJSON bool `json:"enforce_json,omitempty"`
	// ResponseValidationSchema is the JSON schema to validate the AI response.
	ResponseValidationSchema map[string]any `json:"validation_schema,omitempty"`
}

// PromptResultDTO encapsulates the result of a prompt request.
type PromptResultDTO struct {
	// ID is the unique identifier for the request record.
	ID uuid.UUID `json:"id" validate:"required"`
	// Content is the content of the response. It is nil if an error occurred.
	Content *string `json:"content,omitempty"`
	// Error is an error object, or nil if no error occurred.
	Error error `json:"error,omitempty"`
}

// RequestConfigurationDTO encapsulates the configuration for a prompt request.
type RequestConfigurationDTO struct {
	// PromptConfig is the prompt configuration.
	PromptConfig PromptConfig `json:"prompt_config" validate:"required"`
	// ExpectedTemplateVariables is a mapping of expected template variables to their JSON schema objects.
	ExpectedTemplateVariables map[string]any `json:"expected_template_variables" validate:"required"`
}
