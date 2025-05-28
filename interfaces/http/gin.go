package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/victorximenis/multitenant/core"
	"github.com/victorximenis/multitenant/tenantcontext"
)

// GinMiddlewareConfig holds configuration for the Gin tenant middleware
type GinMiddlewareConfig struct {
	TenantService core.TenantService
	HeaderName    string
	ErrorHandler  func(*gin.Context, error)
}

// DefaultGinErrorHandler provides default error handling for Gin middleware
func DefaultGinErrorHandler(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError

	if _, ok := err.(core.TenantNotFoundError); ok {
		statusCode = http.StatusNotFound
	} else if _, ok := err.(core.TenantInactiveError); ok {
		statusCode = http.StatusForbidden
	}

	c.JSON(statusCode, gin.H{
		"success": false,
		"message": err.Error(),
		"errors":  []gin.H{{"field": "tenant", "message": err.Error()}},
	})
	c.Abort()
}

// TenantMiddleware creates a Gin middleware for tenant resolution
func TenantMiddleware(config GinMiddlewareConfig) gin.HandlerFunc {
	if config.HeaderName == "" {
		config.HeaderName = "X-Tenant-Id"
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultGinErrorHandler
	}

	return func(c *gin.Context) {
		tenantName := c.GetHeader(config.HeaderName)
		if tenantName == "" {
			config.ErrorHandler(c, fmt.Errorf("tenant header %s not provided", config.HeaderName))
			return
		}

		tenant, err := config.TenantService.GetTenant(c.Request.Context(), tenantName)
		if err != nil {
			config.ErrorHandler(c, err)
			return
		}

		// Store tenant in context
		ctx := tenantcontext.WithTenant(c.Request.Context(), tenant)
		c.Request = c.Request.WithContext(ctx)

		// Add tenant to response headers for debugging
		c.Header("X-Tenant-Name", tenant.Name)

		// Propagate tenant to tracing span if available
		tenantcontext.PropagateToSpan(ctx)

		c.Next()
	}
}
