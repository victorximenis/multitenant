package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/victorximenis/multitenant/tenantcontext"
)

func TestGinTenantMiddleware(t *testing.T) {
	// Setup mock tenant service
	mockService := NewMockTenantService()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := TenantMiddleware(GinMiddlewareConfig{
		TenantService: mockService,
		HeaderName:    "X-Tenant-Id",
	})

	router.Use(middleware)

	router.GET("/test", func(c *gin.Context) {
		tenant, ok := tenantcontext.GetTenant(c.Request.Context())
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant not in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tenant": tenant.Name})
	})

	t.Run("Valid tenant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-Id", "test-tenant")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "test-tenant")
		assert.Equal(t, "test-tenant", w.Header().Get("X-Tenant-Name"))
	})

	t.Run("Missing header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "tenant header X-Tenant-Id not provided")
	})

	t.Run("Tenant not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-Id", "unknown-tenant")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "tenant not found")
	})

	t.Run("Inactive tenant", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-Id", "inactive-tenant")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "tenant is inactive")
	})
}

func TestGinTenantMiddleware_CustomHeaderName(t *testing.T) {
	mockService := NewMockTenantService()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := TenantMiddleware(GinMiddlewareConfig{
		TenantService: mockService,
		HeaderName:    "Custom-Tenant-Header",
	})

	router.Use(middleware)

	router.GET("/test", func(c *gin.Context) {
		tenant, ok := tenantcontext.GetTenant(c.Request.Context())
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant not in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tenant": tenant.Name})
	})

	t.Run("Valid tenant with custom header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Custom-Tenant-Header", "test-tenant")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "test-tenant")
	})

	t.Run("Missing custom header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "tenant header Custom-Tenant-Header not provided")
	})
}

func TestGinTenantMiddleware_CustomErrorHandler(t *testing.T) {
	mockService := NewMockTenantService()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	customErrorHandler := func(c *gin.Context, err error) {
		c.JSON(http.StatusTeapot, gin.H{
			"custom_error": true,
			"message":      err.Error(),
		})
		c.Abort()
	}

	middleware := TenantMiddleware(GinMiddlewareConfig{
		TenantService: mockService,
		HeaderName:    "X-Tenant-Id",
		ErrorHandler:  customErrorHandler,
	})

	router.Use(middleware)

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	t.Run("Custom error handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTeapot, w.Code)
		assert.Contains(t, w.Body.String(), "custom_error")
		assert.Contains(t, w.Body.String(), "tenant header X-Tenant-Id not provided")
	})
}
