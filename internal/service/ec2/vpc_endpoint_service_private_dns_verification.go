// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_endpoint_service_private_dns_verification", name="VPC Endpoint Service Private DNS Verification")
func newVPCEndpointServicePrivateDNSVerificationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcEndpointServicePrivateDNSVerificationResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)

	return r, nil
}

type vpcEndpointServicePrivateDNSVerificationResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpRead
	framework.WithNoUpdate
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (*vpcEndpointServicePrivateDNSVerificationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_vpc_endpoint_service_private_dns_verification"
}

func (r *vpcEndpointServicePrivateDNSVerificationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"wait_for_verification": schema.BoolAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *vpcEndpointServicePrivateDNSVerificationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcEndpointServicePrivateDNSVerificationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.StartVpcEndpointServicePrivateDnsVerificationInput{
		ServiceId: fwflex.StringFromFramework(ctx, data.ServiceID),
	}

	_, err := conn.StartVpcEndpointServicePrivateDnsVerification(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("starting VPC Endpoint Service Private DNS Verification (%s)", data.ServiceID.ValueString()), err.Error())

		return
	}

	if data.WaitForVerification.ValueBool() {
		if _, err := waitVPCEndpointServicePrivateDNSNameVerified(ctx, conn, data.ServiceID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Endpoint Service Private DNS Verification (%s)", data.ServiceID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

type vpcEndpointServicePrivateDNSVerificationResourceModel struct {
	ServiceID           types.String   `tfsdk:"service_id"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
	WaitForVerification types.Bool     `tfsdk:"wait_for_verification"`
}
