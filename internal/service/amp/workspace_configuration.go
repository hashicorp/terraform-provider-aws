// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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

// @FrameworkResource("aws_prometheus_workspace_configuration", name="WorkspaceConfiguration")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amp/types;types.WorkspaceConfigurationDescription")
func newWorkspaceConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &workspaceConfigurationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	return r, nil
}

type workspaceConfigurationResourceModel struct {
	ID                    types.String                                            `tfsdk:"id"`
	WorkspaceID           types.String                                            `tfsdk:"workspace_id"`
	RetentionPeriodInDays types.Int32                                             `tfsdk:"retention_period_in_days"`
	LimitsPerLabelSet     fwtypes.ListNestedObjectValueOf[limitsPerLabelSetModel] `tfsdk:"limits_per_label_set"`
	Timeouts              timeouts.Value                                          `tfsdk:"timeouts"`
}

type limitsPerLabelSetModel struct {
	LabelSet fwtypes.MapValueOf[types.String]                             `tfsdk:"label_set"`
	Limits   fwtypes.ListNestedObjectValueOf[limitsPerLabelSetEntryModel] `tfsdk:"limits"`
}

type limitsPerLabelSetEntryModel struct {
	MaxSeries types.Int64 `tfsdk:"max_series"`
}

type workspaceConfigurationResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoOpDelete
	framework.WithImportByID
}

func (r *workspaceConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"workspace_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"retention_period_in_days": schema.Int32Attribute{
				Optional: true,
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
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"max_series": schema.Int64Attribute{
										Required: true,
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

func (r *workspaceConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data workspaceConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	var input amp.UpdateWorkspaceConfigurationInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	workspaceID := data.WorkspaceID.ValueString()
	data.ID = types.StringValue(workspaceID)

	input.ClientToken = aws.String(sdkid.UniqueId())
	input.WorkspaceId = &workspaceID

	_, err := conn.UpdateWorkspaceConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("updating Workspace configuration", err.Error())
		return
	}

	if err := waitWorkspaceConfigurationUpdated(ctx, conn, workspaceID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for workspace configuration (%s) update", data.ID.ValueString()), err.Error())
		return
	}

	// Save the plan to state
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *workspaceConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data workspaceConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	out, err := findWorkspaceConfigurationByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("reading Workspace configuration", err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workspaceConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	workspaceID := req.ID

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), workspaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
}

func (r *workspaceConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data workspaceConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	var input amp.UpdateWorkspaceConfigurationInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	workspaceID := data.WorkspaceID.ValueString()
	data.ID = types.StringValue(workspaceID)

	input.ClientToken = aws.String(sdkid.UniqueId())
	input.WorkspaceId = &workspaceID

	_, err := conn.UpdateWorkspaceConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("updating Workspace configuration", err.Error())
		return
	}

	if err := waitWorkspaceConfigurationUpdated(ctx, conn, workspaceID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for workspace configuration (%s) update", data.ID.ValueString()), err.Error())
		return
	}

	// Save the plan to state
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func findWorkspaceConfigurationByID(ctx context.Context, conn *amp.Client, id string) (*amp.DescribeWorkspaceConfigurationOutput, error) {
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

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func waitWorkspaceConfigurationUpdated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkspaceConfigurationStatusCodeUpdating),
		Target:  enum.Slice(awstypes.WorkspaceConfigurationStatusCodeActive),
		Refresh: statusWorkspaceConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if _, ok := outputRaw.(*amp.DescribeWorkspaceConfigurationOutput); ok {
		return err
	}

	return nil
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

		return output, string(output.WorkspaceConfiguration.Status.StatusCode), nil
	}
}
