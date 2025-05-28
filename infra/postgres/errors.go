package postgres

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error codes
const (
	// Unique violation
	UniqueViolationCode = "23505"
	// Foreign key violation
	ForeignKeyViolationCode = "23503"
	// Check constraint violation
	CheckViolationCode = "23514"
	// Not null violation
	NotNullViolationCode = "23502"
)

// mapPostgreSQLError maps PostgreSQL errors to application-level errors
func mapPostgreSQLError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		// Not a PostgreSQL error, return as-is
		return err
	}

	switch pgErr.Code {
	case UniqueViolationCode:
		return mapUniqueViolationError(pgErr)
	case ForeignKeyViolationCode:
		return mapForeignKeyViolationError(pgErr)
	case CheckViolationCode:
		return mapCheckViolationError(pgErr)
	case NotNullViolationCode:
		return mapNotNullViolationError(pgErr)
	default:
		// Return the original error with additional context
		return errors.New("database error: " + pgErr.Message)
	}
}

// mapUniqueViolationError maps unique constraint violations
func mapUniqueViolationError(pgErr *pgconn.PgError) error {
	if strings.Contains(pgErr.Detail, "tenants_name_key") {
		return errors.New("tenant name already exists")
	}
	return errors.New("unique constraint violation: " + pgErr.Detail)
}

// mapForeignKeyViolationError maps foreign key constraint violations
func mapForeignKeyViolationError(pgErr *pgconn.PgError) error {
	if strings.Contains(pgErr.Detail, "tenant_id") {
		return errors.New("referenced tenant does not exist")
	}
	return errors.New("foreign key constraint violation: " + pgErr.Detail)
}

// mapCheckViolationError maps check constraint violations
func mapCheckViolationError(pgErr *pgconn.PgError) error {
	if strings.Contains(pgErr.Detail, "role") {
		return errors.New("invalid datasource role: must be one of 'read', 'write', 'rw'")
	}
	return errors.New("check constraint violation: " + pgErr.Detail)
}

// mapNotNullViolationError maps not null constraint violations
func mapNotNullViolationError(pgErr *pgconn.PgError) error {
	return errors.New("required field cannot be null: " + pgErr.ColumnName)
}
