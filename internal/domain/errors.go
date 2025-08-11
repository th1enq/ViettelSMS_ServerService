package domain

import "errors"

var (
	ErrServerExist = errors.New("server already exists with the same name or ID")
)
