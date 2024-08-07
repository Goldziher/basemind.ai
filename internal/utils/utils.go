package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/basemind-ai/monorepo/gen/proto/v1"
	"github.com/basemind-ai/monorepo/internal/utils/config"
	"github.com/basemind-ai/monorepo/internal/utils/cryptoutils"
	"github.com/basemind-ai/monorepo/internal/utils/datatypes"
	"github.com/basemind-ai/monorepo/internal/utils/exc"
	"github.com/basemind-ai/monorepo/internal/utils/grpcutils"
	"github.com/basemind-ai/monorepo/internal/utils/ptr"
	"github.com/basemind-ai/monorepo/internal/utils/rediscache"
	"github.com/basemind-ai/monorepo/services/api-gateway/internal/dto"
	"github.com/basemind-ai/monorepo/shared/go/db"
	"github.com/basemind-ai/monorepo/shared/go/db/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

// RetrievePromptConfig retrieves the prompt config - either using the provided ID, or the application default.
func RetrievePromptConfig(
	ctx context.Context,
	applicationID pgtype.UUID,
	promptConfigID *string,
) (*datatypes.PromptConfigDTO, error) {
	if promptConfigID != nil {
		uuid, uuidErr := db.StringToUUID(*promptConfigID)
		if uuidErr != nil {
			return nil, status.Errorf(
				codes.InvalidArgument, "invalid prompt-config id: %v", uuidErr,
			)
		}

		promptConfig, retrievalErr := db.GetQueries().RetrievePromptConfig(ctx, *uuid)
		if retrievalErr != nil {
			return nil, fmt.Errorf("failed to retrieve prompt config - %w", retrievalErr)
		}

		return &datatypes.PromptConfigDTO{
			ID:                        db.UUIDToString(&promptConfig.ID),
			Name:                      promptConfig.Name,
			ModelParameters:           ptr.To(json.RawMessage(promptConfig.ModelParameters)),
			ModelType:                 promptConfig.ModelType,
			ModelVendor:               promptConfig.ModelVendor,
			ProviderPromptMessages:    ptr.To(json.RawMessage(promptConfig.ProviderPromptMessages)),
			ExpectedTemplateVariables: promptConfig.ExpectedTemplateVariables,
			IsDefault:                 promptConfig.IsDefault,
			CreatedAt:                 promptConfig.CreatedAt.Time,
			UpdatedAt:                 promptConfig.UpdatedAt.Time,
		}, nil
	}

	promptConfig, retrieveDefaultErr := db.
		GetQueries().
		RetrieveDefaultPromptConfig(ctx, applicationID)

	if retrieveDefaultErr != nil {
		return nil, fmt.Errorf(
			"failed to retrieve default prompt config - %w",
			retrieveDefaultErr,
		)
	}
	return &datatypes.PromptConfigDTO{
		ID:                        db.UUIDToString(&promptConfig.ID),
		Name:                      promptConfig.Name,
		ModelParameters:           ptr.To(json.RawMessage(promptConfig.ModelParameters)),
		ModelType:                 promptConfig.ModelType,
		ModelVendor:               promptConfig.ModelVendor,
		ProviderPromptMessages:    ptr.To(json.RawMessage(promptConfig.ProviderPromptMessages)),
		ExpectedTemplateVariables: promptConfig.ExpectedTemplateVariables,
		IsDefault:                 promptConfig.IsDefault,
		CreatedAt:                 promptConfig.CreatedAt.Time,
		UpdatedAt:                 promptConfig.UpdatedAt.Time,
	}, nil
}

// RetrieveProviderModelPricing retrieves the pricing model for the given model type and vendor.
func RetrieveProviderModelPricing(
	ctx context.Context,
	modelType models.ModelType,
	modelVendor models.ModelVendor,
) datatypes.ProviderModelPricingDTO {
	log.Debug().
		Str("modelType", string(modelType)).
		Str("modelVendor", string(modelVendor)).
		Msg("retrieving provider model pricing")

	providerModelPricing := exc.MustResult(db.GetQueries().
		RetrieveActiveProviderModelPricing(ctx, models.RetrieveActiveProviderModelPricingParams{
			ModelType:   modelType,
			ModelVendor: modelVendor,
		}))

	inputDecimalValue := exc.MustResult(db.NumericToDecimal(providerModelPricing.InputTokenPrice))
	outputDecimalValue := exc.MustResult(db.NumericToDecimal(providerModelPricing.OutputTokenPrice))

	return datatypes.ProviderModelPricingDTO{
		ID:               db.UUIDToString(&providerModelPricing.ID),
		InputTokenPrice:  *inputDecimalValue,
		OutputTokenPrice: *outputDecimalValue,
		TokenUnitSize:    providerModelPricing.TokenUnitSize,
		ActiveFromDate:   providerModelPricing.ActiveFromDate.Time,
	}
}

// RetrieveRequestConfiguration retrieves the request configuration for the given application and prompt config ID.
func RetrieveRequestConfiguration(
	ctx context.Context,
	applicationID pgtype.UUID,
	promptConfigID *string,
) func() (*dto.RequestConfigurationDTO, error) {
	return func() (*dto.RequestConfigurationDTO, error) {
		application, applicationQueryErr := db.GetQueries().RetrieveApplication(ctx, applicationID)
		if applicationQueryErr != nil {
			return nil, status.Errorf(
				codes.NotFound,
				"application does not exist: %v",
				applicationQueryErr,
			)
		}

		promptConfig, retrievalError := RetrievePromptConfig(
			ctx,
			applicationID,
			promptConfigID,
		)
		if retrievalError != nil {
			return nil, status.Errorf(
				codes.NotFound,
				"the application does not have an active prompt configuration: %v",
				retrievalError,
			)
		}

		promptConfigUUID := exc.MustResult(db.StringToUUID(promptConfig.ID))

		return &dto.RequestConfigurationDTO{
			ApplicationID:    application.ID,
			PromptConfigID:   *promptConfigUUID,
			PromptConfigData: *promptConfig,
			ProviderModelPricing: RetrieveProviderModelPricing(
				ctx, promptConfig.ModelType, promptConfig.ModelVendor,
			),
		}, nil
	}
}

// CheckProjectCredits checks that the project has sufficient credits.
func CheckProjectCredits(
	ctx context.Context,
	projectID pgtype.UUID,
) func() (*status.Status, error) {
	return func() (*status.Status, error) {
		project, projectQueryErr := db.GetQueries().RetrieveProject(ctx, projectID)
		if projectQueryErr != nil {
			return &status.Status{}, fmt.Errorf("failed to retrieve project - %w", projectQueryErr)
		}

		decimalValue := exc.MustResult(db.NumericToDecimal(project.Credits))

		if decimalValue.IsNegative() || decimalValue.IsZero() {
			return status.New(
				codes.ResourceExhausted,
				ErrorInsufficientCredits,
			), nil
		}

		return &status.Status{}, nil
	}
}

// CreateProviderAPIKeyContext creates a context with the provider API key.
// The provider API key is retrieved from the database, decrypted and set in the context.
// We intentionally use the projectID to cache here - because we need to invalidate the cache if the provider api key is deleted.
func CreateProviderAPIKeyContext(
	ctx context.Context,
	projectID pgtype.UUID,
	modelVendor models.ModelVendor,
) context.Context {
	providerKey, providerKeyRetrievalErr := rediscache.With[models.RetrieveProviderKeyRow](
		ctx,
		db.UUIDToString(&projectID),
		&models.RetrieveProviderKeyRow{},
		30*time.Minute,
		func() (*models.RetrieveProviderKeyRow, error) {
			providerKey, retrievalErr := db.GetQueries().RetrieveProviderKey(
				ctx,
				models.RetrieveProviderKeyParams{
					ProjectID:   projectID,
					ModelVendor: modelVendor,
				},
			)
			return &providerKey, retrievalErr
		},
	)
	if providerKeyRetrievalErr != nil {
		log.Debug().Err(providerKeyRetrievalErr).Msg("provider key is not set")
		return ctx
	}

	decryptedKey := cryptoutils.Decrypt(providerKey.EncryptedApiKey, config.Get(ctx).CryptoPassKey)

	// we append the encrypted provider key to the outgoing context
	// it will be retrieved by the connector.
	return metadata.AppendToOutgoingContext(ctx, "X-API-Key", decryptedKey)
}

// ValidateExpectedVariables validates that the expected template variables are present in the request.
func ValidateExpectedVariables(
	templateVariables map[string]string,
	expectedVariables []string,
) error {
	var missingVariables []string

	for _, expectedVariable := range expectedVariables {
		if _, ok := templateVariables[expectedVariable]; !ok {
			missingVariables = append(missingVariables, expectedVariable)
		}
	}

	if len(missingVariables) > 0 {
		log.Debug().
			Interface("missingVariables", missingVariables).
			Msg("missing template variables")
		return status.Errorf(
			codes.InvalidArgument,
			"missing template variables: %v",
			missingVariables,
		)
	}

	return nil
}

// StreamFromChannel streams the prompt results from the channel to the stream server.
func StreamFromChannel[T any](
	ctx context.Context,
	channel chan dto.PromptResultDTO,
	streamServer grpc.ServerStream,
	messageFactory func(context.Context, dto.PromptResultDTO) (T, bool),
) error {
	for {
		select {
		case result, isOpen := <-channel:
			msg, isFinished := messageFactory(ctx, result)

			if sendErr := streamServer.SendMsg(msg); sendErr != nil {
				log.Error().Err(sendErr).Msg("failed to send message")
				return status.Error(codes.Internal, "failed to send message")
			}

			if result.Error != nil {
				log.Error().Err(result.Error).Msg("error in prompt request")
				return status.Error(codes.Internal, "error communicating with AI provider")
			}

			if isFinished || !isOpen {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// CreateAPIGatewayStreamMessage creates a stream message for the API Gateway.
func CreateAPIGatewayStreamMessage(
	ctx context.Context,
	result dto.PromptResultDTO,
) (*gateway.gateway, bool) {
	msg := &gateway.StreamingPromptResponse{}

	isFinished := false

	if result.Error != nil {
		isFinished = true
		reason := "ERROR"
		msg.FinishReason = &reason
	} else if result.RequestRecord != nil {
		isFinished = true
		reason := string(result.RequestRecord.FinishReason)
		msg.FinishReason = &reason
	}

	if result.RequestRecord != nil {
		requestTokens := uint32(result.RequestRecord.RequestTokens)
		responseTokens := uint32(result.RequestRecord.ResponseTokens)
		streamDuration := uint32(
			result.RequestRecord.FinishTime.Time.
				Sub(result.RequestRecord.StartTime.Time).
				Milliseconds(),
		)
		msg.RequestTokens = &requestTokens
		msg.ResponseTokens = &responseTokens
		msg.StreamDuration = &streamDuration

		go DeductCredit(ctx, result.RequestRecord)
	}

	if content := ptr.Deref(result.Content, ""); len(content) > 0 {
		msg.Content = content
	}

	return msg, isFinished
}

// DeductCredit deducts the credit from the project.
func DeductCredit(
	ctx context.Context, requestRecord *models.PromptRequestRecord,
) {
	projectID, ok := ctx.Value(grpcutils.ProjectIDContextKey).(pgtype.UUID)
	if !ok {
		log.Error().Msg("project id not in context")
		return
	}

	requestCost := *exc.MustResult(db.NumericToDecimal(requestRecord.RequestTokensCost))
	responseCost := exc.MustResult(db.NumericToDecimal(requestRecord.ResponseTokensCost))
	totalCost := responseCost.Add(requestCost).Mul(decimal.NewFromInt(-1))

	exc.LogIfErr(
		db.GetQueries().
			UpdateProjectCredits(context.Background(), models.UpdateProjectCreditsParams{
				ID:      projectID,
				Credits: *exc.MustResult(db.StringToNumeric(totalCost.String())),
			}),
	)
}
