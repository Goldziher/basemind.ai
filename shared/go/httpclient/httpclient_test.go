package httpclient_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/basemind-ai/monorepo/shared/go/exc"
	"github.com/basemind-ai/monorepo/shared/go/httpclient"
	"github.com/basemind-ai/monorepo/shared/go/testutils"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Body struct {
	Message string
}

func TestClient(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("returns a new client", func(t *testing.T) {
			client := httpclient.New("http://localhost:8080", nil)
			assert.NotNil(t, client)
			assert.Equal(t, client.BaseURL, "http://localhost:8080")
		})
		t.Run("allows passing in an httpClient instance", func(t *testing.T) {
			httpClient := &http.Client{}
			client := httpclient.New("http://localhost:8080", httpClient)
			assert.Equal(t, client.HTTPClient, httpClient)
		})
	})

	t.Run("Request", func(t *testing.T) {
		t.Run("makes a request", func(t *testing.T) {
			client := testutils.CreateTestHTTPClient(
				t,
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					assert.Equal(t, request.Method, http.MethodGet)
					writer.WriteHeader(http.StatusOK)
				}),
			)
			res, err := client.Request(context.TODO(), http.MethodGet, "/", nil)
			assert.NoError(t, err)
			assert.Equal(t, res.StatusCode, http.StatusOK)
		})

		t.Run("makes a request with a body", func(t *testing.T) {
			client := testutils.CreateTestHTTPClient(
				t,
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					assert.Equal(t, request.Method, http.MethodPost)
					defer func() {
						exc.LogIfErr(request.Body.Close(), "error closing request body")
					}()
					data, _ := io.ReadAll(request.Body)
					body := Body{}
					_ = json.Unmarshal(data, &body)
					assert.Equal(t, body.Message, "ABC")
					writer.WriteHeader(http.StatusOK)
				}),
			)
			res, err := client.Request(context.TODO(), http.MethodPost, "/", &Body{
				Message: "ABC",
			})
			assert.NoError(t, err)
			assert.Equal(t, res.StatusCode, http.StatusOK)
		})

		t.Run("returns an error if the request times out", func(t *testing.T) {
			client := testutils.CreateTestHTTPClient(
				t,
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				}),
			)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			res, err := client.Request(ctx, http.MethodGet, "/", nil)
			assert.Error(t, err)
			assert.Nil(t, res)
		})
	})

	for _, testCase := range []struct {
		Body   *Body
		Method string
	}{
		{
			Method: "GET",
			Body:   nil,
		},
		{
			Method: "DELETE",
			Body:   nil,
		},
		{
			Method: "POST",
			Body:   &Body{Message: "ABC"},
		},
		{
			Method: "PUT",
			Body:   &Body{Message: "ABC"},
		},
		{
			Method: "PATCH",
			Body:   &Body{Message: "ABC"},
		},
	} {
		t.Run(fmt.Sprintf("TestClient: %s", testCase.Method), func(t *testing.T) {
			client := testutils.CreateTestHTTPClient(
				t,
				http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					assert.Equal(t, request.Method, testCase.Method)

					defer func() {
						exc.LogIfErr(request.Body.Close(), "error closing request body")
					}()

					if testCase.Body != nil {
						data, _ := io.ReadAll(request.Body)
						body := Body{}
						_ = json.Unmarshal(data, &body)
						assert.Equal(t, body.Message, testCase.Body.Message)
					}
					writer.WriteHeader(http.StatusOK)
				}),
			)
			var (
				res *http.Response
				err error
			)
			switch testCase.Method {
			case http.MethodGet:
				res, err = client.Get(context.TODO(), "/")
			case http.MethodDelete:
				res, err = client.Delete(context.TODO(), "/")
			case http.MethodPost:
				res, err = client.Post(context.TODO(), "/", testCase.Body)
			case http.MethodPut:
				res, err = client.Put(context.TODO(), "/", testCase.Body)
			case http.MethodPatch:
				res, err = client.Patch(context.TODO(), "/", testCase.Body)
			}
			assert.NoError(t, err)
			assert.Equal(t, res.StatusCode, http.StatusOK)
		})
	}
}
