package internal_test

import (
	"context"
	"encoding/json"
	"github.com/basemind-ai/monorepo/e2e/factories"
	"github.com/basemind-ai/monorepo/gen/proto/v1"
	"github.com/basemind-ai/monorepo/internal/utils/datatypes"
	"github.com/basemind-ai/monorepo/internal/utils/exc"
	grpcutils2 "github.com/basemind-ai/monorepo/internal/utils/grpcutils"
	"github.com/basemind-ai/monorepo/internal/utils/ptr"
	"github.com/basemind-ai/monorepo/internal/utils/testutils"
	"github.com/basemind-ai/monorepo/services/api-gateway/internal/dto"
	"github.com/basemind-ai/monorepo/services/api-gateway/internal/services"
	"github.com/basemind-ai/monorepo/shared/go/db"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func createRequestConfigurationDTO(
	t *testing.T,
	projectID pgtype.UUID,
) dto.RequestConfigurationDTO {
	t.Helper()
	application, _ := factories.CreateApplication(context.TODO(), projectID)
	promptConfig, _ := factories.CreateOpenAIPromptConfig(
		context.TODO(),
		application.ID,
	)

	pricingModel := services.RetrieveProviderModelPricing(
		context.TODO(), promptConfig.ModelType, promptConfig.ModelVendor)

	return dto.RequestConfigurationDTO{
		ApplicationID:  application.ID,
		PromptConfigID: promptConfig.ID,
		PromptConfigData: datatypes.PromptConfigDTO{
			ID:                        db.UUIDToString(&promptConfig.ID),
			Name:                      promptConfig.Name,
			ModelType:                 promptConfig.ModelType,
			ModelVendor:               promptConfig.ModelVendor,
			ModelParameters:           ptr.To(json.RawMessage(promptConfig.ModelParameters)),
			ProviderPromptMessages:    ptr.To(json.RawMessage(promptConfig.ProviderPromptMessages)),
			ExpectedTemplateVariables: promptConfig.ExpectedTemplateVariables,
			IsDefault:                 promptConfig.IsDefault,
			CreatedAt:                 promptConfig.CreatedAt.Time,
			UpdatedAt:                 promptConfig.UpdatedAt.Time,
		},
		ProviderModelPricing: pricingModel,
	}
}

type mockGatewayServerStream struct {
	grpc.ServerStream
	Ctx      context.Context
	Response *gateway.StreamingPromptResponse
	Msg      any
	Error    error
}

func (m mockGatewayServerStream) Context() context.Context {
	if m.Ctx != nil {
		return m.Ctx
	}
	return context.TODO()
}

func (m mockGatewayServerStream) Send(response *gateway.StreamingPromptResponse) error {
	m.Response = response //nolint: all
	return m.Error
}

func (m mockGatewayServerStream) SendMsg(msg any) error {
	m.Msg = msg //nolint: all
	return m.Error
}

func createGatewayServiceClient(t *testing.T) gateway.gateway {
	t.Helper()
	listener := testutils.CreateTestGRPCServer[gateway.APIGatewayServiceServer](
		t,
		gateway.RegisterAPIGatewayServiceServer,
		services.APIGatewayServer{},
		// we are using the same auth middleware as we do on main.go in this test
		grpc.ChainUnaryInterceptor(
			auth.UnaryServerInterceptor(grpcutils2.NewAuthHandler(JWTSecret).HandleAuth),
		),
		grpc.ChainStreamInterceptor(
			auth.StreamServerInterceptor(grpcutils2.NewAuthHandler(JWTSecret).HandleAuth),
		),
	)
	return testutils.CreateTestGRPCClient[gateway.APIGatewayServiceClient](
		t,
		listener,
		gateway.NewAPIGatewayServiceClient,
	)
}

func TestAPIGatewayService(t *testing.T) {
	srv := services.APIGatewayServer{}
	project, _ := factories.CreateProject(context.TODO())
	_ = factories.CreateProviderPricingModels(context.TODO())

	requestConfigurationDTO := createRequestConfigurationDTO(t, project.ID)

	nonExistentPromptConfigID := "4503ec04-979c-42fe-a24f-15addd5c0509"

	createContext := func(applicationID pgtype.UUID) context.Context {
		return context.WithValue(context.WithValue(
			context.TODO(), grpcutils2.ProjectIDContextKey, project.ID,
		), grpcutils2.ApplicationIDContextKey, applicationID)
	}

	t.Run("RequestPrompt", func(t *testing.T) {
		t.Run("return error when projectID is not set in context", func(t *testing.T) {
			_, err := srv.RequestPrompt(context.TODO(), nil)
			assert.ErrorContains(t, err, services.ErrorProjectIDNotInContext)
		})

		t.Run("return error when applicationID is not set in context", func(t *testing.T) {
			_, err := srv.RequestPrompt(
				context.WithValue(context.TODO(), grpcutils2.ProjectIDContextKey, pgtype.UUID{}),
				nil,
			)
			assert.ErrorContains(t, err, services.ErrorApplicationIDNotInContext)
		})

		t.Run("returns error when a default prompt config is not found", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			_, mockRedis := createTestCache(
				t,
				db.UUIDToString(&application.ID),
			)
			mockRedis.ExpectGet(db.UUIDToString(&project.ID)).RedisNil()
			mockRedis.ExpectGet(db.UUIDToString(&application.ID)).RedisNil()

			_, err := srv.RequestPrompt(createContext(application.ID), &gateway.PromptRequest{})

			assert.ErrorContains(
				t,
				err,
				"the application does not have an active prompt configuration",
			)
		})

		t.Run(
			"returns error when a prompt config with a provided ID is not found",
			func(t *testing.T) {
				application, _ := factories.CreateApplication(
					context.TODO(),
					project.ID,
				)

				_, mockRedis := createTestCache(
					t,
					db.UUIDToString(&application.ID),
				)

				mockRedis.ExpectGet(db.UUIDToString(&application.ID)).RedisNil()

				_, err := srv.RequestPrompt(
					createContext(application.ID),
					&gateway.PromptRequest{PromptConfigId: &nonExistentPromptConfigID},
				)

				assert.ErrorContains(
					t,
					err,
					"the application does not have an active prompt configuration",
				)
			},
		)

		t.Run("returns error when template variables are not valid", func(t *testing.T) {
			cacheClient, mockRedis := createTestCache(
				t,
				db.UUIDToString(&requestConfigurationDTO.ApplicationID),
			)
			expectedCacheValue, marshalErr := cacheClient.Marshal(requestConfigurationDTO)
			assert.NoError(t, marshalErr)

			mockRedis.ExpectGet(db.UUIDToString(&requestConfigurationDTO.ApplicationID)).
				RedisNil()
			mockRedis.ExpectSet(db.UUIDToString(&requestConfigurationDTO.ApplicationID), expectedCacheValue, time.Hour/2).
				SetVal("OK")

			mockRedis.ExpectGet(db.UUIDToString(&project.ID)).
				RedisNil()
			mockRedis.ExpectSet(db.UUIDToString(&project.ID), exc.MustResult(cacheClient.Marshal(status.Status{})), time.Minute*5).
				SetVal("OK")

			mockRedis.ExpectGet(db.UUIDToString(&project.ID)).
				RedisNil()
			mockRedis.ExpectSet(db.UUIDToString(&project.ID), expectedCacheValue, time.Minute*5).
				RedisNil()

			_, err := srv.RequestPrompt(
				createContext(requestConfigurationDTO.ApplicationID),
				&gateway.PromptRequest{
					TemplateVariables: map[string]string{"name": "John"},
				},
			)

			assert.ErrorContains(t, err, "missing template variable")
		})
	})

	t.Run("RequestStreamingPrompt", func(t *testing.T) {
		t.Run("return error when projectID context value is not set", func(t *testing.T) {
			err := srv.RequestStreamingPrompt(nil, mockGatewayServerStream{})
			assert.ErrorContains(t, err, services.ErrorProjectIDNotInContext)
		})

		t.Run("return error when applicationID context value is not set", func(t *testing.T) {
			err := srv.RequestStreamingPrompt(nil, mockGatewayServerStream{
				Ctx: context.WithValue(
					context.TODO(),
					grpcutils2.ProjectIDContextKey,
					pgtype.UUID{},
				),
			})
			assert.ErrorContains(t, err, services.ErrorApplicationIDNotInContext)
		})

		t.Run("returns error when prompt config is not found", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			err := srv.RequestStreamingPrompt(
				&gateway.PromptRequest{},
				mockGatewayServerStream{Ctx: createContext(application.ID)},
			)
			assert.ErrorContains(
				t,
				err,
				"the application does not have an active prompt configuration",
			)
		})

		t.Run(
			"returns error when a prompt config with a provided ID is not found",
			func(t *testing.T) {
				_, mockRedis := createTestCache(
					t,
					db.UUIDToString(&requestConfigurationDTO.ApplicationID),
				)

				mockRedis.ExpectGet(db.UUIDToString(&requestConfigurationDTO.ApplicationID)).
					RedisNil()

				err := srv.RequestStreamingPrompt(
					&gateway.PromptRequest{PromptConfigId: &nonExistentPromptConfigID},
					mockGatewayServerStream{
						Ctx: createContext(requestConfigurationDTO.ApplicationID),
					},
				)

				assert.ErrorContains(
					t,
					err,
					"the application does not have an active prompt configuration",
				)
			},
		)

		t.Run("returns error when template variables are not valid", func(t *testing.T) {
			cacheClient, mockRedis := createTestCache(
				t,
				db.UUIDToString(&requestConfigurationDTO.ApplicationID),
			)
			expectedCacheValue, marshalErr := cacheClient.Marshal(requestConfigurationDTO)
			assert.NoError(t, marshalErr)

			mockRedis.ExpectGet(db.UUIDToString(&requestConfigurationDTO.ApplicationID)).
				RedisNil()
			mockRedis.ExpectSet(db.UUIDToString(&requestConfigurationDTO.ApplicationID), expectedCacheValue, time.Hour/2).
				SetVal("OK")

			mockRedis.ExpectGet(db.UUIDToString(&project.ID)).
				RedisNil()
			mockRedis.ExpectSet(db.UUIDToString(&project.ID), exc.MustResult(cacheClient.Marshal(status.Status{})), time.Minute*5).
				SetVal("OK")

			err := srv.RequestStreamingPrompt(&gateway.PromptRequest{
				TemplateVariables: map[string]string{"name": "John"},
			}, mockGatewayServerStream{Ctx: createContext(requestConfigurationDTO.ApplicationID)})

			assert.ErrorContains(t, err, "missing template variables")
		})
	})
}
