// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type WaitOpts struct {
	ContinuousTargetOccurence int           // Number of times the target state has to occur continuously.
	Delay                     time.Duration // Wait this time before starting checks.
	MinTimeout                time.Duration // Smallest time to wait before refreshes.
	PollInterval              time.Duration // Override MinTimeout/backoff and only poll this often.
}

const (
	targetStateError = "ERROR"
	targetStateFalse = "FALSE"
	targetStateTrue  = "TRUE"
)

// WaitUntil waits for the function `f` to return `true`.
// If `f` returns an error, return immediately with that error.
// If `timeout` is exceeded before `f` returns `true`, return an error.
// Waits between calls to `f` using exponential backoff, except when waiting for the target state to reoccur.
func WaitUntil(ctx context.Context, timeout time.Duration, f func() (bool, error), opts WaitOpts) error {
	refresh := func() (interface{}, string, error) {
		done, err := f()

		if err != nil {
			return nil, targetStateError, err
		}

		if done {
			return "", targetStateTrue, nil
		}

		return "", targetStateFalse, nil
	}

	stateConf := &retry.StateChangeConf{
		Pending:                   []string{targetStateFalse},
		Target:                    []string{targetStateTrue},
		Refresh:                   refresh,
		Timeout:                   timeout,
		ContinuousTargetOccurence: opts.ContinuousTargetOccurence,
		Delay:                     opts.Delay,
		MinTimeout:                opts.MinTimeout,
		PollInterval:              opts.PollInterval,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
