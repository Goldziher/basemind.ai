package serialization_test

import (
	"github.com/basemind-ai/gateway/internal/utils/serialization"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorReader struct {
	io.ReadCloser
}

func (*errorReader) Read([]byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (*errorReader) Close() error {
	return nil
}

func TestSerializationUtils(t *testing.T) {
	t.Run("ReadBody", func(t *testing.T) {
		t.Run("it reads the response body", func(t *testing.T) {
			body := io.NopCloser(strings.NewReader("Hello World"))
			result, err := serialization.ReadBody(body)
			assert.Nil(t, err)
			assert.Equal(t, result, []byte("Hello World"))
		})

		t.Run("it returns an error if the body cannot be read", func(t *testing.T) {
			body := &errorReader{}
			_, err := serialization.ReadBody(body)
			assert.NotNil(t, err)
		})
	})

	t.Run("DeserializedJSON", func(t *testing.T) {
		t.Run("it deserializes the body to the target", func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(`{"message":"Hello World"}`))

			target := struct {
				Message string
			}{}
			err := serialization.DeserializeJSON(body, &target)
			assert.Nil(t, err)
			assert.Equal(t, target.Message, "Hello World")
		})

		t.Run("it returns an error if the body cannot be read", func(t *testing.T) {
			body := io.NopCloser(&errorReader{})
			err := serialization.DeserializeJSON(body, map[string]string{})
			assert.NotNil(t, err)
		})

		t.Run("it returns an error if the body cannot be deserialized", func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(`invalid"`))
			err := serialization.DeserializeJSON(body, map[string]string{})
			assert.NotNil(t, err)
		})
	})

	t.Run("SerializeJSON", func(t *testing.T) {
		t.Run("it serializes the target to json", func(t *testing.T) {
			target := struct {
				Message string
			}{
				Message: "Hello World",
			}
			result := serialization.SerializeJSON(target)
			assert.Equal(t, result, []byte(`{"Message":"Hello World"}`))
		})

		t.Run("it panics if the target cannot be serialized", func(t *testing.T) {
			assert.Panics(t, func() {
				serialization.SerializeJSON(func() {})
			})
		})
	})

	t.Run("RenderJSONResponse", func(t *testing.T) {
		t.Run("renders response with expected status", func(t *testing.T) {
			statusCode := http.StatusOK
			body := struct {
				Message string
			}{
				Message: "Hello World",
			}
			w := httptest.NewRecorder()

			serialization.RenderJSONResponse(w, statusCode, body)

			assert.Equal(t, w.Code, statusCode)
			assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
			assert.Equal(t, w.Body.String(), `{"Message":"Hello World"}`+"\n")
		})
	})
}
