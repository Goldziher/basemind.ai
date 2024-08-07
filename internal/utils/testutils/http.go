package testutils

import (
	"github.com/basemind-ai/monorepo/internal/utils/httpclient"
	"net/http"
	"net/http/httptest"
	"testing"
)

func CreateTestHTTPClient(t *testing.T, handler http.Handler) httpclient.Client {
	t.Helper()
	server := httptest.NewServer(handler)

	t.Cleanup(func() {
		server.Close()
	})

	client := server.Client()
	// we want to test that redirects are returned correctly by the API
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return httpclient.New(server.URL, client)
}
