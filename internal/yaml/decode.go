// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package yaml

import (
	"bytes"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// DecodeFromBytes decodes (unmarshals) the given byte slice, containing valid YAML, into `to`.
func DecodeFromBytes(b []byte, to any) error {
	return DecodeFromReader(bytes.NewReader(b), to)
}

// DecodeFromReader decodes (unmarshals) the given io.Reader, pointing to a YAML stream, into `to`.
func DecodeFromReader(r io.Reader, to any) error {
	dec := yaml.NewDecoder(r)

	for {
		if err := dec.Decode(to); err == io.EOF { //nolint:errorlint // io.EOF is returned unwrapped
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

// DecodeFromString decodes (unmarshals) the given string, containing valid YAML, into `to`.
func DecodeFromString(s string, to any) error {
	return DecodeFromReader(strings.NewReader(s), to)
}
