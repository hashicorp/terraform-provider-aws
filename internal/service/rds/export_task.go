// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rds_export_task", name="Export Task")
func newExportTaskResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &exportTaskResource{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

const (
	// Use string constants as the RDS package does not provide status enums.
	exportTaskStatusCanceled   = "CANCELED"
	exportTaskStatusCanceling  = "CANCELING"
	exportTaskStatusComplete   = "COMPLETE"
	exportTaskStatusFailed     = "FAILED"
	exportTaskStatusInProgress = "IN_PROGRESS"
	exportTaskStatusStarting   = "STARTING"
)

type exportTaskResource struct {
	framework.ResourceWithModel[exportTaskResourceModel]
	framework.WithNoUpdate
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *exportTaskResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"export_only": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
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
				CustomType: fwtypes.ARNType,
				Required:   true,
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
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"source_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
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
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"task_start_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
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

func (r *exportTaskResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data exportTaskResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ExportTaskIdentifier)
	var input rds.StartExportTaskInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}
	input.S3BucketName = fwflex.StringFromFramework(ctx, data.S3Bucket)

	_, err := conn.StartExportTask(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("starting RDS Export Task (%s)", id), err.Error())

		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, id)

	exportTask, err := waitExportTaskCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Export Task (%s) create", id), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, exportTask, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *exportTaskResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data exportTaskResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	output, err := findExportTaskByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading RDS Export Task (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *exportTaskResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data exportTaskResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	// Attempt to cancel the export task, but ignore errors where the task is in an invalid
	// state (ie. completed or failed) which can't be cancelled.
	id := fwflex.StringValueFromFramework(ctx, data.ID)
	input := rds.CancelExportTaskInput{
		ExportTaskIdentifier: aws.String(id),
	}
	_, err := conn.CancelExportTask(ctx, &input)

	if errs.IsA[*awstypes.InvalidExportTaskStateFault](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("canceling RDS Export Task (%s)", id), err.Error())

		return
	}

	if _, err := waitExportTaskDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Export Task (%s) delete", id), err.Error())

		return
	}
}

func findExportTaskByID(ctx context.Context, conn *rds.Client, id string) (*awstypes.ExportTask, error) {
	input := rds.DescribeExportTasksInput{
		ExportTaskIdentifier: aws.String(id),
	}

	return findExportTask(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.ExportTask]())
}

func findExportTask(ctx context.Context, conn *rds.Client, input *rds.DescribeExportTasksInput, filter tfslices.Predicate[*awstypes.ExportTask]) (*awstypes.ExportTask, error) {
	output, err := findExportTasks(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findExportTasks(ctx context.Context, conn *rds.Client, input *rds.DescribeExportTasksInput, filter tfslices.Predicate[*awstypes.ExportTask]) ([]awstypes.ExportTask, error) {
	var output []awstypes.ExportTask

	pages := rds.NewDescribeExportTasksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ExportTasks {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusExportTask(conn *rds.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findExportTaskByID(ctx, conn, id)

		if retry.NotFound(err) {
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
		Pending:    []string{exportTaskStatusStarting, exportTaskStatusInProgress},
		Target:     []string{exportTaskStatusComplete, exportTaskStatusFailed},
		Refresh:    statusExportTask(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ExportTask); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.FailureCause)))

		return output, err
	}

	return nil, err
}

func waitExportTaskDeleted(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*awstypes.ExportTask, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{exportTaskStatusStarting, exportTaskStatusInProgress, exportTaskStatusCanceling},
		Target:  []string{},
		Refresh: statusExportTask(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ExportTask); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.FailureCause)))

		return output, err
	}

	return nil, err
}

type exportTaskResourceModel struct {
	framework.WithRegionModel
	ExportOnly           fwtypes.ListOfString `tfsdk:"export_only"`
	ExportTaskIdentifier types.String         `tfsdk:"export_task_identifier"`
	FailureCause         types.String         `tfsdk:"failure_cause"`
	IAMRoleArn           fwtypes.ARN          `tfsdk:"iam_role_arn"`
	ID                   types.String         `tfsdk:"id"`
	KMSKeyID             types.String         `tfsdk:"kms_key_id"`
	PercentProgress      types.Int64          `tfsdk:"percent_progress"`
	S3Bucket             types.String         `tfsdk:"s3_bucket_name"`
	S3Prefix             types.String         `tfsdk:"s3_prefix"`
	SnapshotTime         timetypes.RFC3339    `tfsdk:"snapshot_time"`
	SourceARN            fwtypes.ARN          `tfsdk:"source_arn"`
	SourceType           types.String         `tfsdk:"source_type"`
	Status               types.String         `tfsdk:"status"`
	TaskEndTime          timetypes.RFC3339    `tfsdk:"task_end_time"`
	TaskStartTime        timetypes.RFC3339    `tfsdk:"task_start_time"`
	Timeouts             timeouts.Value       `tfsdk:"timeouts"`
	WarningMessage       types.String         `tfsdk:"warning_message"`
}
