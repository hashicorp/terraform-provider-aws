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

func waitPluginCreated(ctx context.Context, conn *qbusiness.Client, plugin_id string, timeout time.Duration) (*qbusiness.GetPluginOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.PluginBuildStatusCreateInProgress),
		Target:     enum.Slice(types.PluginBuildStatusReady),
		Refresh:    statusPluginAvailability(ctx, conn, plugin_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetPluginOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.BuildStatus)))

		return output, err
	}
	return nil, err
}

func waitPluginUpdated(ctx context.Context, conn *qbusiness.Client, plugin_id string, timeout time.Duration) (*qbusiness.GetPluginOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.PluginBuildStatusUpdateInProgress),
		Target:     enum.Slice(types.PluginBuildStatusReady),
		Refresh:    statusPluginAvailability(ctx, conn, plugin_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetPluginOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.BuildStatus)))

		return output, err
	}
	return nil, err
}

func waitPluginDeleted(ctx context.Context, conn *qbusiness.Client, plugin_id string, timeout time.Duration) (*qbusiness.GetPluginOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.PluginBuildStatusDeleteInProgress),
		Target:     []string{},
		Refresh:    statusPluginAvailability(ctx, conn, plugin_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetPluginOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.BuildStatus)))

		return output, err
	}
	return nil, err
}

func waitDatasourceCreated(ctx context.Context, conn *qbusiness.Client, datasource_id string, timeout time.Duration) (*qbusiness.GetDataSourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.DataSourceStatusCreating, types.DataSourceStatusPendingCreation),
		Target:     enum.Slice(types.DataSourceStatusActive),
		Refresh:    statusDatasourceAvailability(ctx, conn, datasource_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetDataSourceOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitDatasourceUpdated(ctx context.Context, conn *qbusiness.Client, datasource_id string, timeout time.Duration) (*qbusiness.GetDataSourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.DataSourceStatusUpdating),
		Target:     enum.Slice(types.DataSourceStatusActive),
		Refresh:    statusDatasourceAvailability(ctx, conn, datasource_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetDataSourceOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitDatasourceDeleted(ctx context.Context, conn *qbusiness.Client, datasource_id string, timeout time.Duration) (*qbusiness.GetDataSourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.DataSourceStatusActive, types.DataSourceStatusDeleting),
		Target:     []string{},
		Refresh:    statusDatasourceAvailability(ctx, conn, datasource_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetDataSourceOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitWebexperienceCreated(ctx context.Context, conn *qbusiness.Client, id string, timeout time.Duration) (*qbusiness.GetWebExperienceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.WebExperienceStatusCreating),
		Target:     enum.Slice(types.WebExperienceStatusActive, types.WebExperienceStatusPendingAuthConfig),
		Refresh:    statusWebexperienceAvailability(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetWebExperienceOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitWebexperienceDeleted(ctx context.Context, conn *qbusiness.Client, id string, timeout time.Duration) (*qbusiness.GetWebExperienceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WebExperienceStatusActive, types.WebExperienceStatusDeleting,
			types.WebExperienceStatusPendingAuthConfig, types.WebExperienceStatusFailed),
		Target:     []string{},
		Refresh:    statusWebexperienceAvailability(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetWebExperienceOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}
