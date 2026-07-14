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

func waitFunctionCreated(ctx context.Context, conn *webfunctions.Client, name string, timeout time.Duration) (*webfunctions.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FunctionStatePending),
		Target:  enum.Slice(awstypes.FunctionStateActive),
		Refresh: statusFunction(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*webfunctions.GetWebFunctionOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionUpdated(ctx context.Context, conn *webfunctions.Client, name string, timeout time.Duration) (*webfunctions.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FunctionStatePending),
		Target:  enum.Slice(awstypes.FunctionStateActive),
		Refresh: statusFunction(ctx, conn, name),
		Timeout: timeout,
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
		Pending: enum.Slice(awstypes.EndpointStatePending),
		Target:  enum.Slice(awstypes.EndpointStateActive),
		Refresh: statusEndpoint(ctx, conn, functionName, endpointName),
		Timeout: timeout,
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
