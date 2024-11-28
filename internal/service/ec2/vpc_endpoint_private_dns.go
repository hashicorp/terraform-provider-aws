// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_endpoint_private_dns", name="VPC Endpoint Private DNS")
func newVPCEndpointPrivateDNSResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &vpcEndpointPrivateDNSResource{}, nil
}

type vpcEndpointPrivateDNSResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (*vpcEndpointPrivateDNSResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_vpc_endpoint_private_dns"
}

func (r *vpcEndpointPrivateDNSResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"private_dns_enabled": schema.BoolAttribute{
				Required: true,
			},
			names.AttrVPCEndpointID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *vpcEndpointPrivateDNSResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcEndpointPrivateDNSResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.ModifyVpcEndpointInput{
		PrivateDnsEnabled: fwflex.BoolFromFramework(ctx, data.PrivateDNSEnabled),
		VpcEndpointId:     fwflex.StringFromFramework(ctx, data.VPCEndpointID),
	}

	_, err := conn.ModifyVpcEndpoint(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating VPC Endpoint Private DNS (%s)", data.VPCEndpointID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *vpcEndpointPrivateDNSResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcEndpointPrivateDNSResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	vpce, err := findVPCEndpointByID(ctx, conn, data.VPCEndpointID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Endpoint (%s)", data.VPCEndpointID.ValueString()), err.Error())

		return
	}

	data.PrivateDNSEnabled = fwflex.BoolToFramework(ctx, vpce.PrivateDnsEnabled)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcEndpointPrivateDNSResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data vpcEndpointPrivateDNSResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.ModifyVpcEndpointInput{
		PrivateDnsEnabled: fwflex.BoolFromFramework(ctx, data.PrivateDNSEnabled),
		VpcEndpointId:     fwflex.StringFromFramework(ctx, data.VPCEndpointID),
	}

	_, err := conn.ModifyVpcEndpoint(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Updating VPC Endpoint Private DNS (%s)", data.VPCEndpointID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcEndpointPrivateDNSResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrVPCEndpointID), request, response)
}

type vpcEndpointPrivateDNSResourceModel struct {
	PrivateDNSEnabled types.Bool   `tfsdk:"private_dns_enabled"`
	VPCEndpointID     types.String `tfsdk:"vpc_endpoint_id"`
}
