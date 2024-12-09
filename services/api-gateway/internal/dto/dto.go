package dto

import (
	"github.com/basemind-ai/monorepo/shared/go/datatypes"
	"github.com/basemind-ai/monorepo/shared/go/db/models"
	"github.com/jackc/pgx/v5/pgtype"
)

// PromptResultDTO is a data type used to encapsulate the result of a prompt request.
type PromptResultDTO struct { // skipcq: TCV-001
	Content       *string
	Error         error
	RequestRecord *models.PromptRequestRecord
}

// RequestConfigurationDTO is a data type used encapsulate the current application prompt configuration.
type RequestConfigurationDTO struct { // skipcq: TCV-001
	// ProviderModelPricing is the pricing information for the model vendor
	ProviderModelPricing datatypes.ProviderModelPricingDTO `json:"providerModelPricing"`
	// PromptConfigData the prompt config DB record
	PromptConfigData datatypes.PromptConfigDTO `json:"promptConfigDTO"`
	// ApplicationID is the application DB ID
	ApplicationID pgtype.UUID `json:"applicationUUID"`
	// PromptConfigID is the promptConfig DB ID
	PromptConfigID pgtype.UUID `json:"promptConfigId,omitempty"`
}
