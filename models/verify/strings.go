// Package verify implements certain verification methods for simple data structures.
package verify

import (
	"regexp"
	"strings"
)

// StringVerify are string verifiers.
type StringVerify func(string) error

// Names specify that the names must be a non-empty string of maximum length 32.
func Names(s string) error {
	return String(s, StringNonEmpty, StringMaxLength(32))
}

// Password specify that the passwords must be a string of length between 9 and 32.
func Password(s string) error {
	return String(s, StringMinLength(9), StringMaxLength(32))
}

// String runs verification on a string against a list of verifiers.
func String(s string, verifiers ...StringVerify) error {
	for _, v := range verifiers {
		if err := v(s); err != nil {
			return err
		}
	}
	return nil
}

// StringNonEmpty verifies that a string is not empty.
func StringNonEmpty(s string) error {
	return StringMinLength(1)(s)
}

// StringMinLength verifies that a string must have a minimum length of "l".
func StringMinLength(l int) StringVerify {
	return func(s string) error {
		if len(s) < l {
			return Errorf("must have a length of at least %v", l)
		}
		return nil
	}
}

// StringMaxLength verifies that a string must have a maximum length of "l".
func StringMaxLength(l int) StringVerify {
	return func(s string) error {
		if len(s) > l {
			return Errorf("must have a length of at most %v", l)
		}
		return nil
	}
}

// Regexp verifies that a string matches a regular expression exactly.
func Regexp(r *regexp.Regexp) StringVerify {
	return func(s string) error {
		loc := r.FindStringIndex(s)
		if loc == nil || s[loc[0]:loc[1]] != s {
			return Errorf("must match the regular expression `%s`", r.String())
		}
		return nil
	}
}

// Enum verifies that a string is in a set of values.
func Enum(values ...string) StringVerify {
	return func(s string) error {
		for _, value := range values {
			if value == s {
				return nil
			}
		}
		return Errorf("value must be in [%v]", strings.Join(values, ", "))
	}
}
