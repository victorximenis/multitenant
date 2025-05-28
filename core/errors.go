package core

import "fmt"

// TenantNotFoundError represents an error when a tenant is not found
type TenantNotFoundError struct {
	Name string
}

// Error implements the error interface for TenantNotFoundError
func (e TenantNotFoundError) Error() string {
	return fmt.Sprintf("tenant not found: %s", e.Name)
}

// TenantInactiveError represents an error when a tenant is inactive
type TenantInactiveError struct {
	Name string
}

// Error implements the error interface for TenantInactiveError
func (e TenantInactiveError) Error() string {
	return fmt.Sprintf("tenant is inactive: %s", e.Name)
}
