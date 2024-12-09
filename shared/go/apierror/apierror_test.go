package apierror_test

import (
	"context"
	"github.com/basemind-ai/monorepo/shared/go/apierror"
	"github.com/basemind-ai/monorepo/shared/go/serialization"
	"github.com/basemind-ai/monorepo/shared/go/testutils"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError(t *testing.T) {
	t.Parallel()

	t.Run("APIError", func(t *testing.T) {
		t.Run("Render renders as expected", func(t *testing.T) {
			client := testutils.CreateTestHTTPClient(
				t,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					apierror.New(http.StatusNotFound, "err").Render(w)
				}),
			)
			res, err := client.Get(context.TODO(), "/")
			assert.Nil(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			body := apierror.APIError{}
			_ = serialization.DeserializeJSON[*apierror.APIError](res.Body, &body)
			assert.Equal(t, "err", body.Message)
			assert.Equal(t, http.StatusText(http.StatusNotFound), body.StatusText)
			assert.Equal(t, http.StatusNotFound, body.StatusCode)
		})

		t.Run("render with extra renders as expected", func(t *testing.T) {
			client := testutils.CreateTestHTTPClient(
				t,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					apierror.New(http.StatusNotFound, "err").RenderWithExtra(
						context.WithValue(r.Context(),
							apierror.ExtraContextKey,
							"extra",
						),
						w,
					)
				}),
			)

			res, err := client.Get(context.TODO(), "/")
			assert.Nil(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			body := apierror.APIError{}
			_ = serialization.DeserializeJSON[*apierror.APIError](res.Body, &body)
			assert.Equal(t, "err", body.Message)
			assert.Equal(t, http.StatusText(http.StatusNotFound), body.StatusText)
			assert.Equal(t, http.StatusNotFound, body.StatusCode)
			assert.Equal(t, "extra", body.Extra)
		})

		t.Run("Error returns the error message", func(t *testing.T) {
			assert.Equal(
				t,
				"status: 404, message: err",
				apierror.New(http.StatusNotFound, "err").Error(),
			)
		})
	})

	for _, testCase := range []struct {
		APIErrType      *apierror.APIError
		ExpectedMessage string
		ExpectedStatus  int
	}{
		{
			APIErrType:      apierror.NotFound("err"),
			ExpectedStatus:  http.StatusNotFound,
			ExpectedMessage: "err",
		},
		{
			APIErrType:      apierror.BadRequest("err"),
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedMessage: "err",
		},
		{
			APIErrType:      apierror.BadRequest(),
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedMessage: http.StatusText(http.StatusBadRequest),
		},
		{
			APIErrType:      apierror.Unauthorized("err"),
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedMessage: "err",
		},
		{
			APIErrType:      apierror.Forbidden("err"),
			ExpectedStatus:  http.StatusForbidden,
			ExpectedMessage: "err",
		},
		{
			APIErrType:      apierror.Unauthorized(),
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedMessage: http.StatusText(http.StatusUnauthorized),
		},
		{
			APIErrType:      apierror.UnprocessableContent("err"),
			ExpectedStatus:  http.StatusUnprocessableEntity,
			ExpectedMessage: "err",
		},
		{
			APIErrType:      apierror.UnprocessableContent(),
			ExpectedStatus:  http.StatusUnprocessableEntity,
			ExpectedMessage: http.StatusText(http.StatusUnprocessableEntity),
		},
		{
			APIErrType:      apierror.InternalServerError("err"),
			ExpectedStatus:  http.StatusInternalServerError,
			ExpectedMessage: "err",
		},
		{
			APIErrType:      apierror.InternalServerError(),
			ExpectedStatus:  http.StatusInternalServerError,
			ExpectedMessage: http.StatusText(http.StatusInternalServerError),
		},
	} {
		t.Run(http.StatusText(testCase.ExpectedStatus), func(t *testing.T) {
			client := testutils.CreateTestHTTPClient(
				t,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					testCase.APIErrType.Render(w)
				}),
			)
			res, err := client.Get(context.TODO(), "/")
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedStatus, res.StatusCode)

			var body apierror.APIError
			_ = serialization.DeserializeJSON(res.Body, &body)
			assert.Equal(t, testCase.ExpectedMessage, body.Message)
		},
		)
	}
}
