// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// NotFound returns true if the error represents a "resource not found" condition.
// Specifically, NotFound returns true if the error or a wrapped error is of type
// retry.NotFoundError.
//
// Deprecated: NotFound is an alias to a function of the same name in internal/retry
// which handles both Plugin SDK V2 and internal error types. For net-new usage,
// prefer calling retry.NotFound directly.
var NotFound = retry.NotFound

// TimedOut returns true if the error represents a "wait timed out" condition.
// Specifically, TimedOut returns true if the error matches all these conditions:
//   - err is of type retry.TimeoutError
//   - TimeoutError.LastError is nil
//
// Deprecated: TimedOut is an alias to a function of the same name in internal/retry
// which handles both Plugin SDK V2 and internal error types. For net-new usage,
// prefer calling retry.TimedOut directly.
var TimedOut = retry.TimedOut

// SetLastError sets the LastError field on the error if supported.
// If lastErr is nil it is ignored.
//
// Deprecated: SetLastError is an alias to a function of the same name in internal/retry
// which handles both Plugin SDK V2 and internal error types. For net-new usage,
// prefer calling retry.SetLastError directly.
var SetLastError = retry.SetLastError
