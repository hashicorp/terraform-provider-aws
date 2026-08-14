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

func findFunctionByName(ctx context.Context, conn *lambdaweb.Client, name string) (*lambdaweb.GetWebFunctionOutput, error) {
	input := lambdaweb.GetWebFunctionInput{
		FunctionName: aws.String(name),
	}

	out, err := conn.GetWebFunction(ctx, &input)
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

func findRevisionByID(ctx context.Context, conn *lambdaweb.Client, functionName, revisionID string) (*lambdaweb.GetWebFunctionRevisionOutput, error) {
	input := lambdaweb.GetWebFunctionRevisionInput{
		FunctionName: aws.String(functionName),
		RevisionId:   aws.String(revisionID),
	}

	out, err := conn.GetWebFunctionRevision(ctx, &input)
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

func findLatestRevisionID(ctx context.Context, conn *lambdaweb.Client, functionName string) (*string, error) {
	input := lambdaweb.ListWebFunctionRevisionsInput{
		FunctionName: aws.String(functionName),
	}

	var latest *awstypes.FunctionRevisionSummary
	for {
		out, err := conn.ListWebFunctionRevisions(ctx, &input)
		if err != nil {
			return nil, smarterr.NewError(err)
		}
		for i := range out.Revisions {
			if latest == nil || createdBefore(aws.ToString(latest.CreatedAt), aws.ToString(out.Revisions[i].CreatedAt)) {
				latest = &out.Revisions[i]
			}
		}
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	if latest == nil {
		return nil, nil
	}
	return latest.RevisionId, nil
}

// createdBefore reports whether timestamp a is earlier than b. Timestamps are
// modeled as strings, so they are parsed rather than compared
// lexicographically: string comparison silently picks the wrong item if the
// service ever returns fractional seconds with variable precision. Values that
// cannot be parsed fall back to string comparison.
func createdBefore(a, b string) bool {
	at, aerr := time.Parse(time.RFC3339Nano, a)
	bt, berr := time.Parse(time.RFC3339Nano, b)
	if aerr == nil && berr == nil {
		return at.Before(bt)
	}

	return a < b
}

func statusFunction(conn *lambdaweb.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findFunctionByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
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

func statusRevision(conn *lambdaweb.Client, functionName, revisionID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findRevisionByID(ctx, conn, functionName, revisionID)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func waitRevisionActive(ctx context.Context, conn *lambdaweb.Client, functionName, revisionID string, timeout time.Duration) (*lambdaweb.GetWebFunctionRevisionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RevisionStatePending),
		Target:                    enum.Slice(awstypes.RevisionStateActive),
		Refresh:                   statusRevision(conn, functionName, revisionID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionRevisionOutput); ok {
		if reason := aws.ToString(out.StateReason); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionCreated(ctx context.Context, conn *lambdaweb.Client, name string, timeout time.Duration) (*lambdaweb.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.FunctionStatePending),
		Target:                    enum.Slice(awstypes.FunctionStateActive),
		Refresh:                   statusFunction(conn, name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionOutput); ok {
		if reason := aws.ToString(out.StateReason); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionUpdated(ctx context.Context, conn *lambdaweb.Client, name string, timeout time.Duration) (*lambdaweb.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.FunctionStatePending),
		Target:                    enum.Slice(awstypes.FunctionStateActive),
		Refresh:                   statusFunction(conn, name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionOutput); ok {
		if reason := aws.ToString(out.StateReason); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionDeleted(ctx context.Context, conn *lambdaweb.Client, name string, timeout time.Duration) (*lambdaweb.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FunctionStateActive, awstypes.FunctionStateDeleting),
		Target:  []string{},
		Refresh: statusFunction(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionOutput); ok {
		if reason := aws.ToString(out.StateReason); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
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
