package verify

import "fmt"

// Error is error by verify
type Error struct {
	err string
}

// Error implements the error interface
func (e Error) Error() string {
	return e.err
}

// Errorf creates a verify error
func Errorf(s string, args ...interface{}) error {
	return Error{
		err: fmt.Sprintf(s, args...),
	}
}
