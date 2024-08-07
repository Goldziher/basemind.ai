package dto

import (
	"github.com/basemind-ai/monorepo/internal/utils/datatypes"
	"github.com/basemind-ai/monorepo/shared/go/db/models"
	"github.com/jackc/pgx/v5/pgtype"
)

type PromptConfig struct {
	ID                        pgtype.UUID        `json:"id"`
	Name                      string             `json:"name"`
	ModelParameters           []byte             `json:"modelParameters"`
	ModelType                 ModelType          `json:"modelType"`
	ModelVendor               ModelVendor        `json:"modelVendor"`
	ProviderPromptMessages    []byte             `json:"providerPromptMessages"`
	ExpectedTemplateVariables []string           `json:"expectedTemplateVariables"`
	IsDefault                 bool               `json:"isDefault"`
	IsTestConfig              bool               `json:"isTestConfig"`
	CreatedAt                 pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt                 pgtype.Timestamptz `json:"updatedAt"`
	DeletedAt                 pgtype.Timestamptz `json:"deletedAt"`
	ApplicationID             pgtype.UUID        `json:"applicationId"`
}

type PromptRequestRecord struct {
	ID                     pgtype.UUID        `json:"id"`
	IsStreamResponse       bool               `json:"isStreamResponse"`
	RequestTokens          int32              `json:"requestTokens"`
	ResponseTokens         int32              `json:"responseTokens"`
	RequestTokensCost      pgtype.Numeric     `json:"requestTokensCost"`
	ResponseTokensCost     pgtype.Numeric     `json:"responseTokensCost"`
	StartTime              pgtype.Timestamptz `json:"startTime"`
	FinishTime             pgtype.Timestamptz `json:"finishTime"`
	FinishReason           PromptFinishReason `json:"finishReason"`
	DurationMs             pgtype.Int4        `json:"durationMs"`
	PromptConfigID         pgtype.UUID        `json:"promptConfigId"`
	ErrorLog               pgtype.Text        `json:"errorLog"`
	CreatedAt              pgtype.Timestamptz `json:"createdAt"`
	DeletedAt              pgtype.Timestamptz `json:"deletedAt"`
	ProviderModelPricingID pgtype.UUID        `json:"providerModelPricingId"`
}

type PromptTestRecord struct {
	ID                    pgtype.UUID        `json:"id"`
	VariableValues        []byte             `json:"variableValues"`
	Response              string             `json:"response"`
	CreatedAt             pgtype.Timestamptz `json:"createdAt"`
	PromptRequestRecordID pgtype.UUID        `json:"promptRequestRecordId"`
}

type ProviderKey struct {
	ID              pgtype.UUID        `json:"id"`
	ModelVendor     ModelVendor        `json:"modelVendor"`
	EncryptedApiKey string             `json:"encryptedApiKey"`
	CreatedAt       pgtype.Timestamptz `json:"createdAt"`
	ProjectID       pgtype.UUID        `json:"projectId"`
}

type ProviderModelPricing struct {
	ID               pgtype.UUID        `json:"id"`
	ModelType        ModelType          `json:"modelType"`
	ModelVendor      ModelVendor        `json:"modelVendor"`
	InputTokenPrice  pgtype.Numeric     `json:"inputTokenPrice"`
	OutputTokenPrice pgtype.Numeric     `json:"outputTokenPrice"`
	TokenUnitSize    int32              `json:"tokenUnitSize"`
	CreatedAt        pgtype.Timestamptz `json:"createdAt"`
	ActiveFromDate   pgtype.Date        `json:"activeFromDate"`
	ActiveToDate     pgtype.Date        `json:"activeToDate"`
}

type UserAccount struct {
	ID          pgtype.UUID        `json:"id"`
	DisplayName string             `json:"displayName"`
	Email       string             `json:"email"`
	FirebaseID  string             `json:"firebaseId"`
	PhoneNumber string             `json:"phoneNumber"`
	PhotoUrl    string             `json:"photoUrl"`
	CreatedAt   pgtype.Timestamptz `json:"createdAt"`
}

type UserProject struct {
	UserID     pgtype.UUID          `json:"userId"`
	ProjectID  pgtype.UUID          `json:"projectId"`
	Permission AccessPermissionType `json:"permission"`
	CreatedAt  pgtype.Timestamptz   `json:"createdAt"`
	UpdatedAt  pgtype.Timestamptz   `json:"updatedAt"`
}

// PromptResultDTO is a data type used to encapsulate the result of a prompt request.
type PromptResultDTO struct { // skipcq: TCV-001
	Content       *string
	Error         error
	RequestRecord *models.PromptRequestRecord
}

// RequestConfigurationDTO is a data type used encapsulate the current application prompt configuration.
type RequestConfigurationDTO struct { // skipcq: TCV-001
	// ApplicationID is the application DB ID
	ApplicationID pgtype.UUID `json:"applicationUUID"`
	// PromptConfigID is the promptConfig DB ID
	PromptConfigID pgtype.UUID `json:"promptConfigId,omitempty"`
	// PromptConfigData the prompt config DB record
	PromptConfigData datatypes.PromptConfigDTO `json:"promptConfigDTO"`
	// ProviderModelPricing is the pricing information for the model vendor
	ProviderModelPricing datatypes.ProviderModelPricingDTO `json:"providerModelPricing"`
}
