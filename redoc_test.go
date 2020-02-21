package micro

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedoc(t *testing.T) {
	var should = require.New(t)

	req := httptest.NewRequest("GET", "/docs", nil)
	recorder := httptest.NewRecorder()
	redoc := &RedocOpts{}

	redoc.Serve(recorder, req, map[string]string{})

	should.Equal(200, recorder.Code)
	should.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	should.Contains(recorder.Body.String(), "<title>API documentation</title>")
}

func TestRedoc2(t *testing.T) {
	var should = require.New(t)

	req := httptest.NewRequest("GET", "/docs", nil)
	recorder := httptest.NewRecorder()

	redoc := &RedocOpts{}
	redoc.AddSpec("PetStore", "https://rebilly.github.io/ReDoc/swagger.yaml")
	redoc.AddSpec("Instagram", "https://api.apis.guru/v2/specs/instagram.com/1.0.0/swagger.yaml")
	redoc.AddSpec("Google Calendar", "https://api.apis.guru/v2/specs/googleapis.com/calendar/v3/swagger.yaml")

	redoc.Serve(recorder, req, map[string]string{})

	should.Equal(200, recorder.Code)
	should.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	should.Contains(recorder.Body.String(), "<title>API documentation</title>")
	should.Contains(recorder.Body.String(), "Google Calendar")
}
