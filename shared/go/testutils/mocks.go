package testutils

import (
	"context"
	cohereconnector "github.com/basemind-ai/monorepo/gen/go/cohere/v1"
	openaiconnector "github.com/basemind-ai/monorepo/gen/go/openai/v1"
	"github.com/basemind-ai/monorepo/gen/go/ptesting/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

// OpenAI Service

type MockOpenAIService struct {
	openaiconnector.UnimplementedOpenAIServiceServer
	Context         context.Context
	Error           error
	ExpectedRequest *openaiconnector.OpenAIPromptRequest
	Response        *openaiconnector.OpenAIPromptResponse
	T               *testing.T
	Stream          []*openaiconnector.OpenAIStreamResponse
}

func (m MockOpenAIService) OpenAIPrompt(
	_ context.Context,
	request *openaiconnector.OpenAIPromptRequest,
) (*openaiconnector.OpenAIPromptResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	assert.NotNil(m.T, m.Response)

	if m.ExpectedRequest != nil {
		assert.Equal(m.T, m.ExpectedRequest.Model, request.Model)
		assert.Equal(m.T, m.ExpectedRequest.ApplicationId, request.ApplicationId)
		assert.Len(m.T, m.ExpectedRequest.Messages, len(request.Messages))

		assert.Equal(m.T, len(m.ExpectedRequest.Messages), len(request.Messages))
		for i, message := range m.ExpectedRequest.Messages {
			assert.Equal(m.T, message.Role, request.Messages[i].Role)
			assert.Equal(m.T, message.Content, request.Messages[i].Content)
		}

		assert.Equal(m.T, m.ExpectedRequest.Parameters, request.Parameters)
	}

	return m.Response, nil
}

func (m MockOpenAIService) OpenAIStream(
	request *openaiconnector.OpenAIPromptRequest,
	stream openaiconnector.OpenAIService_OpenAIStreamServer,
) error {
	if m.Error != nil {
		return m.Error
	}

	assert.NotNil(m.T, m.Stream)

	if m.ExpectedRequest != nil {
		assert.Equal(m.T, m.ExpectedRequest.ApplicationId, request.ApplicationId)
		assert.Equal(m.T, m.ExpectedRequest.Model, request.Model)

		assert.Equal(m.T, len(m.ExpectedRequest.Messages), len(request.Messages))
		for i, message := range m.ExpectedRequest.Messages {
			assert.Equal(m.T, message.Content, request.Messages[i].Content)
			assert.Equal(m.T, message.Role, request.Messages[i].Role)
		}

		assert.Equal(m.T, m.ExpectedRequest.Parameters, request.Parameters)
	}

	for _, response := range m.Stream {
		_ = stream.Send(response)
	}

	return nil
}

// Prompt Testing Service

type MockPromptTestingService struct {
	ptesting.UnimplementedPromptTestingServiceServer
	Context         context.Context
	Error           error
	ExpectedRequest *ptesting.PromptTestRequest
	T               *testing.T
	Stream          []*ptesting.PromptTestingStreamingPromptResponse
}

func (m MockPromptTestingService) TestPrompt(
	request *ptesting.PromptTestRequest,
	stream ptesting.PromptTestingService_TestPromptServer,
) error {
	if m.Error != nil {
		return m.Error
	}

	assert.NotNil(m.T, m.Stream)

	if m.ExpectedRequest != nil {
		assert.Equal(m.T, m.ExpectedRequest.ApplicationId, request.ApplicationId)
		assert.Equal(m.T, m.ExpectedRequest.ModelParameters, request.ModelParameters)
		assert.Equal(m.T, m.ExpectedRequest.ModelVendor, request.ModelVendor)
		assert.Equal(m.T, m.ExpectedRequest.PromptConfigId, request.PromptConfigId)
		assert.Equal(m.T, m.ExpectedRequest.ProviderPromptMessages, request.ProviderPromptMessages)
		assert.Equal(m.T, m.ExpectedRequest.TemplateVariables, request.TemplateVariables)
	}

	for _, response := range m.Stream {
		_ = stream.Send(response)
	}

	return nil
}

// Cohere Service

type MockCohereService struct {
	cohereconnector.UnimplementedCohereServiceServer
	Context         context.Context
	Error           error
	ExpectedRequest *cohereconnector.CoherePromptRequest
	Response        *cohereconnector.CoherePromptResponse
	T               *testing.T
	Stream          []*cohereconnector.CohereStreamResponse
}

func (m MockCohereService) CoherePrompt(
	_ context.Context,
	request *cohereconnector.CoherePromptRequest,
) (*cohereconnector.CoherePromptResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	assert.NotNil(m.T, m.Response)

	if m.ExpectedRequest != nil {
		assert.Equal(m.T, m.ExpectedRequest.Model, request.Model)
		assert.Equal(m.T, m.ExpectedRequest.Message, request.Message)
		assert.Equal(m.T, m.ExpectedRequest.Parameters, request.Parameters)
	}

	return m.Response, nil
}

func (m MockCohereService) CohereStream(
	request *cohereconnector.CoherePromptRequest,
	stream cohereconnector.CohereService_CohereStreamServer,
) error {
	if m.Error != nil {
		return m.Error
	}

	assert.NotNil(m.T, m.Stream)

	if m.ExpectedRequest != nil {
		assert.Equal(m.T, m.ExpectedRequest.Model, request.Model)
		assert.Equal(m.T, m.ExpectedRequest.Message, request.Message)
		assert.Equal(m.T, m.ExpectedRequest.Parameters, request.Parameters)
	}

	for _, response := range m.Stream {
		_ = stream.Send(response)
	}

	return nil
}
