package verify

import "errors"

// NotNull verifies that the []byte is not null.
func NotNull(b []byte) error {
	if b == nil {
		return errors.New("cannot be empty or null")
	}
	return nil
}
