// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// +build appengine

package internal

func BytesToString(b []byte) string {
	return string(b)
}

func StringToBytes(s string) []byte {
	return []byte(s)
}
