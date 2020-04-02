package verify

// NotNull verifies that the []byte is not null.
func NotNull(b []byte) error {
	if b == nil {
		return Errorf("cannot be empty or null")
	}
	return nil
}
