// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Resource Gateway")
// @Tags(identifierAttribute="arn")
func newResourceResourceGateway(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceResourceGateway{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameResourceGateway = "Resource Gateway"
)

type resourceResourceGateway struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceResourceGateway) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_vpclattice_resource_gateway"
}

func (r *resourceResourceGateway) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"ip_address_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *resourceResourceGateway) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().VPCLatticeClient(ctx)

	var plan resourceResourceGatewayData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &vpclattice.CreateResourceGatewayInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create uses attribute name `VpcIdentifier` instead of VpcId
	in.VpcIdentifier = aws.String(plan.VPCID.ValueString())
	in.ClientToken = aws.String(sdkid.UniqueId())
	in.Tags = getTagsIn(ctx)

	out, err := conn.CreateResourceGateway(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionCreating, ResNameResourceGateway, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.Arn)
	plan.ID = flex.StringToFramework(ctx, out.Id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	resourceGateway, err := waitResourceGatewayActive(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForCreation, ResNameResourceGateway, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, resourceGateway, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceResourceGateway) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceResourceGatewayData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	out, err := findResourceGatewayByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionSetting, ResNameResourceGateway, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceResourceGateway) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceResourceGatewayData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	// Only security group IDs can be updated
	if !plan.SecurityGroupIDs.Equal(state.SecurityGroupIDs) {

		in := &vpclattice.UpdateResourceGatewayInput{
			ResourceGatewayIdentifier: aws.String(plan.ID.ValueString()),
			SecurityGroupIds:          flex.ExpandFrameworkStringValueSet(ctx, plan.SecurityGroupIDs),
		}

		out, err := conn.UpdateResourceGateway(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VPCLattice, create.ErrActionUpdating, ResNameResourceGateway, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VPCLattice, create.ErrActionUpdating, ResNameResourceGateway, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.Arn)
		plan.ID = flex.StringToFramework(ctx, out.Id)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitResourceGatewayActive(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForUpdate, ResNameResourceGateway, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceResourceGateway) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().VPCLatticeClient(ctx)

	var state resourceResourceGatewayData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &vpclattice.DeleteResourceGatewayInput{
		ResourceGatewayIdentifier: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteResourceGateway(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionDeleting, ResNameResourceGateway, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitResourceGatewayDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameResourceGateway, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResourceGateway) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitResourceGatewayActive(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetResourceGatewayOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceGatewayStatusCreateInProgress, awstypes.ResourceGatewayStatusUpdateInProgress),
		Target:                    enum.Slice(awstypes.ResourceGatewayStatusActive),
		Refresh:                   statusResourceGateway(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetResourceGatewayOutput); ok {
		return out, err
	}

	return nil, err
}

func waitResourceGatewayDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetResourceGatewayOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceGatewayStatusDeleteInProgress),
		Target:  []string{},
		Refresh: statusResourceGateway(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetResourceGatewayOutput); ok {
		return out, err
	}

	return nil, err
}

func statusResourceGateway(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findResourceGatewayByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Status)), nil
	}
}

func findResourceGatewayByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetResourceGatewayOutput, error) {
	in := &vpclattice.GetResourceGatewayInput{
		ResourceGatewayIdentifier: aws.String(id),
	}

	out, err := conn.GetResourceGateway(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceResourceGatewayData struct {
	ARN              types.String   `tfsdk:"arn"`
	ID               types.String   `tfsdk:"id"`
	IPAddressType    types.String   `tfsdk:"ip_address_type"`
	Name             types.String   `tfsdk:"name"`
	SecurityGroupIDs types.Set      `tfsdk:"security_group_ids"`
	Status           types.String   `tfsdk:"status"`
	SubnetIDs        types.Set      `tfsdk:"subnet_ids"`
	Tags             tftags.Map     `tfsdk:"tags"`
	TagsAll          tftags.Map     `tfsdk:"tags_all"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
	VPCID            types.String   `tfsdk:"vpc_id"`
}
