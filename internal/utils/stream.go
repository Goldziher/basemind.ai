package utils

import (
	"errors"
	"github.com/basemind-ai/monorepo/internal/utils/ptr"
	"github.com/basemind-ai/monorepo/services/api-gateway/internal/dto"
	"github.com/basemind-ai/monorepo/shared/go/db/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"io"
	"time"
)

// Stream is an interface for a gRPC stream.
type Stream[T any] interface {
	Recv() (*T, error)
}

// StreamMessage is a generic data type used to encapsulate the result of a streaming response from a connector.
type StreamMessage struct {
	Content            *string
	FinishReason       *string
	RequestTokenCount  *uint32
	ResponseTokenCount *uint32
}

type StreamFinishResult struct {
	FinishReason       models.PromptFinishReason
	RequestTokenCount  uint32
	ResponseTokenCount uint32
}

// StreamFromClient is a generic function that handles the streaming response from an LLM API.
func StreamFromClient[T any]( //nolint: revive
	channel chan<- dto.PromptResultDTO,
	finalResult *dto.PromptResultDTO,
	recordParams *models.CreatePromptRequestRecordParams,
	startTime time.Time,
	stream Stream[T],
	parseMessage func(*T) *StreamMessage,
) *StreamFinishResult {
	var streamResult *StreamFinishResult

	for {
		msg, receiveErr := stream.Recv()

		isFinished := false

		parsedMessage := parseMessage(msg)
		if parsedMessage.FinishReason != nil {
			streamResult = &StreamFinishResult{
				FinishReason:       models.PromptFinishReason(*parsedMessage.FinishReason),
				RequestTokenCount:  ptr.Deref(parsedMessage.RequestTokenCount, 0),
				ResponseTokenCount: ptr.Deref(parsedMessage.ResponseTokenCount, 0),
			}

			isFinished = true
		}

		if receiveErr != nil {
			if !errors.Is(receiveErr, io.EOF) {
				log.Debug().Err(receiveErr).Msg("received stream error")
				finalResult.Error = receiveErr
			}

			isFinished = true
		}

		if recordParams.DurationMs.Int32 == 0 {
			duration := int32(time.Until(startTime).Milliseconds())
			recordParams.DurationMs = pgtype.Int4{Int32: duration, Valid: true}
		}

		if isFinished {
			break
		}

		channel <- dto.PromptResultDTO{
			Content: parsedMessage.Content,
		}
	}

	recordParams.FinishTime = pgtype.Timestamptz{Time: time.Now(), Valid: true}

	return streamResult
}
