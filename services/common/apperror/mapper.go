package apperror

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func MapPostgresError(err error, context string) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return mapPgError(pgErr, context)
	}

	return Database("database operation failed: "+context, err)
}

func mapPgError(pgErr *pgconn.PgError, context string) AppError {
	switch pgErr.Code {
	case "23505": // unique_violation
		return mapUniqueViolation(pgErr)

	case "23503": // foreign_key_violation
		return BadRequest(CodeValidationFailed, "referenced resource does not exist")

	case "23514": // check_violation
		return mapCheckViolation(pgErr)

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

func mapUniqueViolation(pgErr *pgconn.PgError) AppError {
	switch pgErr.ConstraintName {
	case "idx_users_email_unique":
		return EmailAlreadyExists()
	case "idx_users_username_unique":
		return UsernameAlreadyExists()
	default:
		return Conflict(CodeAlreadyExists, "", "resource already exists")
	}
}

func mapCheckViolation(pgErr *pgconn.PgError) AppError {
	switch pgErr.ConstraintName {
	case "no_self_friendship":
		return BadRequest(CodeCannotBefriendSelf, "cannot befriend yourself")
	case "no_self_message":
		return BadRequest(CodeCannotMessageSelf, "cannot message yourself")
	case "content_not_empty":
		return Validation(CodeFieldRequired, "content", "content cannot be empty")
	case "content_max_length":
		return Validation(CodeFieldTooLong, "content", "content is too long")
	default:
		return BadRequest(CodeValidationFailed, "validation constraint violated")
	}
}
