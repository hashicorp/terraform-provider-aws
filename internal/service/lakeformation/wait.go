// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

const (
	permissionsReadyTimeout       = 1 * time.Minute
	permissionsDeleteRetryTimeout = 30 * time.Second

	statusAvailable = "AVAILABLE"
	statusIAMDelay  = "IAM DELAY"
)

func waitPermissionsReady(ctx context.Context, conn *lakeformation.Client, input *lakeformation.ListPermissionsInput, filter PermissionsFilter) ([]awstypes.PrincipalResourcePermissions, error) {
	return waitPermissionsReadyWithRefresh(ctx, statusPermissions(conn, input, filter), permissionsReadyTimeout)
}

// waitPermissionsReadyWithRefresh waits for refresh to report matching permissions.
//
// With the default backoff, the timeout is reached before NotFoundChecks is exhausted, so
// permissions that consistently do not exist surface as a TimeoutError rather than a
// NotFoundError. To let callers distinguish "not found" from "could not be determined", a
// timeout is converted to a NotFoundError only if the most recent ListPermissions call
// completed successfully and found no matching permissions (or reported that the principal
// does not exist). A timeout whose most recent call failed (for example, because it was
// throttled until the deadline) is returned unchanged.
func waitPermissionsReadyWithRefresh(ctx context.Context, refresh retry.StateRefreshFunc, timeout time.Duration) ([]awstypes.PrincipalResourcePermissions, error) {
	var (
		emptyResults int
		lastEmpty    bool
	)

	stateConf := &retry.StateChangeConf{
		Pending: []string{statusIAMDelay},
		Target:  []string{statusAvailable},
		Refresh: func(ctx context.Context) (any, string, error) {
			// After the deadline, WaitForStateContext makes one more call with an already
			// expired context. That call observes nothing, so it must not change lastEmpty.
			if err := ctx.Err(); err != nil {
				return nil, "", err
			}

			output, status, err := refresh(ctx)

			// No matching permissions, or the principal does not (yet) exist.
			lastEmpty = err == nil && output == nil
			if lastEmpty {
				emptyResults++
			}

			return output, status, err
		},
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if retry.TimedOut(err) && lastEmpty {
		return nil, &retry.NotFoundError{
			LastError: err,
			Retries:   emptyResults,
		}
	}

	if output, ok := outputRaw.([]awstypes.PrincipalResourcePermissions); ok {
		return output, err
	}

	return nil, err
}
