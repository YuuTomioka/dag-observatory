package marketdata

import (
	"errors"
	"fmt"

	apprepository "dag-observatory/dag-core/internal/application/marketdata/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %v", apprepository.ErrNotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: %s", apprepository.ErrConflict, pgErr.ConstraintName)
		case "23503", "23514", "22P02":
			return fmt.Errorf("%w: %s", apprepository.ErrInvalidArgument, pgErr.Message)
		}
	}

	return err
}
