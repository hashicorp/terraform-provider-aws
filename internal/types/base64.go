// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"encoding/base64"

	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func MustBase64Decode(s string) []byte {
	return errs.Must(Base64Decode(s))
}

func Base64Encode(blob []byte) string {
	return base64.StdEncoding.EncodeToString(blob)
}

// Base64EncodeOnce encodes the input blob using base64.StdEncoding.EncodeToString.
// If the blob is already base64 encoded, return the original input unchanged.
func Base64EncodeOnce(blob []byte) string {
	if s := string(blob); IsBase64Encoded(s) {
		return s
	}

	return Base64Encode(blob)
}

// IsBase64Encoded checks if the input string is base64 encoded.
func IsBase64Encoded(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
