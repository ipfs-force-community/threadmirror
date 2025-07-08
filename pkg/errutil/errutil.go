package errutil

import "errors"

// ErrNotFound is returned when a requested resource could not be located.
// It can be used with errors.Is to check for the not-found condition.
var ErrNotFound = errors.New("not found")
