package core_errors

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invlid argument")
	ErrConflict        = errors.New("conflict")
)
