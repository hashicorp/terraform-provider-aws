// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"bytes"
	"encoding/json"
)

// EncodeToBytes JSON encodes (marshals) `from` into a byte slice.
func EncodeToBytes(from any) ([]byte, error) {
	to, err := encodeToBuffer(from, "", "")

	if err != nil {
		return nil, err
	}

	return to.Bytes(), nil
}

// EncodeToString JSON encodes (marshals) `from` into a string.
func EncodeToString(from any) (string, error) {
	return EncodeToStringIndent(from, "", "")
}

// EncodeToString JSON encodes (marshals) `from` into a string applying the specified indentation.
func EncodeToStringIndent(from any, prefix, indent string) (string, error) {
	to, err := encodeToBuffer(from, prefix, indent)

	if err != nil {
		return "", err
	}

	return to.String(), nil
}

func encodeToBuffer(from any, prefix, indent string) (*bytes.Buffer, error) {
	to := new(bytes.Buffer)
	enc := json.NewEncoder(to)
	enc.SetIndent(prefix, indent)

	if err := enc.Encode(from); err != nil {
		return nil, err
	}

	return to, nil
}
