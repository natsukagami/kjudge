package models

// NormalizeEndingsNative normalize file line endings to the current OS's endings
func NormalizeEndingsNative(content []byte) ([]byte, error) {
	return crlftoLF(content)
}
