package repository

import "errors"

var (
	ErrNotFound        = errors.New("marketdata repository: not found")
	ErrConflict        = errors.New("marketdata repository: conflict")
	ErrInvalidArgument = errors.New("marketdata repository: invalid argument")
	ErrNotConfigured   = errors.New("marketdata repository: not configured")
)
