package errors

import "errors"

var (
	ErrInvalidValue              = errors.New("invalid value")
	ErrUnsupportedFunctionFields = errors.New("function fields are not supported")
)
