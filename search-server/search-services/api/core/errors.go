package core

import (
	"errors"
)

var ErrBadArguments = errors.New("arguments are not acceptable")
var ErrAlreadyExists = errors.New("resource or task already exists")
var ErrLimit = errors.New("too much message")
var ErrAlreadyUpdating = errors.New("already updating")
