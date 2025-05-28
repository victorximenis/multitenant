package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/victorximenis/multitenant/core"
	"github.com/victorximenis/multitenant/tenantcontext"
)

// FiberMiddlewareConfig holds configuration for the Fiber tenant middleware
type FiberMiddlewareConfig struct {
	TenantService core.TenantService
	HeaderName    string
	ErrorHandler  func(*fiber.Ctx, error) error
}

// DefaultFiberErrorHandler provides default error handling for Fiber middleware
func DefaultFiberErrorHandler(c *fiber.Ctx, err error) error {
	statusCode := fiber.StatusInternalServerError

	if _, ok := err.(core.TenantNotFoundError); ok {
		statusCode = fiber.StatusNotFound
	} else if _, ok := err.(core.TenantInactiveError); ok {
		statusCode = fiber.StatusForbidden
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"message": err.Error(),
		"errors":  []fiber.Map{{"field": "tenant", "message": err.Error()}},
	})
}

// FiberTenantMiddleware creates a Fiber middleware for tenant resolution
func FiberTenantMiddleware(config FiberMiddlewareConfig) fiber.Handler {
	if config.HeaderName == "" {
		config.HeaderName = "X-Tenant-Id"
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultFiberErrorHandler
	}

	return func(c *fiber.Ctx) error {
		tenantName := c.Get(config.HeaderName)
		if tenantName == "" {
			return config.ErrorHandler(c, fmt.Errorf("tenant header %s not provided", config.HeaderName))
		}

		tenant, err := config.TenantService.GetTenant(c.UserContext(), tenantName)
		if err != nil {
			return config.ErrorHandler(c, err)
		}

		// Store tenant in context
		ctx := tenantcontext.WithTenant(c.UserContext(), tenant)
		c.SetUserContext(ctx)

		// Add tenant to response headers for debugging
		c.Set("X-Tenant-Name", tenant.Name)

		// Propagate tenant to tracing span if available
		tenantcontext.PropagateToSpan(ctx)

		return c.Next()
	}
}
