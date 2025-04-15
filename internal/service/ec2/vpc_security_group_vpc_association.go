// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_security_group_vpc_association", name="Security Group VPC Association")
func newResourceSecurityGroupVPCAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSecurityGroupVPCAssociation{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameSecurityGroupVPCAssociation = "Security Group VPC Association"
	securityGroupVPCAssociationIDParts = 2
)

type resourceSecurityGroupVPCAssociation struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithTimeouts
}

func (r *resourceSecurityGroupVPCAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"security_group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SecurityGroupVpcAssociationState](),
				Computed:   true,
			},
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
				Delete: true,
			}),
		},
	}
}

func (r *resourceSecurityGroupVPCAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceSecurityGroupVPCAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.AssociateSecurityGroupVpcInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.AssociateSecurityGroupVpc(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameSecurityGroupVPCAssociation, plan.GroupId.String(), err),
			err.Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := waitSecurityGroupVPCAssociationCreated(ctx, conn, plan.GroupId.ValueString(), plan.VpcId.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameSecurityGroupVPCAssociation, plan.GroupId.String(), err),
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

func (r *resourceSecurityGroupVPCAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceSecurityGroupVPCAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindSecurityGroupVPCAssociationByTwoPartKey(ctx, conn, state.GroupId.ValueString(), state.VpcId.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameSecurityGroupVPCAssociation, state.GroupId.String(), err),
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

func (r *resourceSecurityGroupVPCAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceSecurityGroupVPCAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DisassociateSecurityGroupVpcInput{
		GroupId: state.GroupId.ValueStringPointer(),
		VpcId:   state.VpcId.ValueStringPointer(),
	}

	_, err := conn.DisassociateSecurityGroupVpc(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound) {
			return
		}
		if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameSecurityGroupVPCAssociation, state.GroupId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitSecurityGroupVPCAssociationDeleted(ctx, conn, state.GroupId.ValueString(), state.VpcId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameSecurityGroupVPCAssociation, state.GroupId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceSecurityGroupVPCAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, securityGroupVPCAssociationIDParts, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: security_group_id,vpc_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("security_group_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrVPCID), parts[1])...)
}

func waitSecurityGroupVPCAssociationCreated(ctx context.Context, conn *ec2.Client, groupId string, vpcId string, timeout time.Duration) (*awstypes.SecurityGroupVpcAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.SecurityGroupVpcAssociationStateAssociating),
		Target:                    enum.Slice(awstypes.SecurityGroupVpcAssociationStateAssociated),
		Refresh:                   statusSecurityGroupVPCAssociation(ctx, conn, groupId, vpcId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SecurityGroupVpcAssociation); ok {
		return out, err
	}

	return nil, err
}

func waitSecurityGroupVPCAssociationDeleted(ctx context.Context, conn *ec2.Client, groupId string, vpcId string, timeout time.Duration) (*awstypes.SecurityGroupVpcAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SecurityGroupVpcAssociationStateDisassociating),
		Target:  []string{},
		Refresh: statusSecurityGroupVPCAssociation(ctx, conn, groupId, vpcId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SecurityGroupVpcAssociation); ok {
		return out, err
	}

	return nil, err
}

func statusSecurityGroupVPCAssociation(ctx context.Context, conn *ec2.Client, groupId string, vpcId string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindSecurityGroupVPCAssociationByTwoPartKey(ctx, conn, groupId, vpcId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindSecurityGroupVPCAssociationByTwoPartKey(ctx context.Context, conn *ec2.Client, groupId string, vpcId string) (*awstypes.SecurityGroupVpcAssociation, error) {
	in := &ec2.DescribeSecurityGroupVpcAssociationsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("group-id"),
				Values: []string{groupId},
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	}

	paginator := ec2.NewDescribeSecurityGroupVpcAssociationsPaginator(conn, in)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("reading VPC Security Group Associations: %w", err)
		}

		for _, association := range page.SecurityGroupVpcAssociations {
			if association.GroupId == nil {
				continue
			}

			if association.GroupId != nil && association.VpcId != nil {
				return &association, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(in)
}

type resourceSecurityGroupVPCAssociationModel struct {
	GroupId  types.String                                                  `tfsdk:"security_group_id"`
	State    fwtypes.StringEnum[awstypes.SecurityGroupVpcAssociationState] `tfsdk:"state"`
	VpcId    types.String                                                  `tfsdk:"vpc_id"`
	Timeouts timeouts.Value                                                `tfsdk:"timeouts"`
}
