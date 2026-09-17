// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"bytes"
	"encoding/json"
)

// CompactBytes compacts the given byte slice containing JSON by removing insignificant space characters.
func CompactBytes(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, b); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// CompactString compacts the given string containing JSON by removing insignificant space characters.`
func CompactString(s string) (string, error) {
	out, err := CompactBytes([]byte(s))
	if err != nil {
		return "", err
	}

	return string(out), nil
}
