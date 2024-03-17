// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/aws/smithy-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	iamPropagationTimeout   = 2 * time.Minute
	dataSourceCreateTimeout = 5 * time.Minute
	dataSourceUpdateTimeout = 5 * time.Minute
)

// waitCreated waits for a DataSource to return CREATION_SUCCESSFUL
func waitCreated(ctx context.Context, conn *quicksight.Client, accountId, dataSourceId string) (*types.DataSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ResourceStatusCreationInProgress),
		Target:  enum.Slice(types.ResourceStatusCreationSuccessful),
		Refresh: status(ctx, conn, accountId, dataSourceId),
		Timeout: dataSourceCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DataSource); ok {
		if status, errorInfo := output.Status, output.ErrorInfo; status == types.ResourceStatusCreationFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", flex.StringValueToFramework(ctx, errorInfo.Type).String(), aws.ToString(errorInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

// waitUpdated waits for a DataSource to return UPDATE_SUCCESSFUL
func waitUpdated(ctx context.Context, conn *quicksight.Client, accountId, dataSourceId string) (*types.DataSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ResourceStatusUpdateInProgress),
		Target:  enum.Slice(types.ResourceStatusUpdateSuccessful),
		Refresh: status(ctx, conn, accountId, dataSourceId),
		Timeout: dataSourceUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DataSource); ok {
		if status, errorInfo := output.Status, output.ErrorInfo; status == types.ResourceStatusUpdateFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", flex.StringValueToFramework(ctx, errorInfo.Type).String(), aws.ToString(errorInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTemplateCreated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*types.Template, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ResourceStatusCreationInProgress),
		Target:                    enum.Slice(types.ResourceStatusCreationSuccessful),
		Refresh:                   statusTemplate(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.Template); ok {
		if status, apiErrors := out.Version.Status, out.Version.Errors; status == types.ResourceStatusCreationFailed {
			var errs []error
			for _, apiError := range apiErrors {
				genericSmithyError := &smithy.GenericAPIError{
					Code:    flex.StringValueToFramework(ctx, apiError.Type).String(),
					Message: aws.ToString(apiError.Message),
					Fault:   0,
				}
				errs = append(errs, genericSmithyError)
			}
			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitTemplateUpdated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*types.Template, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ResourceStatusUpdateInProgress, types.ResourceStatusCreationInProgress),
		Target:                    enum.Slice(types.ResourceStatusUpdateSuccessful, types.ResourceStatusCreationSuccessful),
		Refresh:                   statusTemplate(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.Template); ok {
		if status, apiErrors := out.Version.Status, out.Version.Errors; status == types.ResourceStatusCreationFailed {
			var errs []error
			for _, apiError := range apiErrors {
				genericSmithyError := &smithy.GenericAPIError{
					Code:    flex.StringValueToFramework(ctx, apiError.Type).String(),
					Message: aws.ToString(apiError.Message),
					Fault:   0,
				}
				errs = append(errs, genericSmithyError)
			}
			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitDashboardCreated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*types.Dashboard, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ResourceStatusCreationInProgress),
		Target:                    enum.Slice(types.ResourceStatusCreationSuccessful),
		Refresh:                   statusDashboard(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.Dashboard); ok {
		if status, apiErrors := out.Version.Status, out.Version.Errors; status == types.ResourceStatusCreationFailed {
			var errs []error
			for _, apiError := range apiErrors {
				genericSmithyError := &smithy.GenericAPIError{
					Code:    flex.StringValueToFramework(ctx, apiError.Type).String(),
					Message: aws.ToString(apiError.Message),
					Fault:   0,
				}
				errs = append(errs, genericSmithyError)
			}
			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitDashboardUpdated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*types.Dashboard, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ResourceStatusUpdateInProgress, types.ResourceStatusCreationInProgress),
		Target:                    enum.Slice(types.ResourceStatusUpdateSuccessful, types.ResourceStatusCreationSuccessful),
		Refresh:                   statusDashboard(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.Dashboard); ok {
		if status, apiErrors := out.Version.Status, out.Version.Errors; status == types.ResourceStatusCreationFailed {
			var errs []error
			for _, apiError := range apiErrors {
				genericSmithyError := &smithy.GenericAPIError{
					Code:    flex.StringValueToFramework(ctx, apiError.Type).String(),
					Message: aws.ToString(apiError.Message),
					Fault:   0,
				}
				errs = append(errs, genericSmithyError)
			}
			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitAnalysisCreated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*types.Analysis, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ResourceStatusCreationInProgress),
		Target:                    enum.Slice(types.ResourceStatusCreationSuccessful),
		Refresh:                   statusAnalysis(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.Analysis); ok {
		if status, apiErrors := out.Status, out.Errors; status == types.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error
			for _, apiError := range apiErrors {
				genericSmithyError := &smithy.GenericAPIError{
					Code:    flex.StringValueToFramework(ctx, apiError.Type).String(),
					Message: aws.ToString(apiError.Message),
					Fault:   0,
				}
				errs = append(errs, genericSmithyError)
			}
			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitAnalysisUpdated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*types.Analysis, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ResourceStatusUpdateInProgress, types.ResourceStatusCreationInProgress),
		Target:                    enum.Slice(types.ResourceStatusUpdateSuccessful, types.ResourceStatusCreationSuccessful),
		Refresh:                   statusAnalysis(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.Analysis); ok {
		if status, apiErrors := out.Status, out.Errors; status == types.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error
			for _, apiError := range apiErrors {
				genericSmithyError := &smithy.GenericAPIError{
					Code:    flex.StringValueToFramework(ctx, apiError.Type).String(),
					Message: aws.ToString(apiError.Message),
					Fault:   0,
				}
				errs = append(errs, genericSmithyError)
			}
			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitThemeCreated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*types.Theme, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ResourceStatusCreationInProgress),
		Target:                    enum.Slice(types.ResourceStatusCreationSuccessful),
		Refresh:                   statusTheme(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.Theme); ok {
		if status, apiErrors := out.Version.Status, out.Version.Errors; status == types.ResourceStatusCreationFailed {
			var errs []error
			for _, apiError := range apiErrors {
				genericSmithyError := &smithy.GenericAPIError{
					Code:    flex.StringValueToFramework(ctx, apiError.Type).String(),
					Message: aws.ToString(apiError.Message),
					Fault:   0,
				}
				errs = append(errs, genericSmithyError)
			}
			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitThemeUpdated(ctx context.Context, conn *quicksight.Client, id string, timeout time.Duration) (*types.Theme, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ResourceStatusUpdateInProgress, types.ResourceStatusCreationInProgress),
		Target:                    enum.Slice(types.ResourceStatusUpdateSuccessful, types.ResourceStatusCreationSuccessful),
		Refresh:                   statusTheme(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.Theme); ok {
		if status, apiErrors := out.Version.Status, out.Version.Errors; status == types.ResourceStatusCreationFailed {
			var errs []error
			for _, apiError := range apiErrors {
				genericSmithyError := &smithy.GenericAPIError{
					Code:    flex.StringValueToFramework(ctx, apiError.Type).String(),
					Message: aws.ToString(apiError.Message),
					Fault:   0,
				}
				errs = append(errs, genericSmithyError)
			}
			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}
