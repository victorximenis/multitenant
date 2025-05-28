package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/victorximenis/multitenant/core"
	"github.com/victorximenis/multitenant/tenantcontext"
)

// ChiMiddlewareConfig holds configuration for the Chi tenant middleware
type ChiMiddlewareConfig struct {
	TenantService core.TenantService
	HeaderName    string
	ErrorHandler  func(http.ResponseWriter, *http.Request, error)
}

// DefaultChiErrorHandler provides default error handling for Chi middleware
func DefaultChiErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusInternalServerError

	if _, ok := err.(core.TenantNotFoundError); ok {
		statusCode = http.StatusNotFound
	} else if _, ok := err.(core.TenantInactiveError); ok {
		statusCode = http.StatusForbidden
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success": false,
		"message": err.Error(),
		"errors":  []map[string]string{{"field": "tenant", "message": err.Error()}},
	}

	json.NewEncoder(w).Encode(response)
}

// ChiTenantMiddleware creates a Chi middleware for tenant resolution
func ChiTenantMiddleware(config ChiMiddlewareConfig) func(http.Handler) http.Handler {
	if config.HeaderName == "" {
		config.HeaderName = "X-Tenant-Id"
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultChiErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantName := r.Header.Get(config.HeaderName)
			if tenantName == "" {
				config.ErrorHandler(w, r, fmt.Errorf("tenant header %s not provided", config.HeaderName))
				return
			}

			tenant, err := config.TenantService.GetTenant(r.Context(), tenantName)
			if err != nil {
				config.ErrorHandler(w, r, err)
				return
			}

			// Store tenant in context
			ctx := tenantcontext.WithTenant(r.Context(), tenant)
			r = r.WithContext(ctx)

			// Add tenant to response headers for debugging
			w.Header().Set("X-Tenant-Name", tenant.Name)

			// Propagate tenant to tracing span if available
			tenantcontext.PropagateToSpan(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
