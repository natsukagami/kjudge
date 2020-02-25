package isolate

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Meta is a meta-file parser.
type Meta struct {
	Fields map[string]string
	error
}

// ReadMetaFile reads a meta-file and returns a parser ready to parse.
func ReadMetaFile(file string) (*Meta, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	fields := make(map[string]string)

	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) < 2 {
			continue
		}
		fields[parts[0]] = parts[1]
	}

	return &Meta{Fields: fields, error: nil}, nil
}

// Int parses an int.
func (m *Meta) Int(key string) int {
	if m.error != nil {
		return 0
	}
	if val, ok := m.Fields[key]; ok {
		r, err := strconv.Atoi(val)
		m.error = errors.Wrapf(err, "key %s", key)
		return r
	} else {
		m.error = errors.Errorf("key %s: does not exist")
		return 0
	}
}

// Float64 parses an float64.
func (m *Meta) Float64(key string) float64 {
	if m.error != nil {
		return 0
	}
	if val, ok := m.Fields[key]; ok {
		r, err := strconv.ParseFloat(val, 64)
		m.error = errors.Wrapf(err, "key %s", key)
		return r
	} else {
		m.error = errors.Errorf("key %s: does not exist")
		return 0
	}
}

// String parses a string.
func (m *Meta) String(key string) string {
	if m.error != nil {
		return ""
	}
	if val, ok := m.Fields[key]; ok {
		return val
	}
	m.error = errors.Errorf("key %s: does not exist")
	return ""
}

// Error returns the error encountered in parsing.
func (m *Meta) Error() error { return m.error }
