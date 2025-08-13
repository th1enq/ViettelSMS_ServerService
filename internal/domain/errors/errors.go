package domain

import "errors"

var (
	ErrInternalServer = errors.New("internal server error")
	ErrServerExist    = errors.New("server already exists with the same name or ID")
	ErrServerNotFound = errors.New("server not found")

	ErrInvalidFile = errors.New("invalid file format or content")
)
