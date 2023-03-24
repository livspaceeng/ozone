package unit_tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livspaceeng/ozone/internal/server"
	"github.com/livspaceeng/ozone/configs"
	"github.com/stretchr/testify/assert"
)

func TestHealthController_Status(t *testing.T) {
	configs.Init()
	r := server.NewRouter()

	req, _ := http.NewRequest("GET", "/health", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, "OK!", w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
}