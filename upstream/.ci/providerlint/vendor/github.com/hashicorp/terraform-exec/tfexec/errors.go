// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfexec

import (
	"context"
	"fmt"
)

// this file contains non-parsed exported errors

type ErrNoSuitableBinary struct {
	err error
}

func (e *ErrNoSuitableBinary) Error() string {
	return fmt.Sprintf("no suitable terraform binary could be found: %s", e.err.Error())
}

func (e *ErrNoSuitableBinary) Unwrap() error {
	return e.err
}

// ErrVersionMismatch is returned when the detected Terraform version is not compatible with the
// command or flags being used in this invocation.
type ErrVersionMismatch struct {
	MinInclusive string
	MaxExclusive string
	Actual       string
}

func (e *ErrVersionMismatch) Error() string {
	return fmt.Sprintf("unexpected version %s (min: %s, max: %s)", e.Actual, e.MinInclusive, e.MaxExclusive)
}

// ErrManualEnvVar is returned when an env var that should be set programatically via an option or method
// is set via the manual environment passing functions.
type ErrManualEnvVar struct {
	Name string
}

func (err *ErrManualEnvVar) Error() string {
	return fmt.Sprintf("manual setting of env var %q detected", err.Name)
}

// cmdErr is a custom error type to be returned when a cmd exits with a context
// error such as context.Canceled or context.DeadlineExceeded.
// The type is specifically designed to respond true to errors.Is for these two
// errors.
// See https://github.com/golang/go/issues/21880 for why this is necessary.
type cmdErr struct {
	err    error
	ctxErr error
}

func (e cmdErr) Is(target error) bool {
	switch target {
	case context.DeadlineExceeded, context.Canceled:
		return e.ctxErr == context.DeadlineExceeded || e.ctxErr == context.Canceled
	}
	return false
}

func (e cmdErr) Error() string {
	return e.err.Error()
}
