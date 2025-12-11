// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

type WaitOpts struct {
	ContinuousTargetOccurence int           // Number of times the target state has to occur continuously.
	Delay                     time.Duration // Wait this time before starting checks.
	MinTimeout                time.Duration // Smallest time to wait before refreshes.
	PollInterval              time.Duration // Override MinTimeout/backoff and only poll this often.
}

type targetState string

const (
	targetStateError targetState = "ERROR"
	targetStateFalse targetState = "FALSE"
	targetStateTrue  targetState = "TRUE"
)

const (
	// Required so that we're not returning a zero-value from the `refresh` function
	dummy string = "x"
)

// WaitUntil waits for the function `f` to return `true`.
// If `f` returns an error, return immediately with that error.
// If `timeout` is exceeded before `f` returns `true`, return an error.
// Waits between calls to `f` using exponential backoff.
func WaitUntil(ctx context.Context, timeout time.Duration, f func(context.Context) (bool, error), opts WaitOpts) error {
	refresh := func(ctx context.Context) (any, targetState, error) {
		done, err := f(ctx)

		if err != nil {
			return nil, targetStateError, err
		}

		if done {
			return dummy, targetStateTrue, nil
		}

		return dummy, targetStateFalse, nil
	}

	stateConf := &retry.StateChangeConfOf[any, targetState]{
		Pending:                   enum.EnumSlice(targetStateFalse),
		Target:                    enum.EnumSlice(targetStateTrue),
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
