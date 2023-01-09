package rds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	sdkv2resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	_sp.registerFrameworkResourceFactory(newResourceExportTask)
}

func newResourceExportTask(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceExportTask{}, nil
}

const (
	ResNameExportTask = "ExportTask"
)

type resourceExportTask struct {
	framework.ResourceWithConfigure
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
			"iam_role_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"id": framework.IDAttribute(),
			"kms_key_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"percent_progress": schema.Int64Attribute{
				Computed: true,
			},
			"s3_bucket_name": schema.StringAttribute{
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
			"source_type": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
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
	}
}

func (r *resourceExportTask) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RDSClient()

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

	out, err := conn.StartExportTask(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameExportTask, plan.ExportTaskIdentifier.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameExportTask, plan.ExportTaskIdentifier.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	state.refreshFromStartOutput(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceExportTask) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RDSClient()

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

	state.refreshFromDescribeOutput(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// There is no update API, so this method is a no-op
func (r *resourceExportTask) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceExportTask) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RDSClient()

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
}

func (r *resourceExportTask) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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
		return nil, &sdkv2resource.NotFoundError{
			LastRequest: in,
		}
	}
	if len(out.ExportTasks) != 1 {
		return nil, tfresource.NewTooManyResultsError(len(out.ExportTasks), in)
	}

	return &out.ExportTasks[0], nil
}

type resourceExportTaskData struct {
	ExportOnly           types.List   `tfsdk:"export_only"`
	ExportTaskIdentifier types.String `tfsdk:"export_task_identifier"`
	FailureCause         types.String `tfsdk:"failure_cause"`
	IAMRoleArn           types.String `tfsdk:"iam_role_arn"`
	ID                   types.String `tfsdk:"id"`
	KMSKeyID             types.String `tfsdk:"kms_key_id"`
	PercentProgress      types.Int64  `tfsdk:"percent_progress"`
	S3BucketName         types.String `tfsdk:"s3_bucket_name"`
	S3Prefix             types.String `tfsdk:"s3_prefix"`
	SnapshotTime         types.String `tfsdk:"snapshot_time"`
	SourceArn            types.String `tfsdk:"source_arn"`
	SourceType           types.String `tfsdk:"source_type"`
	Status               types.String `tfsdk:"status"`
	TaskEndTime          types.String `tfsdk:"task_end_time"`
	TaskStartTime        types.String `tfsdk:"task_start_time"`
	WarningMessage       types.String `tfsdk:"warning_message"`
}

// refreshFromOutput writes state data from an AWS response object
//
// This variant of the refresh method is for use with the start operation
// response type (StartExportTaskOutput).
func (rd *resourceExportTaskData) refreshFromStartOutput(ctx context.Context, out *rds.StartExportTaskOutput) {
	if out == nil {
		return
	}

	rd.ID = flex.StringToFramework(ctx, out.ExportTaskIdentifier)
	rd.ExportOnly = flex.FlattenFrameworkStringValueList(ctx, out.ExportOnly)
	rd.ExportTaskIdentifier = flex.StringToFramework(ctx, out.ExportTaskIdentifier)
	rd.FailureCause = flex.StringToFramework(ctx, out.FailureCause)
	rd.IAMRoleArn = flex.StringToFramework(ctx, out.IamRoleArn)
	rd.KMSKeyID = flex.StringToFramework(ctx, out.KmsKeyId)
	rd.PercentProgress = types.Int64Value(int64(out.PercentProgress))
	rd.S3BucketName = flex.StringToFramework(ctx, out.S3Bucket)
	rd.S3Prefix = flex.StringToFramework(ctx, out.S3Prefix)
	rd.SnapshotTime = flex.StringValueToFramework(ctx, out.SnapshotTime.String())
	rd.SourceArn = flex.StringToFramework(ctx, out.SourceArn)
	rd.SourceType = flex.StringValueToFramework(ctx, out.SourceType)
	rd.Status = flex.StringToFramework(ctx, out.Status)
	rd.TaskEndTime = timeToFramework(ctx, out.TaskEndTime)
	rd.TaskStartTime = timeToFramework(ctx, out.TaskEndTime)
	rd.WarningMessage = flex.StringToFramework(ctx, out.WarningMessage)
}

// refreshFromOutput writes state data from an AWS response object
//
// This variant of the refresh method is for use with the describe operation
// response type (ExportTask).
func (rd *resourceExportTaskData) refreshFromDescribeOutput(ctx context.Context, out *awstypes.ExportTask) {
	if out == nil {
		return
	}

	rd.ID = flex.StringToFramework(ctx, out.ExportTaskIdentifier)
	rd.ExportOnly = flex.FlattenFrameworkStringValueList(ctx, out.ExportOnly)
	rd.ExportTaskIdentifier = flex.StringToFramework(ctx, out.ExportTaskIdentifier)
	rd.FailureCause = flex.StringToFramework(ctx, out.FailureCause)
	rd.IAMRoleArn = flex.StringToFramework(ctx, out.IamRoleArn)
	rd.KMSKeyID = flex.StringToFramework(ctx, out.KmsKeyId)
	rd.PercentProgress = types.Int64Value(int64(out.PercentProgress))
	rd.S3BucketName = flex.StringToFramework(ctx, out.S3Bucket)
	rd.S3Prefix = flex.StringToFramework(ctx, out.S3Prefix)
	rd.SnapshotTime = flex.StringValueToFramework(ctx, out.SnapshotTime.String())
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
