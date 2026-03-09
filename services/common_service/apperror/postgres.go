package apperror

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

// ConstraintMapper allows a service to register its own constraint-names and corresponding AppErrors.
// Optionally passed to MapPostgresError.
type ConstraintMapper func(constraintName string) AppError

var constraintCodes = map[string]bool{
	"23505": true, // unique_violation
	"23503": true, // foreign_key_violation
	"23514": true, // check_violation
	"23502": true, // not_null_violation
}

func MapPostgresError(err error, context string, mapConstraint ...ConstraintMapper) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return Database("database operation failed: "+context, err)
	}
	// calling service mapper only for constraint errors - only they have a name
	if len(mapConstraint) > 0 && mapConstraint[0] != nil && constraintCodes[pgErr.Code] {
		if appErr := mapConstraint[0](pgErr.ConstraintName); appErr != nil {
			return appErr
		}
	}

	return mapGenericPgError(pgErr, context)
}

func mapGenericPgError(pgErr *pgconn.PgError, context string) AppError {
	switch pgErr.Code {
	case "23505": // unique_violation
		return Conflict(CodeAlreadyExists, "", "resource already exists")

	case "23503": // foreign_key_violation
		return BadRequest(CodeFieldInvalid, "referenced resource does not exist")

	case "23514": // check_violation — fallback
		return BadRequest(CodeValidationFailed, "constraint violation")

	case "23502": // not_null_violation
		return Validation(CodeFieldRequired, pgErr.ColumnName, "field is required")

	case "22001": // string_data_right_truncation
		return Validation(CodeFieldTooLong, pgErr.ColumnName, "field value is too long")

	case "08000", "08003", "08006": // connection errors
		return Database("database connection error: "+context, pgErr)

	case "57P03": // cannot_connect_now
		return Database("database is not accepting connections: "+context, pgErr)

	case "53300": // too_many_connections
		return Database("too many database connections: "+context, pgErr)

	default:
		return Database("database error: "+context, pgErr)
	}
}
