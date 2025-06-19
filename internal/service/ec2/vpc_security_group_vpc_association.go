// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_security_group_vpc_association", name="Security Group VPC Association")
func newSecurityGroupVPCAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &securityGroupVPCAssociationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameSecurityGroupVPCAssociation = "Security Group VPC Association"
	securityGroupVPCAssociationIDParts = 2
)

type securityGroupVPCAssociationResource struct {
	framework.ResourceWithModel[securityGroupVPCAssociationResourceModel]
	framework.WithNoUpdate
	framework.WithTimeouts
}

func (r *securityGroupVPCAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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

func (r *securityGroupVPCAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data securityGroupVPCAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	var input ec2.AssociateSecurityGroupVpcInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.AssociateSecurityGroupVpc(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Security Group (%s) VPC (%s) Association", data.GroupID.ValueString(), data.VPCID.ValueString()), err.Error())

		return
	}

	output, err := waitSecurityGroupVPCAssociationCreated(ctx, conn, data.GroupID.ValueString(), data.VPCID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Security Group (%s) VPC (%s) Association create", data.GroupID.ValueString(), data.VPCID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *securityGroupVPCAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data securityGroupVPCAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	output, err := findSecurityGroupVPCAssociationByTwoPartKey(ctx, conn, data.GroupID.ValueString(), data.VPCID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Group (%s) VPC (%s) Association", data.GroupID.ValueString(), data.VPCID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.State = fwtypes.StringEnumValue(output.State)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *securityGroupVPCAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data securityGroupVPCAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.DisassociateSecurityGroupVpcInput{
		GroupId: fwflex.StringFromFramework(ctx, data.GroupID),
		VpcId:   fwflex.StringFromFramework(ctx, data.VPCID),
	}
	_, err := conn.DisassociateSecurityGroupVpc(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound, errCodeInvalidVPCIDNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Security Group (%s) VPC (%s) Association", data.GroupID.ValueString(), data.VPCID.ValueString()), err.Error())

		return
	}

	if _, err := waitSecurityGroupVPCAssociationDeleted(ctx, conn, data.GroupID.ValueString(), data.VPCID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Security Group (%s) VPC (%s) Association delete", data.GroupID.ValueString(), data.VPCID.ValueString()), err.Error())

		return
	}
}

func (r *securityGroupVPCAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(request.ID, securityGroupVPCAssociationIDParts, false)
	if err != nil {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("unexpected format for ID (%[1]s), expected SECURITY-GROUP-ID%[2]sVPC-ID", request.ID, intflex.ResourceIdSeparator),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("security_group_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrVPCID), parts[1])...)
}

type securityGroupVPCAssociationResourceModel struct {
	framework.WithRegionModel
	GroupID  types.String                                                  `tfsdk:"security_group_id"`
	State    fwtypes.StringEnum[awstypes.SecurityGroupVpcAssociationState] `tfsdk:"state"`
	Timeouts timeouts.Value                                                `tfsdk:"timeouts"`
	VPCID    types.String                                                  `tfsdk:"vpc_id"`
}
