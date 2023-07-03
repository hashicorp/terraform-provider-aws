// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// +build appengine

package msgpack

// bytesToString converts byte slice to string.
func bytesToString(b []byte) string {
	return string(b)
}

// stringToBytes converts string to byte slice.
func stringToBytes(s string) []byte {
	return []byte(s)
}
