package calculator_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	httpcalc "order-pack-calculator/internal/api/http/controller/calculator"
	httpcalcmocks "order-pack-calculator/internal/api/http/controller/calculator/mocks"
	svcalc "order-pack-calculator/internal/usecase/calculator"
	"order-pack-calculator/internal/domain"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestRouter(uc *httpcalcmocks.MockUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	httpcalc.New(uc, discardLogger()).RegisterRoutes(r)
	return r
}

func TestController_Calculate_OK(t *testing.T) {
	m := httpcalcmocks.NewMockUseCase(t)
	m.EXPECT().Calculate(mock.Anything, 100).Return(&svcalc.Result{
		Packs: []svcalc.PackLine{{Pack: 50, Quantity: 2}},
	}, nil)

	r := newTestRouter(m)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"items":100}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body httpcalc.CalculateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "ok", body.Message)
	require.Len(t, body.Packs, 1)
	assert.Equal(t, 50, body.Packs[0].Pack)
	assert.Equal(t, 2, body.Packs[0].Quantity)
}

func TestController_Calculate_ErrNoExactPacking(t *testing.T) {
	m := httpcalcmocks.NewMockUseCase(t)
	m.EXPECT().Calculate(mock.Anything, 5).Return(nil, domain.ErrNoExactPacking)

	r := newTestRouter(m)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"items":5}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestController_ReplacePackSizes_OK(t *testing.T) {
	m := httpcalcmocks.NewMockUseCase(t)
	m.EXPECT().ReplacePackSizes(mock.Anything, []int{23, 31, 53}).Return(nil)
	m.EXPECT().ListPackSizes(mock.Anything).Return([]int{23, 31, 53}, nil)

	r := newTestRouter(m)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/pack-sizes", strings.NewReader(`{"sizes":[23,31,53]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body httpcalc.PackSizesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, []int{23, 31, 53}, body.Sizes)
}

func TestController_ReplacePackSizes_InvalidJSON(t *testing.T) {
	m := httpcalcmocks.NewMockUseCase(t)

	r := newTestRouter(m)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/pack-sizes", strings.NewReader(`{not json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var body httpcalc.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Contains(t, body.Error, "invalid request")
}

func TestController_ReplacePackSizes_EmptySizes(t *testing.T) {
	m := httpcalcmocks.NewMockUseCase(t)

	r := newTestRouter(m)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/pack-sizes", strings.NewReader(`{"sizes":[]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var body httpcalc.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "sizes must not be empty", body.Error)
}

func TestController_Calculate_InvalidJSON(t *testing.T) {
	m := httpcalcmocks.NewMockUseCase(t)

	r := newTestRouter(m)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"items":}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var body httpcalc.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Contains(t, body.Error, "invalid request")
}

func TestController_Calculate_ItemsNotPositive(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"zero", `{"items":0}`},
		{"negative", `{"items":-3}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := httpcalcmocks.NewMockUseCase(t)

			r := newTestRouter(m)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(tt.raw))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusBadRequest, w.Code)
			var body httpcalc.ErrorResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
			assert.Equal(t, "items must be positive", body.Error)
		})
	}
}
