// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	dmssdk "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	pfvalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	replicationTaskAssessmentRunStatusCancelling        = "cancelling"
	replicationTaskAssessmentRunStatusDeleting          = "deleting"
	replicationTaskAssessmentRunStatusErrorExecuting    = "error-executing"
	replicationTaskAssessmentRunStatusErrorProvisioning = "error-provisioning"
	replicationTaskAssessmentRunStatusFailed            = "failed"
	replicationTaskAssessmentRunStatusInvalidState      = "invalid state"
	replicationTaskAssessmentRunStatusPassed            = "passed"
	replicationTaskAssessmentRunStatusProvisioning      = "provisioning"
	replicationTaskAssessmentRunStatusRunning           = "running"
	replicationTaskAssessmentRunStatusStarting          = "starting"
	replicationTaskAssessmentRunStatusWarning           = "warning"

	defaultReplicationTaskAssessmentRunTimeout = 30 * time.Minute
	assessmentRunPollInterval                  = 15 * time.Second
	assessmentRunProgressInterval              = 30 * time.Second
)

// @Action(aws_dms_start_replication_task_assessment_run, name="Start Replication Task Assessment Run")
func newStartReplicationTaskAssessmentRunAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &startReplicationTaskAssessmentRunAction{}, nil
}

var (
	_ action.Action = (*startReplicationTaskAssessmentRunAction)(nil)
)

type startReplicationTaskAssessmentRunAction struct {
	framework.ActionWithModel[startReplicationTaskAssessmentRunActionModel]
}

type startReplicationTaskAssessmentRunActionModel struct {
	framework.WithRegionModel
	AssessmentRunName    types.String         `tfsdk:"assessment_run_name"`
	ReplicationTaskARN   types.String         `tfsdk:"replication_task_arn"`
	ResultLocationBucket types.String         `tfsdk:"result_location_bucket"`
	ServiceAccessRoleARN types.String         `tfsdk:"service_access_role_arn"`
	IncludeOnly          fwtypes.ListOfString `tfsdk:"include_only"`
	Exclude              fwtypes.ListOfString `tfsdk:"exclude"`
	ResultEncryptionMode types.String         `tfsdk:"result_encryption_mode"`
	ResultKMSKeyARN      types.String         `tfsdk:"result_kms_key_arn"`
	ResultLocationFolder types.String         `tfsdk:"result_location_folder"`
	Timeout              types.Int64          `tfsdk:"timeout"`
}

func (a *startReplicationTaskAssessmentRunAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Starts an AWS DMS premigration assessment run for a replication task and waits for the assessment run to reach a terminal state.",
		Attributes: map[string]schema.Attribute{
			"assessment_run_name": schema.StringAttribute{
				Description: "Unique name for the assessment run.",
				Required:    true,
			},
			"replication_task_arn": schema.StringAttribute{
				Description: "ARN of the DMS replication task to assess.",
				Required:    true,
				Validators: []pfvalidator.String{
					fwvalidators.ARN(),
				},
			},
			"result_location_bucket": schema.StringAttribute{
				Description: "Amazon S3 bucket where DMS stores the assessment results.",
				Required:    true,
			},
			"service_access_role_arn": schema.StringAttribute{
				Description: "ARN of the IAM role used by DMS to write assessment results and start the run.",
				Required:    true,
				Validators: []pfvalidator.String{
					fwvalidators.ARN(),
				},
			},
			"include_only": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Description: "Specific individual assessments to include in the run. Cannot be set with exclude.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"exclude": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Description: "Specific individual assessments to exclude from the run. Cannot be set with include_only.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"result_encryption_mode": schema.StringAttribute{
				Description: fmt.Sprintf("Encryption mode for assessment results. Valid values are %q and %q.", encryptionModeSseKMS, encryptionModeSseS3),
				Optional:    true,
			},
			"result_kms_key_arn": schema.StringAttribute{
				Description: "ARN of the KMS key used when result_encryption_mode is SSE_KMS.",
				Optional:    true,
				Validators: []pfvalidator.String{
					fwvalidators.ARN(),
				},
			},
			"result_location_folder": schema.StringAttribute{
				Description: "Folder within the S3 bucket where DMS stores the assessment results.",
				Optional:    true,
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for the assessment run to complete.",
				Optional:    true,
				Validators: []pfvalidator.Int64{
					int64validator.AtLeast(60),
				},
			},
		},
	}
}

func (a *startReplicationTaskAssessmentRunAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config startReplicationTaskAssessmentRunActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.IncludeOnly.IsNull() && !config.Exclude.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting Assessment Selection",
			"Only one of include_only or exclude can be configured for a DMS replication task assessment run.",
		)
		return
	}

	if !config.ResultKMSKeyARN.IsNull() {
		mode := fwflex.StringValueOr(ctx, config.ResultEncryptionMode, "")
		if mode != encryptionModeSseKMS {
			resp.Diagnostics.AddError(
				"Invalid KMS Encryption Configuration",
				"result_kms_key_arn can only be specified when result_encryption_mode is SSE_KMS.",
			)
			return
		}
	}

	if !config.ResultEncryptionMode.IsNull() {
		mode := config.ResultEncryptionMode.ValueString()
		if mode != encryptionModeSseKMS && mode != encryptionModeSseS3 {
			resp.Diagnostics.AddError(
				"Invalid Encryption Mode",
				fmt.Sprintf("result_encryption_mode must be one of %q or %q.", encryptionModeSseKMS, encryptionModeSseS3),
			)
			return
		}
	}

	conn := a.Meta().DMSClient(ctx)
	assessmentRunName := fwflex.StringValueFromFramework(ctx, config.AssessmentRunName)
	replicationTaskARN := fwflex.StringValueFromFramework(ctx, config.ReplicationTaskARN)
	timeout := fwactions.TimeoutOr(config.Timeout, defaultReplicationTaskAssessmentRunTimeout)

	tflog.Info(ctx, "Starting DMS replication task assessment run", map[string]any{
		"assessment_run_name":  assessmentRunName,
		"replication_task_arn": replicationTaskARN,
		"timeout":              timeout.String(),
	})

	cb := fwactions.NewSendProgressFunc(resp)
	cb(ctx, "Starting DMS assessment run %s...", assessmentRunName)

	input := &dmssdk.StartReplicationTaskAssessmentRunInput{
		AssessmentRunName:    aws.String(assessmentRunName),
		ReplicationTaskArn:   aws.String(replicationTaskARN),
		ResultLocationBucket: aws.String(fwflex.StringValueFromFramework(ctx, config.ResultLocationBucket)),
		ServiceAccessRoleArn: aws.String(fwflex.StringValueFromFramework(ctx, config.ServiceAccessRoleARN)),
	}

	if !config.IncludeOnly.IsNull() && !config.IncludeOnly.IsUnknown() {
		input.IncludeOnly = fwflex.ExpandFrameworkStringValueList(ctx, config.IncludeOnly)
	}

	if !config.Exclude.IsNull() && !config.Exclude.IsUnknown() {
		input.Exclude = fwflex.ExpandFrameworkStringValueList(ctx, config.Exclude)
	}

	if !config.ResultEncryptionMode.IsNull() {
		input.ResultEncryptionMode = config.ResultEncryptionMode.ValueStringPointer()
	}

	if !config.ResultKMSKeyARN.IsNull() {
		input.ResultKmsKeyArn = config.ResultKMSKeyARN.ValueStringPointer()
	}

	if !config.ResultLocationFolder.IsNull() {
		input.ResultLocationFolder = config.ResultLocationFolder.ValueStringPointer()
	}

	output, err := conn.StartReplicationTaskAssessmentRun(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Start DMS Assessment Run",
			fmt.Sprintf("Could not start DMS assessment run %s for replication task %s: %s", assessmentRunName, replicationTaskARN, err),
		)
		return
	}

	if output.ReplicationTaskAssessmentRun == nil || output.ReplicationTaskAssessmentRun.ReplicationTaskAssessmentRunArn == nil {
		resp.Diagnostics.AddError(
			"Missing Assessment Run ARN",
			fmt.Sprintf("DMS started assessment run %s but did not return an assessment run ARN.", assessmentRunName),
		)
		return
	}

	assessmentRunARN := aws.ToString(output.ReplicationTaskAssessmentRun.ReplicationTaskAssessmentRunArn)
	cb(ctx, "Assessment run %s started, waiting for completion...", assessmentRunARN)

	fr, err := actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[*awstypes.ReplicationTaskAssessmentRun], error) {
		run, err := findReplicationTaskAssessmentRunByARN(ctx, conn, assessmentRunARN)
		if retry.NotFound(err) {
			return actionwait.FetchResult[*awstypes.ReplicationTaskAssessmentRun]{
				Status: actionwait.Status(replicationTaskAssessmentRunStatusStarting),
			}, nil
		}
		if err != nil {
			return actionwait.FetchResult[*awstypes.ReplicationTaskAssessmentRun]{}, fmt.Errorf("describing assessment run: %w", err)
		}

		return actionwait.FetchResult[*awstypes.ReplicationTaskAssessmentRun]{
			Status: actionwait.Status(aws.ToString(run.Status)),
			Value:  run,
		}, nil
	}, actionwait.Options[*awstypes.ReplicationTaskAssessmentRun]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(assessmentRunPollInterval),
		ProgressInterval: assessmentRunProgressInterval,
		SuccessStates: []actionwait.Status{
			actionwait.Status(replicationTaskAssessmentRunStatusPassed),
			actionwait.Status(replicationTaskAssessmentRunStatusWarning),
		},
		TransitionalStates: []actionwait.Status{
			actionwait.Status(replicationTaskAssessmentRunStatusStarting),
			actionwait.Status(replicationTaskAssessmentRunStatusProvisioning),
			actionwait.Status(replicationTaskAssessmentRunStatusRunning),
		},
		FailureStates: []actionwait.Status{
			actionwait.Status(replicationTaskAssessmentRunStatusCancelling),
			actionwait.Status(replicationTaskAssessmentRunStatusDeleting),
			actionwait.Status(replicationTaskAssessmentRunStatusFailed),
			actionwait.Status(replicationTaskAssessmentRunStatusErrorProvisioning),
			actionwait.Status(replicationTaskAssessmentRunStatusErrorExecuting),
			actionwait.Status(replicationTaskAssessmentRunStatusInvalidState),
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			cb(ctx, "Assessment run %s is currently %s", assessmentRunName, fr.Status)
		},
	})
	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var failureErr *actionwait.FailureStateError
		var unexpectedErr *actionwait.UnexpectedStateError

		lastFailureMessage := ""
		if fr.Value != nil {
			lastFailureMessage = aws.ToString(fr.Value.LastFailureMessage)
		}

		suffix := ""
		if lastFailureMessage != "" {
			suffix = fmt.Sprintf(" Last failure message: %s", lastFailureMessage)
		}

		switch {
		case errors.As(err, &timeoutErr):
			resp.Diagnostics.AddError(
				"Timeout Waiting for DMS Assessment Run",
				fmt.Sprintf("DMS assessment run %s did not complete within %s.%s", assessmentRunName, timeout, suffix),
			)
		case errors.As(err, &failureErr):
			resp.Diagnostics.AddError(
				"DMS Assessment Run Failed",
				fmt.Sprintf("DMS assessment run %s reached failure status %s.%s", assessmentRunName, failureErr.Status, suffix),
			)
		case errors.As(err, &unexpectedErr):
			resp.Diagnostics.AddError(
				"Unexpected DMS Assessment Run Status",
				fmt.Sprintf("DMS assessment run %s entered unexpected status %s.%s", assessmentRunName, unexpectedErr.Status, suffix),
			)
		default:
			resp.Diagnostics.AddError(
				"Error Waiting for DMS Assessment Run",
				fmt.Sprintf("Error while waiting for DMS assessment run %s: %s.%s", assessmentRunName, err, suffix),
			)
		}
		return
	}

	cb(ctx, "Assessment run %s completed with status %s", assessmentRunName, fr.Status)

	logFields := map[string]any{
		"assessment_run_arn":  assessmentRunARN,
		"assessment_run_name": assessmentRunName,
		"status":              fr.Status,
	}
	if fr.Value != nil && fr.Value.AssessmentProgress != nil {
		logFields["individual_assessment_completed_count"] = fr.Value.AssessmentProgress.IndividualAssessmentCompletedCount
		logFields["individual_assessment_count"] = fr.Value.AssessmentProgress.IndividualAssessmentCount
	}

	tflog.Info(ctx, "DMS replication task assessment run completed", logFields)
}

func findReplicationTaskAssessmentRunByARN(ctx context.Context, conn *dmssdk.Client, arn string) (*awstypes.ReplicationTaskAssessmentRun, error) {
	input := &dmssdk.DescribeReplicationTaskAssessmentRunsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("replication-task-assessment-run-arn"),
				Values: []string{arn},
			},
		},
	}

	return findReplicationTaskAssessmentRun(ctx, conn, input)
}

func findReplicationTaskAssessmentRun(ctx context.Context, conn *dmssdk.Client, input *dmssdk.DescribeReplicationTaskAssessmentRunsInput) (*awstypes.ReplicationTaskAssessmentRun, error) {
	output, err := findReplicationTaskAssessmentRuns(ctx, conn, input)
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReplicationTaskAssessmentRuns(ctx context.Context, conn *dmssdk.Client, input *dmssdk.DescribeReplicationTaskAssessmentRunsInput) ([]awstypes.ReplicationTaskAssessmentRun, error) {
	var output []awstypes.ReplicationTaskAssessmentRun

	pages := dmssdk.NewDescribeReplicationTaskAssessmentRunsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return nil, &retry.NotFoundError{LastError: err}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ReplicationTaskAssessmentRuns...)
	}

	return output, nil
}
