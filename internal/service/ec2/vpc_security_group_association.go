// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_security_group_association", name="Security Group VPC Association")
func newResourceVPCSecurityGroupAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCSecurityGroupAssociation{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameVPCSecurityGroupAssociation = "Security Group VPC Association"
)

type resourceVPCSecurityGroupAssociation struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[resourceVPCSecurityGroupAssociationModel]
	framework.WithTimeouts
}

func (r *resourceVPCSecurityGroupAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_vpc_security_group_association"
}

func (r *resourceVPCSecurityGroupAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrSecurityGroupID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceVPCSecurityGroupAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCSecurityGroupAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.AssociateSecurityGroupVpcInput

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.AssociateSecurityGroupVpc(ctx, &input)

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCSecurityGroupAssociation, plan.GroupId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameVPCSecurityGroupAssociation, plan.GroupId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err = waitVPCSecurityGroupAssociationCreated(ctx, conn, plan.GroupId.ValueString(), plan.VpcId.ValueString(), time.Minute*5)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameVPCSecurityGroupAssociation, plan.GroupId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceVPCSecurityGroupAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCSecurityGroupAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindVPCSecurityGroupAssociationByTwoPartKey(ctx, conn, state.GroupId.ValueString(), state.VpcId.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameVPCSecurityGroupAssociation, state.GroupId.String(), err),
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

func (r *resourceVPCSecurityGroupAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCSecurityGroupAssociationModel
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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameVPCSecurityGroupAssociation, state.GroupId.String(), err),
			err.Error(),
		)
		return
	}
	_, err = waitVPCSecurityGroupAssociationDeleted(ctx, conn, state.GroupId.ValueString(), state.VpcId.ValueString(), time.Minute*5)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameVPCSecurityGroupAssociation, state.GroupId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceVPCSecurityGroupAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ":")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: GroupId:VpcID. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrSecurityGroupID), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrVPCID), idParts[1])...)
}

func waitVPCSecurityGroupAssociationCreated(ctx context.Context, conn *ec2.Client, groupId string, vpcId string, timeout time.Duration) (*awstypes.SecurityGroupVpcAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.SecurityGroupVpcAssociationStateAssociating),
		Target:                    enum.Slice(awstypes.SecurityGroupVpcAssociationStateAssociated),
		Refresh:                   statusVPCSecurityGroupAssociation(ctx, conn, groupId, vpcId),
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

func waitVPCSecurityGroupAssociationDeleted(ctx context.Context, conn *ec2.Client, groupId string, vpcId string, timeout time.Duration) (*awstypes.SecurityGroupVpcAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SecurityGroupVpcAssociationStateDisassociating),
		Target:  []string{},
		Refresh: statusVPCSecurityGroupAssociation(ctx, conn, groupId, vpcId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SecurityGroupVpcAssociation); ok {
		return out, err
	}

	return nil, err
}

func statusVPCSecurityGroupAssociation(ctx context.Context, conn *ec2.Client, groupId string, vpcId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindVPCSecurityGroupAssociationByTwoPartKey(ctx, conn, groupId, vpcId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindVPCSecurityGroupAssociationByTwoPartKey(ctx context.Context, conn *ec2.Client, groupId string, vpcId string) (*awstypes.SecurityGroupVpcAssociation, error) {
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
	pages := ec2.NewDescribeSecurityGroupVpcAssociationsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

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

type resourceVPCSecurityGroupAssociationModel struct {
	GroupId types.String `tfsdk:"security_group_id"`
	VpcId   types.String `tfsdk:"vpc_id"`
}
