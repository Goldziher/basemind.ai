package dto

import (
	"encoding/json"
	"github.com/basemind-ai/monorepo/shared/go/db/models"
	"github.com/shopspring/decimal"
	"time"
)

// ApplicationDTO - DTO for serializing application data.
type ApplicationDTO struct { // skipcq: TCV-001
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// ProjectDTO - DTO for serializing project data.
type ProjectDTO struct { // skipcq: TCV-001
	ID           string           `json:"id,omitempty"`
	Name         string           `json:"name"                   validate:"required"`
	Description  string           `json:"description,omitempty"`
	Credits      decimal.Decimal  `json:"credits"`
	CreatedAt    time.Time        `json:"createdAt,omitempty"`
	UpdatedAt    time.Time        `json:"updatedAt,omitempty"`
	Permission   string           `json:"permission,omitempty"`
	Applications []ApplicationDTO `json:"applications,omitempty"`
}

// PromptConfigCreateDTO - DTO for prompt config CREATE request body.
type PromptConfigCreateDTO struct { // skipcq: TCV-001
	ModelParameters        *json.RawMessage   `json:"modelParameters" validate:"required"`
	ProviderPromptMessages *json.RawMessage   `json:"promptMessages"  validate:"required"`
	Name                   string             `json:"name"            validate:"required"`
	ModelType              models.ModelType   `json:"modelType"       validate:"oneof=gpt-3.5-turbo gpt-3.5-turbo-16k gpt-4 gpt-4-32k command command-light command-nightly command-light-nightly"`
	ModelVendor            models.ModelVendor `json:"modelVendor"     validate:"oneof=OPEN_AI COHERE"`
	IsTest                 bool               `json:"isTest"`
}

// PromptConfigUpdateDTO - DTO for prompt config UPDATE request body.
type PromptConfigUpdateDTO struct { // skipcq: TCV-001
	Name                   *string             `json:"name,omitempty"            validate:"omitempty,required"`
	ModelParameters        *json.RawMessage    `json:"modelParameters,omitempty" validate:"omitempty,required"`
	ModelType              *models.ModelType   `json:"modelType,omitempty"       validate:"omitempty,required"`
	ModelVendor            *models.ModelVendor `json:"modelVendor,omitempty"     validate:"omitempty,oneof=OPEN_AI COHERE"`
	ProviderPromptMessages *json.RawMessage    `json:"promptMessages,omitempty"  validate:"omitempty,required"`
}

// ApplicationAPIKeyDTO - DTO for serializing application api key data.
type ApplicationAPIKeyDTO struct { // skipcq: TCV-001
	CreatedAt time.Time `json:"createdAt"`
	Hash      *string   `json:"hash,omitempty"`
	ID        string    `json:"id"`
	Name      string    `json:"name"           validate:"required"`
}

// AddUserAccountToProjectDTO - DTO for add user to project request body.
type AddUserAccountToProjectDTO struct { // skipcq: TCV-001
	Email      string                      `json:"email,omitempty" validate:"omitempty,required"`
	Permission models.AccessPermissionType `json:"permission"      validate:"required,oneof=ADMIN MEMBER"`
}

// UpdateUserAccountProjectPermissionDTO - DTO for update user account project permission request body.
type UpdateUserAccountProjectPermissionDTO struct { // skipcq: TCV-001
	UserID     string                      `json:"userId"     validate:"required"`
	Permission models.AccessPermissionType `json:"permission" validate:"required,oneof=ADMIN MEMBER"`
}

// ProjectUserAccountDTO - DTO for serializing user account project + permission data.
type ProjectUserAccountDTO struct { // skipcq: TCV-001
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	Email       string    `json:"email"`
	FirebaseID  string    `json:"firebaseId"`
	PhoneNumber string    `json:"phoneNumber"`
	PhotoURL    string    `json:"photoUrl"`
	CreatedAt   time.Time `json:"createdAt"`
	Permission  string    `json:"permission"`
}

// AnalyticsDTO - DTO for serializing analytics data.
type AnalyticsDTO struct { // skipcq: TCV-001
	TokenCost     decimal.Decimal `json:"tokensCost"`
	TotalAPICalls int64           `json:"totalRequests"`
}

// PromptConfigTestDTO - DTO for requesting a prompt config test.
type PromptConfigTestDTO struct {
	ModelParameters        *json.RawMessage   `json:"modelParameters,omitempty"   validate:"omitempty,required"`
	ProviderPromptMessages *json.RawMessage   `json:"promptMessages,omitempty"    validate:"omitempty,required"`
	TemplateVariables      map[string]string  `json:"templateVariables,omitempty"`
	PromptConfigID         *string            `json:"promptConfigId,omitempty"    validate:"omitempty,required,uuid4"`
	ModelType              models.ModelType   `json:"modelType"                   validate:"oneof=gpt-3.5-turbo gpt-3.5-turbo-16k gpt-4 gpt-4-32k command command-light command-nightly command-light-nightly"`
	ModelVendor            models.ModelVendor `json:"modelVendor"                 validate:"oneof=OPEN_AI COHERE"`
}

// PromptConfigTestResultDTO - DTO for serializing prompt config test results.
type PromptConfigTestResultDTO struct { // skipcq: TCV-001
	Content            *string `json:"content,omitempty"`
	ErrorMessage       *string `json:"errorMessage,omitempty"`
	FinishReason       *string `json:"finishReason,omitempty"`
	PromptConfigID     *string `json:"promptConfigId,omitempty"`
	PromptTestRecordID *string `json:"promptTestRecordId,omitempty"`
}

// OtpDTO - DTO for serializing an OTP.
type OtpDTO struct { // skipcq: TCV-001
	OTP string `json:"otp"`
}

// SupportRequestDTO - DTO for a support request email.
type SupportRequestDTO struct { // skipcq: TCV-001
	RequestTopic string `json:"topic"     validate:"required"`
	EmailSubject string `json:"subject"   validate:"required"`
	EmailBody    string `json:"body"      validate:"required"`
	ProjectID    string `json:"projectId" validate:"required"`
}

// ProviderKeyCreateDTO - DTO for creating a provider key.
type ProviderKeyCreateDTO struct { // skipcq: TCV-001
	ModelVendor models.ModelVendor `json:"modelVendor" validate:"oneof=OPEN_AI COHERE"`
	Key         string             `json:"key"         validate:"required"`
}

type ProviderKeyDTO struct { // skipcq: TCV-001
	CreatedAt   time.Time          `json:"createdAt"`
	ID          string             `json:"id"`
	ModelVendor models.ModelVendor `json:"modelVendor"`
}

// PromptTestRecordDTO - DTO for serializing prompt test record data.
type PromptTestRecordDTO struct {
	CreatedAt              time.Time          `json:"createdAt"`
	StartTime              time.Time          `json:"startTime"`
	FinishTime             time.Time          `json:"finishTime"`
	PromptConfigID         *string            `json:"promptConfigId,omitempty"`
	ErrorLog               *string            `json:"errorLog,omitempty"`
	Name                   string             `json:"name"`
	RequestTokensCost      decimal.Decimal    `json:"requestTokensCost"`
	ModelVendor            models.ModelVendor `json:"modelVendor"`
	ID                     string             `json:"id"`
	ModelType              models.ModelType   `json:"modelType"`
	PromptResponse         string             `json:"promptResponse"`
	TotalTokensCost        decimal.Decimal    `json:"totalTokensCost"`
	ResponseTokensCost     decimal.Decimal    `json:"responseTokensCost"`
	ModelParameters        json.RawMessage    `json:"modelParameters"`
	ProviderPromptMessages json.RawMessage    `json:"providerPromptMessages"`
	UserInput              json.RawMessage    `json:"userInput"`
	ResponseTokens         int32              `json:"responseTokens"`
	RequestTokens          int32              `json:"requestTokens"`
	DurationMs             int32              `json:"durationMs"`
}

// ProjectInvitationDTO - DTO for serializing project invitation data.
type ProjectInvitationDTO struct { // skipcq: TCV-001
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Permission string    `json:"permission"`
}
