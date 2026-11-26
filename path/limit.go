package path

import "errors"

// ErrInvalidLimit is returned when a limit is below 1.
var ErrInvalidLimit = errors.New("limit is below 1")
