// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findEndpointByName(ctx context.Context, conn *lambdaweb.Client, functionName, endpointName string) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	input := lambdaweb.GetWebFunctionEndpointInput{
		FunctionName: aws.String(functionName),
		EndpointName: aws.String(endpointName),
	}

	out, err := conn.GetWebFunctionEndpoint(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

func statusEndpoint(conn *lambdaweb.Client, functionName, endpointName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findEndpointByName(ctx, conn, functionName, endpointName)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func statusEndpointUpdate(conn *lambdaweb.Client, functionName, endpointName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findEndpointByName(ctx, conn, functionName, endpointName)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.UpdateStatus), nil
	}
}

func waitEndpointActive(ctx context.Context, conn *lambdaweb.Client, functionName, endpointName string, timeout time.Duration) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EndpointStatePending),
		Target:                    enum.Slice(awstypes.EndpointStateActive),
		Refresh:                   statusEndpoint(conn, functionName, endpointName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionEndpointOutput); ok {
		if reason := endpointStateReason(out); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// endpointStateReason expands an endpoint's state reason with the per-region
// reasons. When a MultiRegion or PerRegion endpoint fails in one of its regions
// the top-level reason only says to "check the regional endpoint states for more
// details", which a Terraform user never sees: the detail lives in
// regionalEndpoints. Appending it turns an opaque failed apply into one that
// names the region and the cause.
func endpointStateReason(out *lambdaweb.GetWebFunctionEndpointOutput) string {
	if out == nil {
		return ""
	}

	return appendRegionalReasons(aws.ToString(out.StateReason), out.RegionalEndpoints,
		func(regional awstypes.RegionalEndpoint) (string, bool) {
			if regional.State == awstypes.EndpointStateActive {
				return "", false
			}
			if reason := aws.ToString(regional.StateReason); reason != "" {
				return reason, true
			}
			return string(regional.State), true
		})
}

// endpointUpdateStatusReason is endpointStateReason for the update path, where
// the regional detail lives in updateStatusReason instead.
func endpointUpdateStatusReason(out *lambdaweb.GetWebFunctionEndpointOutput) string {
	if out == nil {
		return ""
	}

	return appendRegionalReasons(aws.ToString(out.UpdateStatusReason), out.RegionalEndpoints,
		func(regional awstypes.RegionalEndpoint) (string, bool) {
			if regional.UpdateStatus == awstypes.EndpointUpdateStatusSuccessful || regional.UpdateStatus == "" {
				return "", false
			}
			if reason := aws.ToString(regional.UpdateStatusReason); reason != "" {
				return reason, true
			}
			return string(regional.UpdateStatus), true
		})
}

func appendRegionalReasons(reason string, regionalEndpoints map[string]awstypes.RegionalEndpoint, detail func(awstypes.RegionalEndpoint) (string, bool)) string {
	var b strings.Builder
	b.WriteString(reason)

	for _, region := range slices.Sorted(maps.Keys(regionalEndpoints)) {
		if d, ok := detail(regionalEndpoints[region]); ok {
			fmt.Fprintf(&b, " [%s: %s]", region, d)
		}
	}

	return b.String()
}

func waitEndpointUpdated(ctx context.Context, conn *lambdaweb.Client, functionName, endpointName string, timeout time.Duration) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		// An empty updateStatus can be observed if the endpoint has never been
		// updated before and the PATCH has not propagated to the Get yet.
		Pending:                   append(enum.Slice(awstypes.EndpointUpdateStatusInProgress), ""),
		Target:                    enum.Slice(awstypes.EndpointUpdateStatusSuccessful),
		Refresh:                   statusEndpointUpdate(conn, functionName, endpointName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionEndpointOutput); ok {
		if reason := endpointUpdateStatusReason(out); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitEndpointDeleted(ctx context.Context, conn *lambdaweb.Client, functionName, endpointName string, timeout time.Duration) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		// Every state an endpoint can hold on the way out, because the target is
		// "gone" and anything unlisted aborts the destroy. A MultiRegion endpoint
		// sits Pending for minutes while it replicates and settles in Failed when
		// a region cannot take the revision, and both accept a delete: observing
		// either one is a matter of when the first refresh lands.
		Pending: enum.Slice(awstypes.EndpointStatePending, awstypes.EndpointStateActive,
			awstypes.EndpointStateFailed, awstypes.EndpointStateDeleting),
		Target:  []string{},
		Refresh: statusEndpoint(conn, functionName, endpointName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionEndpointOutput); ok {
		if reason := aws.ToString(out.StateReason); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}
