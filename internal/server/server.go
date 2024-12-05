package server

import (
	"context"
	"fmt"
	"github.com/basemind-ai/gateway/gen/gateway/v2"
	"github.com/basemind-ai/gateway/internal/dto"
	"github.com/basemind-ai/gateway/internal/utils"
	"github.com/basemind-ai/gateway/internal/utils/grpcutils"
	"github.com/basemind-ai/gateway/internal/utils/rediscache"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type APIGatewayServer struct {
	gateway.APIGatewayServiceServer
}

func (APIGatewayServer) RequestPrompt(
	ctx context.Context,
	request *gateway.PromptRequest,
) (*gateway.PromptResponse, error) {
	config := dto.RequestConfigurationDTO{}

	if validationError := utils.ValidateExpectedVariables(request.TemplateVariables, config); validationError != nil {
		// the validation error is already a grpc status error
		return nil, validationError
	}

	providerKeyContext := utils.CreateProviderAPIKeyContext(
		ctx,
		config.PromptConfig.ModelConfig,
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
