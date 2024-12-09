package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"github.com/basemind-ai/monorepo/shared/go/exc"
	"github.com/basemind-ai/monorepo/shared/go/serialization"
	"io"
	"net/http"
	"time"
)

// Client is a wrapper around the http client that exposes semantic receivers.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
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

// Get makes a GET request.
func (client *Client) Get(
	ctx context.Context,
	path string,
) (*http.Response, error) {
	return client.Request(ctx, http.MethodGet, path, nil)
}

// Delete makes a DELETE request.
func (client *Client) Delete(
	ctx context.Context,
	path string,
) (*http.Response, error) {
	return client.Request(ctx, http.MethodDelete, path, nil)
}

// Post makes a POST request.
func (client *Client) Post(
	ctx context.Context,
	path string,
	body any,
) (*http.Response, error) {
	return client.Request(ctx, http.MethodPost, path, body)
}

// Put makes a PUT request.
func (client *Client) Put(
	ctx context.Context,
	path string,
	body any,
) (*http.Response, error) {
	return client.Request(ctx, http.MethodPut, path, body)
}

// Patch makes a PATCH request.
func (client *Client) Patch(
	ctx context.Context,
	path string,
	body any,
) (*http.Response, error) {
	return client.Request(ctx, http.MethodPatch, path, body)
}
