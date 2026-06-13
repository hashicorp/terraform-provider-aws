// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_region", name="Region")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstances")
func newRegionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &regionResource{}

	r.SetDefaultCreateTimeout(120 * time.Minute)
	r.SetDefaultDeleteTimeout(120 * time.Minute)

	return r, nil
}

const (
	ResNameRegion = "Region"

	regionIDPartCount = 2
)

type regionResource struct {
	framework.ResourceWithModel[regionResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *regionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"added_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_primary_region": schema.BoolAttribute{
				Computed: true,
			},
			"region_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RegionStatus](),
				Computed:   true,
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

func (r *regionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan regionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := regionCreateResourceID(plan.InstanceARN.ValueString(), plan.RegionName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameRegion, plan.RegionName.ValueString(), err),
			err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(id)

	input := &ssoadmin.AddRegionInput{
		InstanceArn: plan.InstanceARN.ValueStringPointer(),
		RegionName:  plan.RegionName.ValueStringPointer(),
	}

	_, err = conn.AddRegion(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameRegion, plan.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	out, err := waitRegionActive(ctx, conn, plan.InstanceARN.ValueString(), plan.RegionName.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionWaitingForCreation, ResNameRegion, plan.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *regionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state regionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRegionByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionReading, ResNameRegion, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	parts, err := regionParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionReading, ResNameRegion, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.InstanceARN = fwtypes.ARNValue(parts[0])

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *regionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan regionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *regionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state regionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &ssoadmin.RemoveRegionInput{
		InstanceArn: state.InstanceARN.ValueStringPointer(),
		RegionName:  state.RegionName.ValueStringPointer(),
	}

	_, err := conn.RemoveRegion(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionDeleting, ResNameRegion, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	if _, err := waitRegionDeleted(ctx, conn, state.InstanceARN.ValueString(), state.RegionName.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionWaitingForDeletion, ResNameRegion, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func regionCreateResourceID(instanceARN, regionName string) (string, error) {
	return intflex.FlattenResourceId([]string{instanceARN, regionName}, regionIDPartCount, false)
}

func regionParseResourceID(id string) ([]string, error) {
	return intflex.ExpandResourceId(id, regionIDPartCount, false)
}

func findRegionByID(ctx context.Context, conn *ssoadmin.Client, id string) (*ssoadmin.DescribeRegionOutput, error) {
	parts, err := regionParseResourceID(id)
	if err != nil {
		return nil, err
	}

	return findRegionByTwoPartKey(ctx, conn, parts[0], parts[1])
}

func findRegionByTwoPartKey(ctx context.Context, conn *ssoadmin.Client, instanceARN, regionName string) (*ssoadmin.DescribeRegionOutput, error) {
	input := &ssoadmin.DescribeRegionInput{
		InstanceArn: aws.String(instanceARN),
		RegionName:  aws.String(regionName),
	}

	output, err := conn.DescribeRegion(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusRegion(conn *ssoadmin.Client, instanceARN, regionName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRegionByTwoPartKey(ctx, conn, instanceARN, regionName)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitRegionActive(ctx context.Context, conn *ssoadmin.Client, instanceARN, regionName string, timeout time.Duration) (*ssoadmin.DescribeRegionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegionStatusAdding),
		Target:  enum.Slice(awstypes.RegionStatusActive),
		Refresh: statusRegion(conn, instanceARN, regionName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssoadmin.DescribeRegionOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRegionDeleted(ctx context.Context, conn *ssoadmin.Client, instanceARN, regionName string, timeout time.Duration) (*ssoadmin.DescribeRegionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegionStatusActive, awstypes.RegionStatusRemoving),
		Target:  []string{},
		Refresh: statusRegion(conn, instanceARN, regionName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssoadmin.DescribeRegionOutput); ok {
		return output, err
	}

	return nil, err
}

type regionResourceModel struct {
	framework.WithRegionModel
	AddedDate       timetypes.RFC3339                         `tfsdk:"added_date"`
	ID              types.String                              `tfsdk:"id"`
	InstanceARN     fwtypes.ARN                               `tfsdk:"instance_arn"`
	IsPrimaryRegion types.Bool                                `tfsdk:"is_primary_region"`
	RegionName      types.String                              `tfsdk:"region_name"`
	Status          fwtypes.StringEnum[awstypes.RegionStatus] `tfsdk:"status"`
	Timeouts        timeouts.Value                            `tfsdk:"timeouts"`
}
