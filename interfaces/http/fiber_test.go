package http

import (
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/victorximenis/multitenant/tenantcontext"
)

func TestFiberTenantMiddleware(t *testing.T) {
	// Setup mock tenant service
	mockService := NewMockTenantService()

	// Setup Fiber app
	app := fiber.New()

	middleware := FiberTenantMiddleware(FiberMiddlewareConfig{
		TenantService: mockService,
		HeaderName:    "X-Tenant-Id",
	})

	app.Use(middleware)

	app.Get("/test", func(c *fiber.Ctx) error {
		tenant, ok := tenantcontext.GetTenant(c.UserContext())
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "tenant not in context"})
		}
		return c.JSON(fiber.Map{"tenant": tenant.Name})
	})

	t.Run("Valid tenant", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-Id", "test-tenant")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "test-tenant")
		assert.Equal(t, "test-tenant", resp.Header.Get("X-Tenant-Name"))
	})

	t.Run("Missing header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "tenant header X-Tenant-Id not provided")
	})

	t.Run("Tenant not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-Id", "unknown-tenant")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "tenant not found")
	})

	t.Run("Inactive tenant", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Tenant-Id", "inactive-tenant")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "tenant is inactive")
	})
}

func TestFiberTenantMiddleware_CustomHeaderName(t *testing.T) {
	mockService := NewMockTenantService()

	app := fiber.New()

	middleware := FiberTenantMiddleware(FiberMiddlewareConfig{
		TenantService: mockService,
		HeaderName:    "Custom-Tenant-Header",
	})

	app.Use(middleware)

	app.Get("/test", func(c *fiber.Ctx) error {
		tenant, ok := tenantcontext.GetTenant(c.UserContext())
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "tenant not in context"})
		}
		return c.JSON(fiber.Map{"tenant": tenant.Name})
	})

	t.Run("Valid tenant with custom header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Custom-Tenant-Header", "test-tenant")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "test-tenant")
	})

	t.Run("Missing custom header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "tenant header Custom-Tenant-Header not provided")
	})
}

func TestFiberTenantMiddleware_CustomErrorHandler(t *testing.T) {
	mockService := NewMockTenantService()

	app := fiber.New()

	customErrorHandler := func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusTeapot).JSON(fiber.Map{
			"custom_error": true,
			"message":      err.Error(),
		})
	}

	middleware := FiberTenantMiddleware(FiberMiddlewareConfig{
		TenantService: mockService,
		HeaderName:    "X-Tenant-Id",
		ErrorHandler:  customErrorHandler,
	})

	app.Use(middleware)

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	t.Run("Custom error handler", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusTeapot, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "custom_error")
		assert.Contains(t, string(body), "tenant header X-Tenant-Id not provided")
	})
}
