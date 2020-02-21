package verify

import "github.com/pkg/errors"

// All verifies a map of fields.
func All(fields map[string]error) error {
	for f, err := range fields {
		if err != nil {
			return errors.Wrapf(err, "field %s", f)
		}
	}
	return nil
}
