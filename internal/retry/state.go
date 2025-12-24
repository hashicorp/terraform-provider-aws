// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

//
// Based on https://github.com/hashicorp/terraform-plugin-sdk/helper/retry/state.go.
//

// StateRefreshFuncOf is a function type used for StateChangeConf that is
// responsible for refreshing the item being watched for a state change.
//
// It returns three results.
// The first is any object that will be returned as the final object after waiting for state change.
// The second is the latest state of that object.
// The third any error that may have happened while refreshing the state.
type StateRefreshFuncOf[T any, S ~string] func(context.Context) (T, S, error)

// StateRefreshFunc is the specialization used in all code using helper/retry.
type StateRefreshFunc = StateRefreshFuncOf[any, string]

// StateChangeConfOf is the configuration struct used for `WaitForState`.
type StateChangeConfOf[T any, S ~string] struct {
	Delay          time.Duration            // Wait this time before starting checks
	Pending        []S                      // States that are "allowed" and will continue trying
	Refresh        StateRefreshFuncOf[T, S] // Refreshes the current state
	Target         []S                      // Target state
	Timeout        time.Duration            // The amount of time to wait before timeout
	MinTimeout     time.Duration            // Smallest time to wait before refreshes
	PollInterval   time.Duration            // Override MinTimeout/backoff and only poll this often
	NotFoundChecks int                      // Number of times to allow not found (nil result from Refresh)

	// This is to work around inconsistent APIs
	ContinuousTargetOccurence int // Number of times the Target state has to occur continuously
}

// StateChangeConf is the specialization used in all code using helper/retry.
type StateChangeConf = StateChangeConfOf[any, string]

// WaitForState watches an object and waits for it to achieve the state
// specified in the configuration using the specified Refresh() func,
// waiting the number of seconds specified in the timeout configuration.
//
// If the Refresh function returns an error, exit immediately with that error.
//
// If the Refresh function returns a state other than the Target state or one
// listed in Pending, return immediately with an error.
//
// If the Timeout is exceeded before reaching the Target state, return an
// error.
//
// Otherwise, the result is the result of the first call to the Refresh function to
// reach the target state.
//
// Cancellation of the passed in context will cancel the refresh loop.
//
// When VCR testing is enabled in replay mode, the DelayFunc is overridden to
// allow interactions to be replayed with no delay between state change refreshes.
func (conf *StateChangeConfOf[T, S]) WaitForStateContext(ctx context.Context) (T, error) {
	// Set a default for times to check for not found.
	if conf.NotFoundChecks == 0 {
		conf.NotFoundChecks = 20
	}
	if conf.ContinuousTargetOccurence == 0 {
		conf.ContinuousTargetOccurence = 1
	}

	// Set a default Delay using the StateChangeConf values
	delay := backoff.SDKv2HelperRetryCompatibleDelay(conf.Delay, conf.PollInterval, conf.MinTimeout)

	// When VCR testing in replay mode, override the default Delay
	if inContext, ok := conns.FromContext(ctx); ok && inContext.VCREnabled() {
		if mode, _ := vcr.Mode(); mode == recorder.ModeReplayOnly {
			delay = backoff.ZeroDelay
		}
	}

	var (
		t                             T
		currentState                  S
		err                           error
		notFoundTick, targetOccurence int
		l                             *backoff.Loop
	)
	for l = backoff.NewLoopWithOptions(conf.Timeout, backoff.WithDelay(delay)); l.Continue(ctx); {
		t, currentState, err = conf.refreshWithTimeout(ctx, l.Remaining())

		if errors.Is(err, context.DeadlineExceeded) {
			break
		}

		if err != nil {
			return t, err
		}

		if inttypes.IsZero(t) {
			// If we're waiting for the absence of a thing, then return.
			if len(conf.Target) == 0 {
				targetOccurence++
				if conf.ContinuousTargetOccurence == targetOccurence {
					return t, err
				}

				continue
			}

			// If we didn't find the resource, check if we have been
			// not finding it for a while, and if so, report an error.
			notFoundTick++
			if notFoundTick > conf.NotFoundChecks {
				return t, &NotFoundError{
					LastError: err,
					Retries:   notFoundTick,
				}
			}
		} else {
			// Reset the counter for when a resource isn't found.
			notFoundTick = 0
			found := false

			if slices.Contains(conf.Target, currentState) {
				found = true
				targetOccurence++
				if conf.ContinuousTargetOccurence == targetOccurence {
					return t, err
				}
			}

			if slices.Contains(conf.Pending, currentState) {
				found = true
				targetOccurence = 0
			}

			if !found && len(conf.Pending) > 0 {
				return t, &UnexpectedStateError{
					LastError:     err,
					State:         string(currentState),
					ExpectedState: tfslices.Strings(conf.Target),
				}
			}

			// Wait between refreshes using exponential backoff, except when
			// waiting for the target state to reoccur.
			if v, ok := delay.(backoff.DelayWithSetIncrementDelay); ok {
				v.SetIncrementDelay(targetOccurence == 0)
			}
		}
	}

	// Timed out or Context canceled.
	if l.Remaining() == 0 {
		return inttypes.Zero[T](), &TimeoutError{
			LastError:     err,
			LastState:     string(currentState),
			Timeout:       conf.Timeout,
			ExpectedState: tfslices.Strings(conf.Target),
		}
	}

	return t, context.Cause(ctx)
}

func (conf *StateChangeConfOf[T, S]) refreshWithTimeout(ctx context.Context, timeout time.Duration) (T, S, error) {
	// Set a deadline on the context here to maintain compatibility with the Plugin SDKv2 implementation.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return conf.Refresh(ctx)
}
