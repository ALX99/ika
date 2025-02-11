package ika

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

type mockError struct {
	statusCode int
	_typeURI   string
	_title     string
	_detail    string
}

func (e *mockError) Error() string {
	return e._detail
}

func (e *mockError) Status() int {
	return e.statusCode
}

func (e *mockError) TypeURI() string {
	return e._typeURI
}

func (e *mockError) Title() string {
	return e._title
}

func (e *mockError) Detail() string {
	return e._detail
}

func Test_defualtErrorHandler(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		acceptHeader   string
		err            error
		expectedStatus int
		expectedBody   string
		expectedType   string
		expectedTitle  string
		expectedDetail string
	}{
		{
			name:           "override detail",
			acceptHeader:   "text/plain",
			err:            &mockError{_detail: "internal error"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal error",
		},
		{
			name:           "default internal server error",
			acceptHeader:   "text/plain",
			err:            &mockError{},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error",
		},
		{
			name:           "override code",
			acceptHeader:   "text/plain",
			err:            &mockError{statusCode: 666},
			expectedStatus: 666,
			expectedBody:   "An error occurred while processing the request",
		},
		{
			name:         "set everything and accept json",
			acceptHeader: "application/json;q=1.0",
			err: &mockError{
				_typeURI:   "type",
				_title:     "title",
				_detail:    "detail",
				statusCode: 666,
			},
			expectedStatus: 666,
			expectedBody:   `{"type":"type","title":"title","detail":"detail","status":666}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", tt.acceptHeader)
			rr := httptest.NewRecorder()

			defaultErrorHandler(rr, req, tt.err)

			res := rr.Result()
			defer res.Body.Close()

			is.Equal(res.StatusCode, tt.expectedStatus)

			body := new(bytes.Buffer)
			_, err := body.ReadFrom(res.Body)
			is.NoErr(err)

			is.Equal(body.String(), tt.expectedBody)
		})
	}
}
