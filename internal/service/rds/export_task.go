// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rds_export_task")
func newResourceExportTask(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceExportTask{}
	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

const (
	ResNameExportTask = "ExportTask"

	// Use string constants as the RDS package does not provide status enums
	StatusCanceled   = "CANCELED"
	StatusCanceling  = "CANCELING"
	StatusComplete   = "COMPLETE"
	StatusFailed     = "FAILED"
	StatusInProgress = "IN_PROGRESS"
	StatusStarting   = "STARTING"
)

type resourceExportTask struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceExportTask) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_rds_export_task"
}

func (r *resourceExportTask) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"export_only": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"export_task_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"failure_cause": schema.StringAttribute{
				Computed: true,
			},
			names.AttrIAMRoleARN: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"percent_progress": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrS3BucketName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"s3_prefix": schema.StringAttribute{
				Optional: true,
				Computed: true, // This attribute can be returned by the Describe API even if unset
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"snapshot_time": schema.StringAttribute{
				Computed: true,
			},
			"source_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			names.AttrSourceType: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			"task_end_time": schema.StringAttribute{
				Computed: true,
			},
			"task_start_time": schema.StringAttribute{
				Computed: true,
			},
			"warning_message": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceExportTask) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RDSClient(ctx)

	var plan resourceExportTaskData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := rds.StartExportTaskInput{
		ExportTaskIdentifier: aws.String(plan.ExportTaskIdentifier.ValueString()),
		IamRoleArn:           aws.String(plan.IAMRoleArn.ValueString()),
		KmsKeyId:             aws.String(plan.KMSKeyID.ValueString()),
		S3BucketName:         aws.String(plan.S3BucketName.ValueString()),
		SourceArn:            aws.String(plan.SourceArn.ValueString()),
	}
	if !plan.ExportOnly.IsNull() {
		in.ExportOnly = flex.ExpandFrameworkStringValueList(ctx, plan.ExportOnly)
	}
	if !plan.S3Prefix.IsNull() && !plan.S3Prefix.IsUnknown() {
		in.S3Prefix = aws.String(plan.S3Prefix.ValueString())
	}

	outStart, err := conn.StartExportTask(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameExportTask, plan.ExportTaskIdentifier.String(), nil),
			err.Error(),
		)
		return
	}
	if outStart == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameExportTask, plan.ExportTaskIdentifier.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitExportTaskCreated(ctx, conn, plan.ExportTaskIdentifier.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameExportTask, plan.ExportTaskIdentifier.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	state.refreshFromOutput(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceExportTask) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RDSClient(ctx)

	var state resourceExportTaskData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindExportTaskByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.AddWarning(
			"AWS Resource Not Found During Refresh",
			fmt.Sprintf("Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original Error: %s", err.Error()),
		)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionReading, ResNameExportTask, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.refreshFromOutput(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// There is no update API, so this method is a no-op
func (r *resourceExportTask) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceExportTask) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RDSClient(ctx)

	var state resourceExportTaskData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Attempt to cancel the export task, but ignore errors where the task is in an invalid
	// state (ie. completed or failed) which can't be cancelled
	_, err := conn.CancelExportTask(ctx, &rds.CancelExportTaskInput{
		ExportTaskIdentifier: aws.String(state.ID.ValueString()),
	})
	if err != nil {
		var stateFault *awstypes.InvalidExportTaskStateFault
		if errors.As(err, &stateFault) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionDeleting, ResNameExportTask, state.ID.String(), nil),
			err.Error(),
		)
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitExportTaskDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionDeleting, ResNameExportTask, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceExportTask) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindExportTaskByID(ctx context.Context, conn *rds.Client, id string) (*awstypes.ExportTask, error) {
	in := &rds.DescribeExportTasksInput{
		ExportTaskIdentifier: aws.String(id),
	}
	out, err := conn.DescribeExportTasks(ctx, in)
	// This API won't return a NotFound error if the identifier can't be found, just
	// an empty result slice.
	if err != nil {
		return nil, err
	}
	if out == nil || len(out.ExportTasks) == 0 {
		return nil, &retry.NotFoundError{
			LastRequest: in,
		}
	}
	if len(out.ExportTasks) != 1 {
		return nil, tfresource.NewTooManyResultsError(len(out.ExportTasks), in)
	}

	return &out.ExportTasks[0], nil
}

func statusExportTask(ctx context.Context, conn *rds.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindExportTaskByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

func waitExportTaskCreated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*awstypes.ExportTask, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{StatusStarting, StatusInProgress},
		Target:     []string{StatusComplete, StatusFailed},
		Refresh:    statusExportTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ExportTask); ok {
		return out, err
	}

	return nil, err
}

func waitExportTaskDeleted(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*awstypes.ExportTask, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusStarting, StatusInProgress, StatusCanceling},
		Target:  []string{},
		Refresh: statusExportTask(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ExportTask); ok {
		return out, err
	}

	return nil, err
}

type resourceExportTaskData struct {
	ExportOnly           types.List     `tfsdk:"export_only"`
	ExportTaskIdentifier types.String   `tfsdk:"export_task_identifier"`
	FailureCause         types.String   `tfsdk:"failure_cause"`
	IAMRoleArn           types.String   `tfsdk:"iam_role_arn"`
	ID                   types.String   `tfsdk:"id"`
	KMSKeyID             types.String   `tfsdk:"kms_key_id"`
	PercentProgress      types.Int64    `tfsdk:"percent_progress"`
	S3BucketName         types.String   `tfsdk:"s3_bucket_name"`
	S3Prefix             types.String   `tfsdk:"s3_prefix"`
	SnapshotTime         types.String   `tfsdk:"snapshot_time"`
	SourceArn            types.String   `tfsdk:"source_arn"`
	SourceType           types.String   `tfsdk:"source_type"`
	Status               types.String   `tfsdk:"status"`
	TaskEndTime          types.String   `tfsdk:"task_end_time"`
	TaskStartTime        types.String   `tfsdk:"task_start_time"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
	WarningMessage       types.String   `tfsdk:"warning_message"`
}

// refreshFromOutput writes state data from an AWS response object
func (rd *resourceExportTaskData) refreshFromOutput(ctx context.Context, out *awstypes.ExportTask) {
	if out == nil {
		return
	}

	rd.ID = flex.StringToFramework(ctx, out.ExportTaskIdentifier)
	rd.ExportOnly = flex.FlattenFrameworkStringValueList(ctx, out.ExportOnly)
	rd.ExportTaskIdentifier = flex.StringToFramework(ctx, out.ExportTaskIdentifier)
	rd.FailureCause = flex.StringToFramework(ctx, out.FailureCause)
	rd.IAMRoleArn = flex.StringToFramework(ctx, out.IamRoleArn)
	rd.KMSKeyID = flex.StringToFramework(ctx, out.KmsKeyId)
	rd.PercentProgress = types.Int64Value(int64(aws.ToInt32(out.PercentProgress)))
	rd.S3BucketName = flex.StringToFramework(ctx, out.S3Bucket)
	rd.S3Prefix = flex.StringToFramework(ctx, out.S3Prefix)
	rd.SnapshotTime = timeToFramework(ctx, out.SnapshotTime)
	rd.SourceArn = flex.StringToFramework(ctx, out.SourceArn)
	rd.SourceType = flex.StringValueToFramework(ctx, out.SourceType)
	rd.Status = flex.StringToFramework(ctx, out.Status)
	rd.TaskEndTime = timeToFramework(ctx, out.TaskEndTime)
	rd.TaskStartTime = timeToFramework(ctx, out.TaskEndTime)
	rd.WarningMessage = flex.StringToFramework(ctx, out.WarningMessage)
}

func timeToFramework(ctx context.Context, t *time.Time) basetypes.StringValue {
	if t == nil {
		return types.StringNull()
	}
	return flex.StringValueToFramework(ctx, t.String())
}
