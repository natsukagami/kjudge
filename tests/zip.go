// Package tests handles multiple test uploading.
package tests

import (
	"archive/zip"
	"io"
	"strings"

	"github.com/natsukagami/kjudge/models"
	"github.com/pkg/errors"
)

// LazyTest is a `models.Test` but input and outputs are `zip.File`s.
type LazyTest struct {
	models.Test
	Input, Output *zip.File
}

// Verify calls the underlying Verify method, respecting their wish to have input and output non-null.
func (lz *LazyTest) Verify() error {
	lz.Test.Input = []byte{}
	lz.Test.Output = []byte{}
	return lz.Test.Verify()
}

type textMap = map[string]*zip.File

// Unpack try to unpack a zip file and extract tests from the given pattern.
func Unpack(zipFile io.ReaderAt, size int64, input, output string) ([]*LazyTest, error) {
	file, err := zip.NewReader(zipFile, size)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	inP, err := ParsePattern(input)
	if err != nil {
		return nil, errors.Wrap(err, "input pattern")
	}
	outP, err := ParsePattern(output)
	if err != nil {
		return nil, errors.Wrap(err, "output pattern")
	}
	inputs := make(textMap)
	outputs := make(textMap)
	for _, f := range file.File {
		if name, ok := inP.Match(f.Name); ok {
			if err := readToMap(f, name, inputs); err != nil {
				return nil, errors.Wrapf(err, "file %s", f.Name)
			}
		}
		if name, ok := outP.Match(f.Name); ok {
			if err := readToMap(f, name, outputs); err != nil {
				return nil, errors.Wrapf(err, "file %s", f.Name)
			}
		}
	}
	return matchTests(inputs, outputs), nil
}

func matchTests(in, out textMap) []*LazyTest {
	var res []*LazyTest
	for name, input := range in {
		output, ok := out[name]
		if !ok {
			continue
		}
		res = append(res, &LazyTest{
			Test:   models.Test{Name: name},
			Input:  input,
			Output: output,
		})
	}
	return res
}

func readToMap(f *zip.File, name string, target textMap) error {
	if _, ok := target[name]; ok {
		return errors.Errorf("duplicate key %s", name)
	}
	target[name] = f
	return nil
}

// Pattern recognizes the patterns from an input string and try to match it.
type Pattern struct {
	Prefix string
	Suffix string
}

// ParsePattern parses a pattern string and create a Pattern struct.
// Given a pattern string, it matches the ONLY question mark from it and split the rest into a prefix and suffix.
func ParsePattern(pattern string) (*Pattern, error) {
	where := -1 // where is the question mark?
	for i, chr := range pattern {
		if chr != '?' {
			continue
		}
		if where != -1 {
			return nil, errors.New("pattern has too many question marks")
		}
		where = i
	}
	if where == -1 {
		return nil, errors.New("pattern does not have a question mark")
	}
	return &Pattern{
		Prefix: pattern[:where],
		Suffix: pattern[where+1:],
	}, nil
}

// Match tries to match a string, extracting its name.
func (p *Pattern) Match(target string) (string, bool) {
	if !strings.HasPrefix(target, p.Prefix) || !strings.HasSuffix(target, p.Suffix) {
		return "", false
	}
	if len(target) <= len(p.Prefix)+len(p.Suffix) {
		return "", false
	}
	return target[len(p.Prefix) : len(target)-len(p.Suffix)], true
}
