// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"context"
	"errors"
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
		for i, rev := range out.Revisions {
			if latest == nil || aws.ToString(rev.CreatedAt) > aws.ToString(latest.CreatedAt) {
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
		if reason := aws.ToString(out.StateReason); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
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
		if reason := aws.ToString(out.UpdateStatusReason); reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitEndpointDeleted(ctx context.Context, conn *lambdaweb.Client, functionName, endpointName string, timeout time.Duration) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EndpointStateActive, awstypes.EndpointStateDeleting),
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
