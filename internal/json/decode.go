// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// DecodeFromBytes decodes (unmarshals) the given byte slice, containing valid JSON, into `to`.
func DecodeFromBytes(b []byte, to any) error {
	return DecodeFromReader(bytes.NewReader(b), to)
}

// DecodeFromReader decodes (unmarshals) the given io.Reader, pointing to a JSON stream, into `to`.
func DecodeFromReader(r io.Reader, to any) error {
	dec := json.NewDecoder(r)

	for {
		if err := dec.Decode(to); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

// DecodeFromString decodes (unmarshals) the given string, containing valid JSON, into `to`.
func DecodeFromString(s string, to any) error {
	return DecodeFromReader(strings.NewReader(s), to)
}
