// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"encoding/base64"
)

// Base64EncodeOnce encodes the input blob using base64.StdEncoding.EncodeToString.
// If the blob is already base64 encoded, return the original input unchanged.
func Base64EncodeOnce(blob []byte) string {
	if s := string(blob); IsBase64Encoded(s) {
		return s
	}

	return base64.StdEncoding.EncodeToString(blob)
}

// IsBase64Encoded checks if the input string is base64 encoded.
func IsBase64Encoded(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
