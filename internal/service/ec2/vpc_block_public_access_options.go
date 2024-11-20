// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_block_public_access_options", name="VPC Block Public Access Options")
func newResourceVPCBlockPublicAccessOptions(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCBlockPublicAccessOptions{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCBlockPublicAccessOptions = "VPC Block Public Access Options"
)

type resourceVPCBlockPublicAccessOptions struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceVPCBlockPublicAccessOptions) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_vpc_block_public_access_options"
}

func (r *resourceVPCBlockPublicAccessOptions) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"aws_account_id": schema.StringAttribute{
				Computed: true,
			},
			"aws_region": schema.StringAttribute{
				Computed: true,
			},
			"internet_gateway_block_mode": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.InternetGatewayBlockMode](),
				},
			},
			"last_update_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"reason": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceVPCBlockPublicAccessOptions) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCBlockPublicAccessOptionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.ModifyVpcBlockPublicAccessOptionsInput

	input.InternetGatewayBlockMode = awstypes.InternetGatewayBlockMode(plan.InternetGatewayBlockMode.ValueString())

	out, err := conn.ModifyVpcBlockPublicAccessOptions(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCBlockPublicAccessOptions, "ModifyVpcBlockPublicAccessOptions", err),
			err.Error(),
		)
		return
	}

	if out == nil || out.VpcBlockPublicAccessOptions == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCBlockPublicAccessOptions, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.VpcBlockPublicAccessOptions, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID+":"+r.Meta().Region)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitVPCBlockPublicAccessOptionsUpdated(ctx, conn, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCBlockPublicAccessOptions, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	desc_out, desc_err := conn.DescribeVpcBlockPublicAccessOptions(ctx, &ec2.DescribeVpcBlockPublicAccessOptionsInput{})
	if desc_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameVPCBlockPublicAccessOptions, plan.ID.String(), desc_err),
			desc_err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, desc_out.VpcBlockPublicAccessOptions, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVPCBlockPublicAccessOptions) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCBlockPublicAccessOptionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.DescribeVpcBlockPublicAccessOptions(ctx, &ec2.DescribeVpcBlockPublicAccessOptionsInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameVPCBlockPublicAccessOptions, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.VpcBlockPublicAccessOptions, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceVPCBlockPublicAccessOptions) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan, state resourceVPCBlockPublicAccessOptionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.InternetGatewayBlockMode.Equal(state.InternetGatewayBlockMode) {
		var input ec2.ModifyVpcBlockPublicAccessOptionsInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test"))...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.ModifyVpcBlockPublicAccessOptions(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCBlockPublicAccessOptions, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.VpcBlockPublicAccessOptions == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCBlockPublicAccessOptions, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out.VpcBlockPublicAccessOptions, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitVPCBlockPublicAccessOptionsUpdated(ctx, conn, updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCBlockPublicAccessOptions, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	desc_out, desc_err := conn.DescribeVpcBlockPublicAccessOptions(ctx, &ec2.DescribeVpcBlockPublicAccessOptionsInput{})
	if desc_err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameVPCBlockPublicAccessOptions, plan.ID.String(), desc_err),
			desc_err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, desc_out.VpcBlockPublicAccessOptions, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceVPCBlockPublicAccessOptions) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCBlockPublicAccessOptionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.ModifyVpcBlockPublicAccessOptionsInput

	// On deletion of this resource set the VPC Block Public Access Options to off
	input.InternetGatewayBlockMode = awstypes.InternetGatewayBlockModeOff

	out, err := conn.ModifyVpcBlockPublicAccessOptions(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCBlockPublicAccessOptions, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.VpcBlockPublicAccessOptions == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCBlockPublicAccessOptions, state.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitVPCBlockPublicAccessOptionsUpdated(ctx, conn, deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameVPCBlockPublicAccessOptions, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCBlockPublicAccessOptions) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitVPCBlockPublicAccessOptionsUpdated(ctx context.Context, conn *ec2.Client, timeout time.Duration) (*awstypes.VpcBlockPublicAccessOptions, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.VpcBlockPublicAccessStateUpdateInProgress)},
		Target:                    []string{string(awstypes.VpcBlockPublicAccessStateUpdateComplete)},
		Refresh:                   statusVPCBlockPublicAccessOptions(ctx, conn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.VpcBlockPublicAccessOptions); ok {
		return out, err
	}

	return nil, err
}

func statusVPCBlockPublicAccessOptions(ctx context.Context, conn *ec2.Client) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.DescribeVpcBlockPublicAccessOptions(ctx, &ec2.DescribeVpcBlockPublicAccessOptionsInput{})
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.VpcBlockPublicAccessOptions.State), nil
	}
}

type resourceVPCBlockPublicAccessOptionsModel struct {
	AWSAccountID             types.String      `tfsdk:"aws_account_id"`
	AWSRegion                types.String      `tfsdk:"aws_region"`
	InternetGatewayBlockMode types.String      `tfsdk:"internet_gateway_block_mode"`
	ID                       types.String      `tfsdk:"id"`
	LastUpdateTimestamp      timetypes.RFC3339 `tfsdk:"last_update_timestamp"`
	Reason                   types.String      `tfsdk:"reason"`
	Timeouts                 timeouts.Value    `tfsdk:"timeouts"`
}
