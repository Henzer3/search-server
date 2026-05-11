package core

import (
	"errors"
)

var (
	ErrBadArguments    = errors.New("arguments are not acceptable")
	ErrAlreadyExists   = errors.New("resource or task already exists")
	ErrLimit           = errors.New("too much message")
	ErrAlreadyUpdating = errors.New("already updating")
	ErrNotFound        = errors.New("not found")
	ErrNoPermissions   = errors.New("no permissions")
)
