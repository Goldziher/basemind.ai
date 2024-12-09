// Update prompt config name
package repositories_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/basemind-ai/monorepo/shared/go/db/models"
	"github.com/basemind-ai/monorepo/shared/go/ptr"
	"github.com/basemind-ai/monorepo/shared/go/serialization"
	"github.com/shopspring/decimal"
	"testing"
	"time"

	"github.com/basemind-ai/monorepo/e2e/factories"
	"github.com/basemind-ai/monorepo/services/dashboard-backend/internal/dto"
	"github.com/basemind-ai/monorepo/services/dashboard-backend/internal/repositories"
	"github.com/basemind-ai/monorepo/shared/go/db"
	"github.com/basemind-ai/monorepo/shared/go/testutils"
	"github.com/stretchr/testify/assert"
)

func TestPromptConfigRepository(t *testing.T) { //nolint: revive
	newModelParameters := ptr.To(
		json.RawMessage(serialization.SerializeJSON(map[string]any{"maxTokens": 100})),
	)

	newName := "new name"
	newModelType := models.ModelTypeGpt4

	newSystemMessage := "Your role is {role}"
	newUserMessage := "Describe what it is to be a {role} in your experience"

	templateVariables := []string{"role"}
	msgs := factories.CreateOpenAIPromptMessages(
		newSystemMessage, newUserMessage, nil,
	)
	newPromptMessages := msgs

	project, _ := factories.CreateProject(context.TODO())

	redisDB, redisMock := testutils.CreateMockRedisClient(t)

	t.Run("CreatePromptConfig", func(t *testing.T) {
		t.Run("creates prompt config", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)

			createPromptConfigDTO := dto.PromptConfigCreateDTO{
				Name:                   "test",
				ModelVendor:            models.ModelVendorOPENAI,
				ModelType:              models.ModelTypeGpt432k,
				ModelParameters:        newModelParameters,
				ProviderPromptMessages: newPromptMessages,
			}
			promptConfig, err := repositories.CreatePromptConfig(
				context.TODO(),
				application.ID,
				createPromptConfigDTO,
			)
			assert.NoError(t, err)
			assert.NotNil(t, promptConfig)

			assert.Equal(t, createPromptConfigDTO.Name, promptConfig.Name)
			assert.Equal(t, createPromptConfigDTO.ModelVendor, promptConfig.ModelVendor)
			assert.Equal(t, createPromptConfigDTO.ModelType, promptConfig.ModelType)
			assert.Equal(t, createPromptConfigDTO.ModelParameters, promptConfig.ModelParameters)
			assert.Equal(t, templateVariables, promptConfig.ExpectedTemplateVariables)
			assert.True(t, promptConfig.IsDefault)
			assert.NotEmpty(t, promptConfig.ID)
			assert.NotEmpty(t, promptConfig.CreatedAt)
			assert.NotEmpty(t, promptConfig.UpdatedAt)
		})
		t.Run("does not set default to true if default already exists", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			_, _ = factories.CreateOpenAIPromptConfig(context.TODO(), application.ID)

			createPromptConfigDTO := dto.PromptConfigCreateDTO{
				Name:                   "test",
				ModelVendor:            models.ModelVendorOPENAI,
				ModelType:              models.ModelTypeGpt432k,
				ModelParameters:        newModelParameters,
				ProviderPromptMessages: newPromptMessages,
			}
			promptConfig, err := repositories.CreatePromptConfig(
				context.TODO(),
				application.ID,
				createPromptConfigDTO,
			)
			assert.NoError(t, err)
			assert.NotNil(t, promptConfig)

			assert.False(t, promptConfig.IsDefault)
		})
		t.Run("returns error if failed to parse prompt messages", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			createPromptConfigDTO := dto.PromptConfigCreateDTO{
				Name:                   "test",
				ModelVendor:            models.ModelVendorOPENAI,
				ModelType:              models.ModelTypeGpt432k,
				ModelParameters:        newModelParameters,
				ProviderPromptMessages: ptr.To(json.RawMessage("invalid")),
			}
			promptConfig, err := repositories.CreatePromptConfig(
				context.TODO(),
				application.ID,
				createPromptConfigDTO,
			)
			assert.Error(t, err)
			assert.Nil(t, promptConfig)
		})

		t.Run("returns error if prompt config name collides", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			existingPromptConfig, _ := factories.CreateOpenAIPromptConfig(
				context.TODO(),
				application.ID,
			)

			createPromptConfigDTO := dto.PromptConfigCreateDTO{
				Name:                   existingPromptConfig.Name,
				ModelVendor:            models.ModelVendorOPENAI,
				ModelType:              models.ModelTypeGpt432k,
				ModelParameters:        newModelParameters,
				ProviderPromptMessages: newPromptMessages,
			}
			promptConfig, err := repositories.CreatePromptConfig(
				context.TODO(),
				application.ID,
				createPromptConfigDTO,
			)
			assert.Error(t, err)
			assert.Nil(t, promptConfig)
		})
	})

	t.Run("UpdateApplicationDefaultPromptConfig", func(t *testing.T) {
		t.Run("updates application default prompt config", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			_, _ = factories.CreateOpenAIPromptConfig(context.TODO(), application.ID)

			createPromptConfigDTO := dto.PromptConfigCreateDTO{
				Name:                   "test",
				ModelVendor:            models.ModelVendorOPENAI,
				ModelType:              models.ModelTypeGpt432k,
				ModelParameters:        newModelParameters,
				ProviderPromptMessages: newPromptMessages,
			}
			promptConfig, _ := repositories.CreatePromptConfig(
				context.TODO(),
				application.ID,
				createPromptConfigDTO,
			)

			assert.False(t, promptConfig.IsDefault)

			dbID, _ := db.StringToUUID(promptConfig.ID)

			err := repositories.UpdateApplicationDefaultPromptConfig(
				context.TODO(),
				application.ID,
				*dbID,
			)
			assert.NoError(t, err)

			defaultPromptConfig, _ := db.GetQueries().
				RetrieveDefaultPromptConfig(context.TODO(), application.ID)

			assert.Equal(t, promptConfig.ID, db.UUIDToString(&defaultPromptConfig.ID))
		})

		t.Run("invalidates prompt-config caches", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			defaultPromptConfig, _ := factories.CreateOpenAIPromptConfig(
				context.TODO(),
				application.ID,
			)

			createPromptConfigDTO := dto.PromptConfigCreateDTO{
				Name:                   "test",
				ModelVendor:            models.ModelVendorOPENAI,
				ModelType:              models.ModelTypeGpt432k,
				ModelParameters:        newModelParameters,
				ProviderPromptMessages: newPromptMessages,
			}
			nonDefaultPromptConfig, _ := repositories.CreatePromptConfig(
				context.TODO(),
				application.ID,
				createPromptConfigDTO,
			)

			applicationID := db.UUIDToString(&application.ID)

			cacheKeys := []string{
				fmt.Sprintf("%s:%s", applicationID, db.UUIDToString(&defaultPromptConfig.ID)),
				fmt.Sprintf("%s:%s", applicationID, nonDefaultPromptConfig.ID),
				applicationID,
			}

			for _, cacheKey := range cacheKeys {
				redisDB.Set(context.TODO(), cacheKey, "test", 0)
				redisMock.ExpectDel(cacheKey).SetVal(1)
			}

			dbID, _ := db.StringToUUID(nonDefaultPromptConfig.ID)

			err := repositories.UpdateApplicationDefaultPromptConfig(
				context.TODO(),
				application.ID,
				*dbID,
			)
			assert.NoError(t, err)

			time.Sleep(testutils.GetSleepTimeout())

			assert.NoError(t, redisMock.ExpectationsWereMet())
		})

		t.Run(
			"returns error if fails to retrieve the default prompt config",
			func(t *testing.T) {
				application, _ := factories.CreateApplication(context.TODO(), project.ID)
				promptConfig, _ := factories.CreateOpenAIPromptConfig(
					context.TODO(),
					application.ID,
				)

				_ = db.GetQueries().DeletePromptConfig(context.TODO(), promptConfig.ID)

				err := repositories.UpdateApplicationDefaultPromptConfig(
					context.TODO(),
					application.ID,
					promptConfig.ID,
				)
				assert.Error(t, err)
			},
		)

		t.Run(
			"returns error if the passed in promptConfigID is for the app default",
			func(t *testing.T) {
				application, _ := factories.CreateApplication(context.TODO(), project.ID)
				promptConfig, _ := factories.CreateOpenAIPromptConfig(
					context.TODO(),
					application.ID,
				)

				err := repositories.UpdateApplicationDefaultPromptConfig(
					context.TODO(),
					application.ID,
					promptConfig.ID,
				)
				assert.Error(t, err)
			},
		)
	})

	t.Run("UpdatePromptConfig", func(t *testing.T) {
		t.Run("updates prompt config", func(t *testing.T) {
			testCases := []struct {
				Dto  dto.PromptConfigUpdateDTO
				Name string
			}{
				{
					Name: "updates prompt config name",
					Dto:  dto.PromptConfigUpdateDTO{Name: &newName},
				},
				{
					Name: "updates prompt config model type",
					Dto:  dto.PromptConfigUpdateDTO{ModelType: &newModelType},
				},
				{
					Name: "updates prompt config model parameters",
					Dto:  dto.PromptConfigUpdateDTO{ModelParameters: newModelParameters},
				},
				{
					Name: "updates prompt config prompt messages",
					Dto:  dto.PromptConfigUpdateDTO{ProviderPromptMessages: newPromptMessages},
				},
				{
					Name: "updates multiple values",
					Dto: dto.PromptConfigUpdateDTO{
						Name:                   &newName,
						ModelType:              &newModelType,
						ModelParameters:        newModelParameters,
						ProviderPromptMessages: newPromptMessages,
					},
				},
			}

			for _, testCase := range testCases {
				t.Run(testCase.Name, func(t *testing.T) {
					application, _ := factories.CreateApplication(context.TODO(), project.ID)
					promptConfig, _ := factories.CreateOpenAIPromptConfig(
						context.TODO(),
						application.ID,
					)

					updatedPromptConfig, err := repositories.UpdatePromptConfig(
						context.TODO(),
						promptConfig.ID,
						testCase.Dto,
					)
					assert.NoError(t, err)

					retrievedPromptConfig, _ := db.GetQueries().
						RetrievePromptConfig(context.TODO(), promptConfig.ID)
					assert.Equal(t, updatedPromptConfig.Name, retrievedPromptConfig.Name)
					assert.Equal(t, updatedPromptConfig.ModelType, retrievedPromptConfig.ModelType)
					assert.Equal(
						t,
						updatedPromptConfig.ModelVendor,
						retrievedPromptConfig.ModelVendor,
					)
					assert.Equal(
						t,
						[]byte(*updatedPromptConfig.ModelParameters),
						retrievedPromptConfig.ModelParameters,
					)
					assert.Equal(
						t,
						[]byte(*updatedPromptConfig.ProviderPromptMessages),
						retrievedPromptConfig.ProviderPromptMessages,
					)
					assert.Equal(
						t,
						updatedPromptConfig.ExpectedTemplateVariables,
						retrievedPromptConfig.ExpectedTemplateVariables,
					)
				})
			}
		})

		t.Run("invalidates prompt-config caches", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			promptConfig, _ := factories.CreateOpenAIPromptConfig(context.TODO(), application.ID)

			applicationID := db.UUIDToString(&application.ID)

			cacheKeys := []string{
				fmt.Sprintf("%s:%s", applicationID, db.UUIDToString(&promptConfig.ID)),
				applicationID,
			}

			for _, cacheKey := range cacheKeys {
				redisDB.Set(context.TODO(), cacheKey, "test", 0)
				redisMock.ExpectDel(cacheKey).SetVal(1)
			}

			_, err := repositories.UpdatePromptConfig(
				context.TODO(),
				promptConfig.ID,
				dto.PromptConfigUpdateDTO{},
			)
			assert.NoError(t, err)

			time.Sleep(testutils.GetSleepTimeout())

			assert.NoError(t, redisMock.ExpectationsWereMet())
		})

		t.Run(
			"returns error if fails to retrieve the default prompt config",
			func(t *testing.T) {
				application, _ := factories.CreateApplication(context.TODO(), project.ID)
				promptConfig, _ := factories.CreateOpenAIPromptConfig(
					context.TODO(),
					application.ID,
				)

				_ = db.GetQueries().DeletePromptConfig(context.TODO(), promptConfig.ID)

				err := repositories.UpdateApplicationDefaultPromptConfig(
					context.TODO(),
					application.ID,
					promptConfig.ID,
				)
				assert.Error(t, err)
			},
		)

		t.Run("returns error if failed to retrieve prompt config", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			promptConfig, _ := factories.CreateOpenAIPromptConfig(context.TODO(), application.ID)

			_ = db.GetQueries().DeletePromptConfig(context.TODO(), promptConfig.ID)

			_, err := repositories.UpdatePromptConfig(
				context.TODO(),
				promptConfig.ID,
				dto.PromptConfigUpdateDTO{Name: &newName},
			)
			assert.Error(t, err)
		})

		t.Run("returns error if failed to parse prompt messages", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			promptConfig, _ := factories.CreateOpenAIPromptConfig(context.TODO(), application.ID)

			badMessage := ptr.To(json.RawMessage("invalid"))
			_, err := repositories.UpdatePromptConfig(
				context.TODO(),
				promptConfig.ID,
				dto.PromptConfigUpdateDTO{ProviderPromptMessages: badMessage},
			)
			assert.Error(t, err)
		})

		t.Run("returns error when prompt config name collides", func(t *testing.T) {
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			existingPromptConfig, _ := factories.CreateOpenAIPromptConfig(
				context.TODO(),
				application.ID,
			)

			promptConfig, _ := db.GetQueries().
				CreatePromptConfig(context.TODO(), models.CreatePromptConfigParams{
					ApplicationID:             application.ID,
					Name:                      factories.RandomString(10),
					ModelVendor:               models.ModelVendorOPENAI,
					ModelType:                 models.ModelTypeGpt4,
					ModelParameters:           existingPromptConfig.ModelParameters,
					ProviderPromptMessages:    existingPromptConfig.ProviderPromptMessages,
					ExpectedTemplateVariables: []string{""},
					IsDefault:                 false,
				})

			_, err := repositories.UpdatePromptConfig(
				context.TODO(),
				promptConfig.ID,
				dto.PromptConfigUpdateDTO{Name: &existingPromptConfig.Name},
			)
			assert.Error(t, err)
		})
	})

	t.Run("DeletePromptConfig", func(t *testing.T) {
		t.Run("deletes a prompt config", func(t *testing.T) {
			project, _ := factories.CreateProject(context.TODO())
			application, _ := factories.CreateApplication(context.TODO(), project.ID)
			promptConfig, _ := factories.CreateOpenAIPromptConfig(context.TODO(), application.ID)

			repositories.DeletePromptConfig(context.TODO(), application.ID, promptConfig.ID)

			_, err := db.GetQueries().RetrievePromptConfig(context.TODO(), promptConfig.ID)
			assert.Error(t, err)
		})
	})

	t.Run("Prompt Config Analytics", func(t *testing.T) {
		project, _ := factories.CreateProject(context.TODO())
		application, _ := factories.CreateApplication(context.TODO(), project.ID)
		promptConfig, _ := factories.CreateOpenAIPromptConfig(context.TODO(), application.ID)
		_, _ = factories.CreatePromptRequestRecord(context.TODO(), promptConfig.ID)

		fromDate := time.Now().AddDate(0, 0, -1)
		toDate := fromDate.AddDate(0, 0, 2)

		t.Run("GetPromptConfigAPIRequestCountByDateRange", func(t *testing.T) {
			t.Run("get total prompt requests by date range", func(t *testing.T) {
				totalRequests := repositories.GetPromptConfigAPIRequestCountByDateRange(
					context.TODO(),
					promptConfig.ID,
					fromDate,
					toDate,
				)
				assert.Equal(t, int64(1), totalRequests)
			})
		})

		t.Run("GetPromptConfigAnalyticsByDateRange", func(t *testing.T) {
			t.Run("get token usage for each model types by date range", func(t *testing.T) {
				promptConfigAnalytics := repositories.GetPromptConfigAnalyticsByDateRange(
					context.TODO(),
					promptConfig.ID,
					fromDate,
					toDate,
				)
				assert.Equal(t, int64(1), promptConfigAnalytics.TotalAPICalls)
				assert.Equal(
					t,
					decimal.RequireFromString("0.0000465").String(),
					promptConfigAnalytics.TokenCost.String(),
				)
			})
		})
	})
}
