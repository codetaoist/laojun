package notification

import "errors"

// Common errors for notification operations.
var (
	ErrNotFound     = errors.New("notification: not found")
	ErrInvalidInput = errors.New("notification: invalid input")
	ErrTimeout      = errors.New("notification: operation timeout")
	ErrConnection   = errors.New("notification: connection failed")
	ErrNotSupported = errors.New("notification: operation not supported")
	
	// TODO: Add your specific errors here
)

// IsNotFound checks if the error is a "not found" error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsTimeout checks if the error is a timeout error.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// TODO: Add more error checking functions as needed
