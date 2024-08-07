package internal

import (
	"context"
	"fmt"
	"github.com/basemind-ai/monorepo/gen/proto/v1"
	"github.com/basemind-ai/monorepo/internal/utils"
	"github.com/basemind-ai/monorepo/internal/utils/grpcutils"
	"github.com/basemind-ai/monorepo/internal/utils/rediscache"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type APIGatewayServer struct {
	gateway.gateway
}

func (APIGatewayServer) RequestPrompt(
	ctx context.Context,
	request *gateway.PromptRequest,
) (*gateway.PromptResponse, error) {
	cacheKey := db.UUIDToString(&applicationID)
	if request.PromptConfigId != nil {
		cacheKey = fmt.Sprintf("%s:%s", db.UUIDToString(&applicationID), *request.PromptConfigId)
	}

	requestConfigurationDTO, retrievalErr := rediscache.With[dto.RequestConfigurationDTO](
		ctx,
		cacheKey,
		&dto.RequestConfigurationDTO{},
		time.Minute*30,
		utils.RetrieveRequestConfiguration(ctx, applicationID, request.PromptConfigId),
	)
	if retrievalErr != nil {
		log.Error().Err(retrievalErr).Msg("failed to retrieve the request configuration from Redis")
		return nil, status.Error(
			codes.NotFound,
			retrievalErr.Error(),
		)
	}

	if validationError := utils.ValidateExpectedVariables(request.TemplateVariables, requestConfigurationDTO.PromptConfigData.ExpectedTemplateVariables); validationError != nil {
		// the validation error is already a grpc status error
		return nil, validationError
	}

	providerKeyContext := utils.CreateProviderAPIKeyContext(
		ctx,
		projectID,
		requestConfigurationDTO.PromptConfigData.ModelVendor,
	)

	if promptResult.Error != nil {
		log.Error().Err(promptResult.Error).Msg("error in prompt request")
		return nil, status.Error(codes.Internal, "error communicating with AI provider")
	}

	return &gateway.PromptResponse{
		Content:        *promptResult.Content,
		RequestTokens:  uint32(promptResult.RequestRecord.RequestTokens),
		ResponseTokens: uint32(promptResult.RequestRecord.ResponseTokens),
	}, nil
}

func (APIGatewayServer) RequestStreamingPrompt(
	request *gateway.PromptRequest,
	streamServer gateway.APIGatewayService_RequestStreamingPromptServer,
) error {
	projectID, ok := streamServer.Context().Value(grpcutils.ProjectIDContextKey).(pgtype.UUID)
	if !ok {
		return status.Errorf(codes.Unauthenticated, ErrorProjectIDNotInContext)
	}

	applicationID, ok := streamServer.Context().Value(grpcutils.ApplicationIDContextKey).(pgtype.UUID)
	if !ok {
		return status.Errorf(codes.Unauthenticated, ErrorApplicationIDNotInContext)
	}

	cacheKey := db.UUIDToString(&applicationID)
	if request.PromptConfigId != nil {
		cacheKey = fmt.Sprintf("%s:%s", db.UUIDToString(&applicationID), *request.PromptConfigId)
	}

	requestConfigurationDTO, retrievalErr := rediscache.With[dto.RequestConfigurationDTO](
		streamServer.Context(),
		cacheKey,
		&dto.RequestConfigurationDTO{},
		time.Minute*30,
		utils.RetrieveRequestConfiguration(streamServer.Context(), applicationID, request.PromptConfigId),
	)
	if retrievalErr != nil {
		log.Error().Err(retrievalErr).Msg("failed to retrieve the request configuration from Redis")
		return status.Error(
			codes.NotFound,
			retrievalErr.Error(),
		)
	}

	if insufficientCreditsErr, retrievalErr := rediscache.With[status.Status](
		streamServer.Context(),
		db.UUIDToString(&projectID),
		&status.Status{},
		time.Minute*5,
		utils.CheckProjectCredits(streamServer.Context(), projectID),
	); retrievalErr != nil {
		return retrievalErr
	} else if insufficientCreditsErr.Code() == codes.ResourceExhausted {
		return insufficientCreditsErr.Err()
	}

	if validationError := utils.ValidateExpectedVariables(request.TemplateVariables, requestConfigurationDTO.PromptConfigData.ExpectedTemplateVariables); validationError != nil {
		// the validation error is already a grpc status error
		return validationError
	}

	providerKeyContext := utils.CreateProviderAPIKeyContext(
		streamServer.Context(),
		projectID,
		requestConfigurationDTO.PromptConfigData.ModelVendor,
	)

	channel := make(chan dto.PromptResultDTO)

	go connectors.GetProviderConnector(requestConfigurationDTO.PromptConfigData.ModelVendor).
		RequestStream(
			providerKeyContext,
			requestConfigurationDTO,
			request.TemplateVariables,
			channel,
		)

	return utils.StreamFromChannel(
		streamServer.Context(),
		channel,
		streamServer,
		utils.CreateAPIGatewayStreamMessage,
	)
}
