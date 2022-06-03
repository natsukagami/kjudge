//go:build !production
// +build !production

package admin

// NewVersionMessageGet checks if there is a new version of kjudge
func NewVersionMessageGet() (string, error) {
	return "", nil
}
