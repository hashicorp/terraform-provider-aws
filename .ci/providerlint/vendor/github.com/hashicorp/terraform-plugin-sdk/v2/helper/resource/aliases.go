// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// Deprecated: Use helper/id package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
const UniqueIdPrefix = id.UniqueIdPrefix

// Helper for a resource to generate a unique identifier w/ default prefix
//
// Deprecated: Use helper/id package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
func UniqueId() string {
	return id.UniqueId()
}

// Deprecated: Use helper/id package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
const UniqueIDSuffixLength = id.UniqueIDSuffixLength

// Helper for a resource to generate a unique identifier w/ given prefix
//
// After the prefix, the ID consists of an incrementing 26 digit value (to match
// previous timestamp output).  After the prefix, the ID consists of a timestamp
// and an incrementing 8 hex digit value The timestamp means that multiple IDs
// created with the same prefix will sort in the order of their creation, even
// across multiple terraform executions, as long as the clock is not turned back
// between calls, and as long as any given terraform execution generates fewer
// than 4 billion IDs.
//
// Deprecated: Use helper/id package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
func PrefixedUniqueId(prefix string) string {
	return id.PrefixedUniqueId(prefix)
}

// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
type NotFoundError = retry.NotFoundError

// UnexpectedStateError is returned when Refresh returns a state that's neither in Target nor Pending
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
type UnexpectedStateError = retry.UnexpectedStateError

// TimeoutError is returned when WaitForState times out
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
type TimeoutError = retry.TimeoutError

// StateRefreshFunc is a function type used for StateChangeConf that is
// responsible for refreshing the item being watched for a state change.
//
// It returns three results. `result` is any object that will be returned
// as the final object after waiting for state change. This allows you to
// return the final updated object, for example an EC2 instance after refreshing
// it. A nil result represents not found.
//
// `state` is the latest state of that object. And `err` is any error that
// may have happened while refreshing the state.
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
type StateRefreshFunc = retry.StateRefreshFunc

// StateChangeConf is the configuration struct used for `WaitForState`.
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
type StateChangeConf = retry.StateChangeConf

// RetryFunc is the function retried until it succeeds.
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
type RetryFunc = retry.RetryFunc

// RetryContext is a basic wrapper around StateChangeConf that will just retry
// a function until it no longer returns an error.
//
// Cancellation from the passed in context will propagate through to the
// underlying StateChangeConf
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
func RetryContext(ctx context.Context, timeout time.Duration, f RetryFunc) error {
	return retry.RetryContext(ctx, timeout, f)
}

// Retry is a basic wrapper around StateChangeConf that will just retry
// a function until it no longer returns an error.
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
func Retry(timeout time.Duration, f RetryFunc) error {
	return retry.Retry(timeout, f)
}

// RetryError is the required return type of RetryFunc. It forces client code
// to choose whether or not a given error is retryable.
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
type RetryError = retry.RetryError

// RetryableError is a helper to create a RetryError that's retryable from a
// given error. To prevent logic errors, will return an error when passed a
// nil error.
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
func RetryableError(err error) *RetryError {
	r := retry.RetryableError(err)

	return &RetryError{
		Err:       r.Err,
		Retryable: r.Retryable,
	}
}

// NonRetryableError is a helper to create a RetryError that's _not_ retryable
// from a given error. To prevent logic errors, will return an error when
// passed a nil error.
//
// Deprecated: Use helper/retry package instead. This is required for migrating acceptance
// testing to terraform-plugin-testing.
func NonRetryableError(err error) *RetryError {
	r := retry.NonRetryableError(err)

	return &RetryError{
		Err:       r.Err,
		Retryable: r.Retryable,
	}
}
