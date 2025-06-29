// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_prometheus_query_logging_configuration", name="QueryLoggingConfiguration")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amp/types;types.QueryLoggingConfigurationMetadata")
func newQueryLoggingConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &queryLoggingConfigurationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type queryLoggingConfigurationResource struct {
	framework.ResourceWithModel[queryLoggingConfigurationResourceModel]
	framework.WithTimeouts
}

func (r *queryLoggingConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"workspace_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"destinations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[queryLoggingDestinationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cloudwatch_logs": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudwatchLogsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
						"filters": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[loggingFilterModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"qsp_threshold": schema.Int64Attribute{
										Required: true,
										Validators: []validator.Int64{
											int64validator.AtLeast(0),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *queryLoggingConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data queryLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	workspaceID := fwflex.StringValueFromFramework(ctx, data.WorkspaceID)
	var input amp.CreateQueryLoggingConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())

	_, err := conn.CreateQueryLoggingConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating AMP Query Logging Configuration (%s)", workspaceID), err.Error())

		return
	}

	// Set the ID to the workspace ID as per the documentation
	data.ID = data.WorkspaceID

	output, err := waitQueryLoggingConfigurationCreated(ctx, conn, workspaceID, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AMP Query Logging Configuration (%s) create", workspaceID), err.Error())

		return
	}

	// Update data with any computed values from the API response
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *queryLoggingConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data queryLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	out, err := findQueryLoggingConfigurationByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading AMP Query Logging Configuration (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *queryLoggingConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new queryLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	workspaceID := fwflex.StringValueFromFramework(ctx, new.WorkspaceID)
	var input amp.UpdateQueryLoggingConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())

	_, err := conn.UpdateQueryLoggingConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating AMP Query Logging Configuration (%s)", workspaceID), err.Error())

		return
	}

	if _, err := waitQueryLoggingConfigurationUpdated(ctx, conn, workspaceID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AMP Query Logging Configuration (%s) update", workspaceID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

func (r *queryLoggingConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data queryLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	workspaceID := fwflex.StringValueFromFramework(ctx, data.WorkspaceID)

	_, err := conn.DeleteQueryLoggingConfiguration(ctx, &amp.DeleteQueryLoggingConfigurationInput{
		WorkspaceId: aws.String(workspaceID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting AMP Query Logging Configuration (%s)", workspaceID), err.Error())

		return
	}

	if _, err := waitQueryLoggingConfigurationDeleted(ctx, conn, workspaceID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AMP Query Logging Configuration (%s) delete", workspaceID), err.Error())

		return
	}
}

func (r *queryLoggingConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("workspace_id"), request, response)
	// Also set the ID to the workspace ID
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), request.ID)...)
}

func findQueryLoggingConfigurationByID(ctx context.Context, conn *amp.Client, id string) (*awstypes.QueryLoggingConfigurationMetadata, error) {
	input := amp.DescribeQueryLoggingConfigurationInput{
		WorkspaceId: aws.String(id),
	}
	output, err := conn.DescribeQueryLoggingConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.QueryLoggingConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.QueryLoggingConfiguration, nil
}

func statusQueryLoggingConfiguration(ctx context.Context, conn *amp.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findQueryLoggingConfigurationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func waitQueryLoggingConfigurationCreated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.QueryLoggingConfigurationMetadata, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.QueryLoggingConfigurationStatusCodeCreating),
		Target:  enum.Slice(awstypes.QueryLoggingConfigurationStatusCodeActive),
		Refresh: statusQueryLoggingConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.QueryLoggingConfigurationMetadata); ok {
		if output.Status != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitQueryLoggingConfigurationUpdated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.QueryLoggingConfigurationMetadata, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.QueryLoggingConfigurationStatusCodeUpdating),
		Target:  enum.Slice(awstypes.QueryLoggingConfigurationStatusCodeActive),
		Refresh: statusQueryLoggingConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.QueryLoggingConfigurationMetadata); ok {
		if output.Status != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitQueryLoggingConfigurationDeleted(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.QueryLoggingConfigurationMetadata, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.QueryLoggingConfigurationStatusCodeDeleting, awstypes.QueryLoggingConfigurationStatusCodeActive),
		Target:  []string{},
		Refresh: statusQueryLoggingConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.QueryLoggingConfigurationMetadata); ok {
		if output.Status != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

type queryLoggingConfigurationResourceModel struct {
	framework.WithRegionModel
	Destinations fwtypes.ListNestedObjectValueOf[queryLoggingDestinationModel] `tfsdk:"destinations"`
	ID           types.String                                                  `tfsdk:"id"`
	Timeouts     timeouts.Value                                                `tfsdk:"timeouts"`
	WorkspaceID  types.String                                                  `tfsdk:"workspace_id"`
}

type queryLoggingDestinationModel struct {
	CloudwatchLogs fwtypes.ListNestedObjectValueOf[cloudwatchLogsModel] `tfsdk:"cloudwatch_logs"`
	Filters        fwtypes.ListNestedObjectValueOf[loggingFilterModel]  `tfsdk:"filters"`
}

type cloudwatchLogsModel struct {
	LogGroupArn fwtypes.ARN `tfsdk:"log_group_arn"`
}

type loggingFilterModel struct {
	QspThreshold types.Int64 `tfsdk:"qsp_threshold"`
}
