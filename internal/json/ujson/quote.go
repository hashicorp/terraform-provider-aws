// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package ujson

import (
	"bytes"
	"strconv"
	"unicode/utf8"
	"unsafe"
)

// ErrSyntax indicates that the value has invalid syntax.
var ErrSyntax = strconv.ErrSyntax

// AppendQuote appends a double-quoted string valid for json key and value, to
// dst and returns the extended buffer.
func AppendQuote(dst []byte, s []byte) []byte {
	return strconv.AppendQuote(dst, unsafeBytesToString(s))
}

// AppendQuoteToASCII appends a double-quoted string valid for json key and
// value, to dst and returns the extended buffer.
func AppendQuoteToASCII(dst []byte, s []byte) []byte {
	return strconv.AppendQuoteToASCII(dst, unsafeBytesToString(s))
}

// AppendQuoteToGraphic appends a double-quoted string valid for json key and
// value, to dst and returns the extended buffer.
func AppendQuoteToGraphic(dst []byte, s []byte) []byte {
	return strconv.AppendQuoteToGraphic(dst, unsafeBytesToString(s))
}

// AppendQuoteString returns a double-quoted string valid for json key or value.
func AppendQuoteString(dst []byte, s string) []byte {
	return strconv.AppendQuote(dst, s)
}

// Unquote decodes a double-quoted string key or value to retrieve the
// original string value. It will avoid allocation whenever possible.
//
// The code is inspired by strconv.Unquote, but only accepts valid json string.
func Unquote(s []byte) ([]byte, error) {
	n := len(s)
	if n < 2 {
		return nil, ErrSyntax
	}
	if s[0] != '"' || s[n-1] != '"' {
		return nil, ErrSyntax
	}
	s = s[1 : n-1]
	if bytes.IndexByte(s, '\n') != -1 {
		return nil, ErrSyntax
	}

	// avoid allocation if the string is trivial
	if bytes.IndexByte(s, '\\') == -1 {
		if utf8.Valid(s) {
			return s, nil
		}
	}

	// the following code is taken from strconv.Unquote (with modification)
	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	for len(s) > 0 {
		// Convert []byte to string for satisfying UnquoteChar. We won't keep
		// the retured string, so it's safe to use unsafe here.
		c, multibyte, tail, err := strconv.UnquoteChar(unsafeBytesToString(s), '"')
		if err != nil {
			return nil, err
		}

		// UnquoteChar returns tail as the remaining unprocess string. Because
		// we are processing []byte, we use len(tail) to get the remaining bytes
		// instead.
		s = s[len(s)-len(tail):]
		if c < utf8.RuneSelf || !multibyte {
			buf = append(buf, byte(c))
		} else {
			n = utf8.EncodeRune(runeTmp[:], c)
			buf = append(buf, runeTmp[:n]...)
		}
	}
	return buf, nil
}

//go:nosplit
func unsafeBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
