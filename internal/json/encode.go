// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"bytes"
	"encoding/json"
)

// EncodeToBytes JSON encodes (marshals) `from` into a byte slice.
func EncodeToBytes(from any) ([]byte, error) {
	to, err := encodeToBuffer(from)

	if err != nil {
		return nil, err
	}

	return to.Bytes(), nil
}

// EncodeToString JSON encodes (marshals) `from` into a string.
func EncodeToString(from any) (string, error) {
	to, err := encodeToBuffer(from)

	if err != nil {
		return "", err
	}

	return to.String(), nil
}

func encodeToBuffer(from any) (*bytes.Buffer, error) {
	to := new(bytes.Buffer)
	enc := json.NewEncoder(to)

	if err := enc.Encode(from); err != nil {
		return nil, err
	}

	return to, nil
}
