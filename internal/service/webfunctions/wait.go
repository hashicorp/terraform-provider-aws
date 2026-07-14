// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package webfunctions

import (
	"context"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/webfunctions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/webfunctions/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findFunctionByName(ctx context.Context, conn *webfunctions.Client, name string) (*webfunctions.GetWebFunctionOutput, error) {
	input := webfunctions.GetWebFunctionInput{
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

func findEndpointByName(ctx context.Context, conn *webfunctions.Client, functionName, endpointName string) (*webfunctions.GetWebFunctionEndpointOutput, error) {
	input := webfunctions.GetWebFunctionEndpointInput{
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

func findRevisionByID(ctx context.Context, conn *webfunctions.Client, functionName, revisionID string) (*webfunctions.GetWebFunctionRevisionOutput, error) {
	input := webfunctions.GetWebFunctionRevisionInput{
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

func findLatestRevisionID(ctx context.Context, conn *webfunctions.Client, functionName string) (*string, error) {
	input := webfunctions.ListWebFunctionRevisionsInput{
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

func statusFunction(ctx context.Context, conn *webfunctions.Client, name string) retry.StateRefreshFunc {
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

func statusEndpoint(ctx context.Context, conn *webfunctions.Client, functionName, endpointName string) retry.StateRefreshFunc {
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

func statusRevision(ctx context.Context, conn *webfunctions.Client, functionName, revisionID string) retry.StateRefreshFunc {
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

func waitRevisionActive(ctx context.Context, conn *webfunctions.Client, functionName, revisionID string, timeout time.Duration) (*webfunctions.GetWebFunctionRevisionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RevisionStatePending),
		Target:                    enum.Slice(awstypes.RevisionStateActive),
		Refresh:                   statusRevision(ctx, conn, functionName, revisionID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*webfunctions.GetWebFunctionRevisionOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionCreated(ctx context.Context, conn *webfunctions.Client, name string, timeout time.Duration) (*webfunctions.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.FunctionStatePending),
		Target:                    enum.Slice(awstypes.FunctionStateActive),
		Refresh:                   statusFunction(ctx, conn, name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*webfunctions.GetWebFunctionOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionUpdated(ctx context.Context, conn *webfunctions.Client, name string, timeout time.Duration) (*webfunctions.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.FunctionStatePending),
		Target:                    enum.Slice(awstypes.FunctionStateActive),
		Refresh:                   statusFunction(ctx, conn, name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*webfunctions.GetWebFunctionOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionDeleted(ctx context.Context, conn *webfunctions.Client, name string, timeout time.Duration) (*webfunctions.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FunctionStateActive, awstypes.FunctionStateDeleting),
		Target:  []string{},
		Refresh: statusFunction(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*webfunctions.GetWebFunctionOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitEndpointActive(ctx context.Context, conn *webfunctions.Client, functionName, endpointName string, timeout time.Duration) (*webfunctions.GetWebFunctionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EndpointStatePending),
		Target:                    enum.Slice(awstypes.EndpointStateActive),
		Refresh:                   statusEndpoint(ctx, conn, functionName, endpointName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*webfunctions.GetWebFunctionEndpointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitEndpointDeleted(ctx context.Context, conn *webfunctions.Client, functionName, endpointName string, timeout time.Duration) (*webfunctions.GetWebFunctionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EndpointStateActive, awstypes.EndpointStateDeleting),
		Target:  []string{},
		Refresh: statusEndpoint(ctx, conn, functionName, endpointName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*webfunctions.GetWebFunctionEndpointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}
