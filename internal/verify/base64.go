// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"encoding/base64"
)

// Base64Encode encodes data if the input isn't already encoded using base64.StdEncoding.EncodeToString.
// If the input is already base64 encoded, return the original input unchanged.
func Base64Encode(data []byte) string {
	// Check whether the data is already Base64 encoded; don't double-encode
	if IsBase64Encoded(data) {
		return string(data)
	}
	// data has not been encoded encode and return
	return base64.StdEncoding.EncodeToString(data)
}

func IsBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))
	return err == nil
}
