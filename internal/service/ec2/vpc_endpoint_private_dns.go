// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_endpoint_private_dns", name="Endpoint Private DNS")
func newResourceEndpointPrivateDNS(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceEndpointPrivateDNS{}, nil
}

const (
	ResNameEndpointPrivateDNS = "Endpoint Private DNS"
)

type resourceEndpointPrivateDNS struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (r *resourceEndpointPrivateDNS) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_vpc_endpoint_private_dns"
}

func (r *resourceEndpointPrivateDNS) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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

func (r *resourceEndpointPrivateDNS) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceEndpointPrivateDNSData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:     aws.String(plan.VpcEndpointID.ValueString()),
		PrivateDnsEnabled: aws.Bool(plan.PrivateDNSEnabled.ValueBool()),
	}

	out, err := conn.ModifyVpcEndpoint(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameEndpointPrivateDNS, plan.VpcEndpointID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameEndpointPrivateDNS, plan.VpcEndpointID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEndpointPrivateDNS) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceEndpointPrivateDNSData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCEndpointByIDV2(ctx, conn, state.VpcEndpointID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, ResNameEndpointPrivateDNS, state.VpcEndpointID.String(), err),
			err.Error(),
		)
		return
	}

	state.PrivateDNSEnabled = flex.BoolToFramework(ctx, out.PrivateDnsEnabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEndpointPrivateDNS) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan, state resourceEndpointPrivateDNSData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.PrivateDNSEnabled.Equal(state.PrivateDNSEnabled) {
		in := &ec2.ModifyVpcEndpointInput{
			VpcEndpointId:     aws.String(plan.VpcEndpointID.ValueString()),
			PrivateDnsEnabled: aws.Bool(plan.PrivateDNSEnabled.ValueBool()),
		}

		out, err := conn.ModifyVpcEndpoint(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameEndpointPrivateDNS, plan.VpcEndpointID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameEndpointPrivateDNS, plan.VpcEndpointID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEndpointPrivateDNS) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrVPCEndpointID), req, resp)
}

type resourceEndpointPrivateDNSData struct {
	VpcEndpointID     types.String `tfsdk:"vpc_endpoint_id"`
	PrivateDNSEnabled types.Bool   `tfsdk:"private_dns_enabled"`
}
