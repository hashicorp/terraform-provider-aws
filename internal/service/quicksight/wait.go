// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	iamPropagationTimeout   = 2 * time.Minute
	dataSourceCreateTimeout = 5 * time.Minute
	dataSourceUpdateTimeout = 5 * time.Minute
)

// waitCreated waits for a DataSource to return CREATION_SUCCESSFUL
func waitCreated(ctx context.Context, conn *quicksight.QuickSight, accountId, dataSourceId string) (*quicksight.DataSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{quicksight.ResourceStatusCreationInProgress},
		Target:  []string{quicksight.ResourceStatusCreationSuccessful},
		Refresh: status(ctx, conn, accountId, dataSourceId),
		Timeout: dataSourceCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*quicksight.DataSource); ok {
		if status, errorInfo := aws.StringValue(output.Status), output.ErrorInfo; status == quicksight.ResourceStatusCreationFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(errorInfo.Type), aws.StringValue(errorInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

// waitUpdated waits for a DataSource to return UPDATE_SUCCESSFUL
func waitUpdated(ctx context.Context, conn *quicksight.QuickSight, accountId, dataSourceId string) (*quicksight.DataSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{quicksight.ResourceStatusUpdateInProgress},
		Target:  []string{quicksight.ResourceStatusUpdateSuccessful},
		Refresh: status(ctx, conn, accountId, dataSourceId),
		Timeout: dataSourceUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*quicksight.DataSource); ok {
		if status, errorInfo := aws.StringValue(output.Status), output.ErrorInfo; status == quicksight.ResourceStatusUpdateFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(errorInfo.Type), aws.StringValue(errorInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTemplateCreated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.Template, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{quicksight.ResourceStatusCreationInProgress},
		Target:                    []string{quicksight.ResourceStatusCreationSuccessful},
		Refresh:                   statusTemplate(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.Template); ok {
		if status, apiErrors := aws.StringValue(out.Version.Status), out.Version.Errors; status == quicksight.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error

			for _, apiError := range apiErrors {
				if apiError == nil {
					continue
				}
				errs = append(errs, awserr.New(aws.StringValue(apiError.Type), aws.StringValue(apiError.Message), nil))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitTemplateUpdated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.Template, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{quicksight.ResourceStatusUpdateInProgress, quicksight.ResourceStatusCreationInProgress},
		Target:                    []string{quicksight.ResourceStatusUpdateSuccessful, quicksight.ResourceStatusCreationSuccessful},
		Refresh:                   statusTemplate(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.Template); ok {
		if status, apiErrors := aws.StringValue(out.Version.Status), out.Version.Errors; status == quicksight.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error

			for _, apiError := range apiErrors {
				if apiError == nil {
					continue
				}

				errs = append(errs, awserr.New(aws.StringValue(apiError.Type), aws.StringValue(apiError.Message), nil))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitDashboardCreated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.Dashboard, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{quicksight.ResourceStatusCreationInProgress},
		Target:                    []string{quicksight.ResourceStatusCreationSuccessful},
		Refresh:                   statusDashboard(ctx, conn, id, DashboardLatestVersion),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.Dashboard); ok {
		if status, apiErrors := aws.StringValue(out.Version.Status), out.Version.Errors; status == quicksight.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error

			for _, apiError := range apiErrors {
				if apiError == nil {
					continue
				}

				errs = append(errs, awserr.New(aws.StringValue(apiError.Type), aws.StringValue(apiError.Message), nil))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitDashboardUpdated(ctx context.Context, conn *quicksight.QuickSight, id string, version int64, timeout time.Duration) (*quicksight.Dashboard, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{quicksight.ResourceStatusUpdateInProgress, quicksight.ResourceStatusCreationInProgress},
		Target:                    []string{quicksight.ResourceStatusUpdateSuccessful, quicksight.ResourceStatusCreationSuccessful},
		Refresh:                   statusDashboard(ctx, conn, id, version),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.Dashboard); ok {
		if status, apiErrors := aws.StringValue(out.Version.Status), out.Version.Errors; status == quicksight.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error

			for _, apiError := range apiErrors {
				if apiError == nil {
					continue
				}

				errs = append(errs, awserr.New(aws.StringValue(apiError.Type), aws.StringValue(apiError.Message), nil))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitAnalysisCreated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.Analysis, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{quicksight.ResourceStatusCreationInProgress},
		Target:                    []string{quicksight.ResourceStatusCreationSuccessful},
		Refresh:                   statusAnalysis(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.Analysis); ok {
		if status, apiErrors := aws.StringValue(out.Status), out.Errors; status == quicksight.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error

			for _, apiError := range apiErrors {
				if apiError == nil {
					continue
				}

				errs = append(errs, awserr.New(aws.StringValue(apiError.Type), aws.StringValue(apiError.Message), nil))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitAnalysisUpdated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.Analysis, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{quicksight.ResourceStatusUpdateInProgress, quicksight.ResourceStatusCreationInProgress},
		Target:                    []string{quicksight.ResourceStatusUpdateSuccessful, quicksight.ResourceStatusCreationSuccessful},
		Refresh:                   statusAnalysis(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.Analysis); ok {
		if status, apiErrors := aws.StringValue(out.Status), out.Errors; status == quicksight.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error

			for _, apiError := range apiErrors {
				if apiError == nil {
					continue
				}

				errs = append(errs, awserr.New(aws.StringValue(apiError.Type), aws.StringValue(apiError.Message), nil))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitThemeCreated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.Theme, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{quicksight.ResourceStatusCreationInProgress},
		Target:                    []string{quicksight.ResourceStatusCreationSuccessful},
		Refresh:                   statusTheme(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.Theme); ok {
		if status, apiErrors := aws.StringValue(out.Version.Status), out.Version.Errors; status == quicksight.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error

			for _, apiError := range apiErrors {
				if apiError == nil {
					continue
				}

				errs = append(errs, awserr.New(aws.StringValue(apiError.Type), aws.StringValue(apiError.Message), nil))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}

func waitThemeUpdated(ctx context.Context, conn *quicksight.QuickSight, id string, timeout time.Duration) (*quicksight.Theme, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{quicksight.ResourceStatusUpdateInProgress, quicksight.ResourceStatusCreationInProgress},
		Target:                    []string{quicksight.ResourceStatusUpdateSuccessful, quicksight.ResourceStatusCreationSuccessful},
		Refresh:                   statusTheme(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*quicksight.Theme); ok {
		if status, apiErrors := aws.StringValue(out.Version.Status), out.Version.Errors; status == quicksight.ResourceStatusCreationFailed && apiErrors != nil {
			var errs []error

			for _, apiError := range apiErrors {
				if apiError == nil {
					continue
				}

				errs = append(errs, awserr.New(aws.StringValue(apiError.Type), aws.StringValue(apiError.Message), nil))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return out, err
	}

	return nil, err
}
