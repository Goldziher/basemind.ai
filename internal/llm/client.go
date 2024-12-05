package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"github.com/basemind-ai/gateway/internal/utils/exc"
	"github.com/basemind-ai/gateway/internal/utils/serialization"
	"io"
	"net/http"
	"time"
)

type ChatCompletionStreamChoiceDelta struct {
	Content   string     `json:"content,omitempty"`
	Role      string     `json:"role,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ChatCompletionStreamChoice struct {
	Index                int                             `json:"index"`
	Delta                ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason         FinishReason                    `json:"finish_reason"`
	ContentFilterResults ContentFilterResults            `json:"content_filter_results,omitempty"`
}

type PromptFilterResult struct {
	Index                int                  `json:"index"`
	ContentFilterResults ContentFilterResults `json:"content_filter_results,omitempty"`
}

type ChatCompletionStreamResponse struct {
	ID                  string                       `json:"id"`
	Object              string                       `json:"object"`
	Created             int64                        `json:"created"`
	Model               string                       `json:"model"`
	Choices             []ChatCompletionStreamChoice `json:"choices"`
	SystemFingerprint   string                       `json:"system_fingerprint"`
	PromptAnnotations   []PromptAnnotation           `json:"prompt_annotations,omitempty"`
	PromptFilterResults []PromptFilterResult         `json:"prompt_filter_results,omitempty"`
	// An optional field that will only be present when you set stream_options: {"include_usage": true} in your request.
	// When present, it contains a null value except for the last chunk which contains the token usage statistics
	// for the entire request.
	Usage *Usage `json:"usage,omitempty"`
}

// ChatCompletionStream
// Note: Perhaps it is more elegant to abstract Stream using generics.
type ChatCompletionStream struct {
	*streamReader[ChatCompletionStreamResponse]
}

// Client is a wrapper around the http client that exposes semantic receivers.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// New returns a new http client instance.
func New(baseURL string, httpClient *http.Client) Client {
	if httpClient != nil {
		return Client{BaseURL: baseURL, HTTPClient: httpClient}
	}
	return Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: time.Duration(1) * time.Second},
	}
}

// Request is a generic method for making http requests.
func (client *Client) Request(
	ctx context.Context,
	method string,
	path string,
	body any,
) (*http.Response, error) {
	var requestBody io.Reader
	if body != nil {
		data := serialization.SerializeJSON(body)
		requestBody = bytes.NewBuffer(data)
	}

	url := fmt.Sprintf("%s%s", client.BaseURL, path)
	request := exc.MustResult(http.NewRequestWithContext(ctx, method, url, requestBody))
	request.Header.Add("Accept", `application/json`)

	return client.HTTPClient.Do(request)
}

// Post makes a POST request.
func (client *Client) Post(
	ctx context.Context,
	path string,
	body any,
) (*http.Response, error) {
	return client.Request(ctx, http.MethodPost, path, body)
}
