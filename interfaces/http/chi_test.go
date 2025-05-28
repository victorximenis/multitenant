package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/victorximenis/multitenant/tenantcontext"
)

func TestChiTenantMiddleware(t *testing.T) {
	// Setup mock tenant service
	mockService := NewMockTenantService()

	// Setup Chi router
	r := chi.NewRouter()

	middleware := ChiTenantMiddleware(ChiMiddlewareConfig{
		TenantService: mockService,
		HeaderName:    "X-Tenant-Id",
	})

	r.Use(middleware)

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		tenant, ok := tenantcontext.GetTenant(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "tenant not in context"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tenant": "` + tenant.Name + `"}`))
	})

	t.Run("Valid tenant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-Id", "test-tenant")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), "test-tenant")
		assert.Equal(t, "test-tenant", w.Header().Get("X-Tenant-Name"))
	})

	t.Run("Missing header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), "tenant header X-Tenant-Id not provided")
	})

	t.Run("Tenant not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-Id", "unknown-tenant")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), "tenant not found")
	})

	t.Run("Inactive tenant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-Id", "inactive-tenant")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), "tenant is inactive")
	})
}

func TestChiTenantMiddleware_CustomHeaderName(t *testing.T) {
	mockService := NewMockTenantService()

	r := chi.NewRouter()

	middleware := ChiTenantMiddleware(ChiMiddlewareConfig{
		TenantService: mockService,
		HeaderName:    "Custom-Tenant-Header",
	})

	r.Use(middleware)

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		tenant, ok := tenantcontext.GetTenant(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "tenant not in context"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tenant": "` + tenant.Name + `"}`))
	})

	t.Run("Valid tenant with custom header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Custom-Tenant-Header", "test-tenant")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), "test-tenant")
	})

	t.Run("Missing custom header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), "tenant header Custom-Tenant-Header not provided")
	})
}

func TestChiTenantMiddleware_CustomErrorHandler(t *testing.T) {
	mockService := NewMockTenantService()

	r := chi.NewRouter()

	customErrorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(`{"custom_error": true, "message": "` + err.Error() + `"}`))
	}

	middleware := ChiTenantMiddleware(ChiMiddlewareConfig{
		TenantService: mockService,
		HeaderName:    "X-Tenant-Id",
		ErrorHandler:  customErrorHandler,
	})

	r.Use(middleware)

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success": true}`))
	})

	t.Run("Custom error handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTeapot, w.Code)
		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), "custom_error")
		assert.Contains(t, string(body), "tenant header X-Tenant-Id not provided")
	})
}
