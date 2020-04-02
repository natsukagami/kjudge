package verify

import (
	"database/sql"
)

// IntVerify are int verifiers.
type IntVerify func(int) error

// Int verifies ints.
func Int(i int, verifiers ...IntVerify) error {
	for _, v := range verifiers {
		if err := v(i); err != nil {
			return err
		}
	}
	return nil
}

// NullInt verifies a NullInt, if it is valid.
func NullInt(i sql.NullInt64, verifiers ...IntVerify) error {
	if i.Valid {
		return Int(int(i.Int64), verifiers...)
	}
	return nil
}

// IntPositive verifies that an int is positive.
func IntPositive(i int) error {
	return Int(i, IntMin(1))
}

// IntRange verifies that an int is in range.
func IntRange(low, high int) IntVerify {
	return func(i int) error {
		return Int(i, IntMin(low), IntMax(high))
	}
}

// IntMin verifies that the int is at least l.
func IntMin(l int) IntVerify {
	return func(i int) error {
		if i < l {
			return Errorf("must be at least %v", l)
		}
		return nil
	}
}

// IntMax verifies that the int is at most l.
func IntMax(l int) IntVerify {
	return func(i int) error {
		if i > l {
			return Errorf("must be at most %v", l)
		}
		return nil
	}
}
