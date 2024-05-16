// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitApplicationCreated(ctx context.Context, conn *qbusiness.Client, id string, timeout time.Duration) (*qbusiness.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ApplicationStatusCreating, types.ApplicationStatusUpdating),
		Target:     enum.Slice(types.ApplicationStatusActive),
		Refresh:    statusAppAvailability(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitApplicationDeleted(ctx context.Context, conn *qbusiness.Client, id string, timeout time.Duration) (*qbusiness.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ApplicationStatusActive, types.ApplicationStatusDeleting),
		Target:     []string{},
		Refresh:    statusAppAvailability(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitIndexCreated(ctx context.Context, conn *qbusiness.Client, index_id string, timeout time.Duration) (*qbusiness.GetIndexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.IndexStatusCreating, types.IndexStatusUpdating),
		Target:     enum.Slice(types.IndexStatusActive),
		Refresh:    statusIndexAvailability(ctx, conn, index_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetIndexOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitIndexUpdated(ctx context.Context, conn *qbusiness.Client, index_id string, timeout time.Duration) (*qbusiness.GetIndexOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.IndexStatusCreating, types.IndexStatusUpdating),
		Target:     enum.Slice(types.IndexStatusActive),
		Refresh:    statusIndexAvailability(ctx, conn, index_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetIndexOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitIndexDeleted(ctx context.Context, conn *qbusiness.Client, index_id string, timeout time.Duration) (*qbusiness.GetIndexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.IndexStatusActive, types.IndexStatusDeleting),
		Target:     []string{},
		Refresh:    statusIndexAvailability(ctx, conn, index_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetIndexOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitRetrieverCreated(ctx context.Context, conn *qbusiness.Client, retriever_id string, timeout time.Duration) (*qbusiness.GetRetrieverOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.RetrieverStatusCreating),
		Target:     enum.Slice(types.RetrieverStatusActive),
		Refresh:    statusRetrieverAvailability(ctx, conn, retriever_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetRetrieverOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitRetrieverDeleted(ctx context.Context, conn *qbusiness.Client, retriever_id string, timeout time.Duration) (*qbusiness.GetRetrieverOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.RetrieverStatusActive),
		Target:     []string{},
		Refresh:    statusRetrieverAvailability(ctx, conn, retriever_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetRetrieverOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}
