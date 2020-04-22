package verify

// NotNull verifies that the []byte is not null.
func NotNull(b []byte) error {
	if len(b) == 0 {
		return Errorf("cannot be empty or null")
	}
	return nil
}
