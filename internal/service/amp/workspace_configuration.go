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
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
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

// @FrameworkResource("aws_prometheus_workspace_configuration", name="WorkspaceConfiguration")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amp/types;types.WorkspaceConfigurationDescription")
func newWorkspaceConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &workspaceConfigurationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

	return r, nil
}

type workspaceConfigurationResource struct {
	framework.ResourceWithModel[workspaceConfigurationResourceModel]
	framework.WithTimeouts
	framework.WithNoOpDelete
}

func (r *workspaceConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"retention_period_in_days": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"limits_per_label_set": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[limitsPerLabelSetModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"label_set": schema.MapAttribute{
							CustomType:  fwtypes.MapOfStringType,
							ElementType: types.StringType,
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"limits": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[limitsPerLabelSetEntryModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"max_series": schema.Int64Attribute{
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
			}),
		},
	}
}

func (r *workspaceConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data workspaceConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	workspaceID := fwflex.StringValueFromFramework(ctx, data.WorkspaceID)
	var input amp.UpdateWorkspaceConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())

	_, err := conn.UpdateWorkspaceConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating AMP Workspace (%s) Configuration", workspaceID), err.Error())

		return
	}

	output, err := waitWorkspaceConfigurationUpdated(ctx, conn, workspaceID, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AMP Workspace (%s) Configuration create", workspaceID), err.Error())

		return
	}

	// Set values for unknowns after creation is complete.
	data.RetentionPeriodInDays = fwflex.Int32ToFramework(ctx, output.RetentionPeriodInDays)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *workspaceConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data workspaceConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	out, err := findWorkspaceConfigurationByID(ctx, conn, data.WorkspaceID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading AMP Workspace (%s) Configuration", data.WorkspaceID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *workspaceConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new workspaceConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	workspaceID := fwflex.StringValueFromFramework(ctx, new.WorkspaceID)
	var input amp.UpdateWorkspaceConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())

	_, err := conn.UpdateWorkspaceConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating AMP Workspace (%s) Configuration", workspaceID), err.Error())

		return
	}

	if _, err := waitWorkspaceConfigurationUpdated(ctx, conn, workspaceID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for AMP Workspace (%s) Configuration update", workspaceID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

func (r *workspaceConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("workspace_id"), request, response)
}

func findWorkspaceConfigurationByID(ctx context.Context, conn *amp.Client, id string) (*awstypes.WorkspaceConfigurationDescription, error) {
	input := amp.DescribeWorkspaceConfigurationInput{
		WorkspaceId: aws.String(id),
	}
	output, err := conn.DescribeWorkspaceConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.WorkspaceConfiguration == nil || output.WorkspaceConfiguration.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.WorkspaceConfiguration, nil
}

func statusWorkspaceConfiguration(ctx context.Context, conn *amp.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findWorkspaceConfigurationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func waitWorkspaceConfigurationUpdated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*awstypes.WorkspaceConfigurationDescription, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkspaceConfigurationStatusCodeUpdating),
		Target:  enum.Slice(awstypes.WorkspaceConfigurationStatusCodeActive),
		Refresh: statusWorkspaceConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.WorkspaceConfigurationDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))

		return output, err
	}

	return nil, err
}

type workspaceConfigurationResourceModel struct {
	framework.WithRegionModel
	LimitsPerLabelSet     fwtypes.ListNestedObjectValueOf[limitsPerLabelSetModel] `tfsdk:"limits_per_label_set"`
	RetentionPeriodInDays types.Int32                                             `tfsdk:"retention_period_in_days"`
	Timeouts              timeouts.Value                                          `tfsdk:"timeouts"`
	WorkspaceID           types.String                                            `tfsdk:"workspace_id"`
}

type limitsPerLabelSetModel struct {
	LabelSet fwtypes.MapOfString                                          `tfsdk:"label_set"`
	Limits   fwtypes.ListNestedObjectValueOf[limitsPerLabelSetEntryModel] `tfsdk:"limits"`
}

type limitsPerLabelSetEntryModel struct {
	MaxSeries types.Int64 `tfsdk:"max_series"`
}
