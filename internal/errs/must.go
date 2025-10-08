// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs

// Must is a generic implementation of the Go Must idiom [1, 2]. It panics if
// the provided error is non-nil and returns x otherwise.
//
// [1]: https://pkg.go.dev/text/template#Must
// [2]: https://pkg.go.dev/regexp#MustCompile
func Must[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}
