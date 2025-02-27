// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_block_public_access_options", name="VPC Block Public Access Options")
func newVPCBlockPublicAccessOptionsResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcBlockPublicAccessOptionsResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type vpcBlockPublicAccessOptionsResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *vpcBlockPublicAccessOptionsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aws_region": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"internet_gateway_block_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InternetGatewayBlockMode](),
				Required:   true,
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

func (r *vpcBlockPublicAccessOptionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcBlockPublicAccessOptionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.ModifyVpcBlockPublicAccessOptionsInput{
		InternetGatewayBlockMode: data.InternetGatewayBlockMode.ValueEnum(),
	}

	output, err := conn.ModifyVpcBlockPublicAccessOptions(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating VPC Block Public Access Options", err.Error())

		return
	}

	// Set values for unknowns.
	data.AWSAccountID = fwflex.StringToFramework(ctx, output.VpcBlockPublicAccessOptions.AwsAccountId)
	data.AWSRegion = fwflex.StringToFramework(ctx, output.VpcBlockPublicAccessOptions.AwsRegion)
	data.ID = data.AWSRegion

	if _, err := waitVPCBlockPublicAccessOptionsUpdated(ctx, conn, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Block Public Access Options (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcBlockPublicAccessOptionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcBlockPublicAccessOptionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	options, err := findVPCBlockPublicAccessOptions(ctx, conn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Block Public Access Options (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, options, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcBlockPublicAccessOptionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new vpcBlockPublicAccessOptionsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.ModifyVpcBlockPublicAccessOptionsInput{
		InternetGatewayBlockMode: new.InternetGatewayBlockMode.ValueEnum(),
	}

	_, err := conn.ModifyVpcBlockPublicAccessOptions(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating VPC Block Public Access Options (%s)", new.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitVPCBlockPublicAccessOptionsUpdated(ctx, conn, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Block Public Access Options (%s) update", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *vpcBlockPublicAccessOptionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpcBlockPublicAccessOptionsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	// On deletion of this resource set the VPC Block Public Access Options to off.
	input := &ec2.ModifyVpcBlockPublicAccessOptionsInput{
		InternetGatewayBlockMode: awstypes.InternetGatewayBlockModeOff,
	}

	_, err := conn.ModifyVpcBlockPublicAccessOptions(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Block Public Access Options (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitVPCBlockPublicAccessOptionsUpdated(ctx, conn, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Block Public Access Options (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

type vpcBlockPublicAccessOptionsResourceModel struct {
	AWSAccountID             types.String                                          `tfsdk:"aws_account_id"`
	AWSRegion                types.String                                          `tfsdk:"aws_region"`
	ID                       types.String                                          `tfsdk:"id"`
	InternetGatewayBlockMode fwtypes.StringEnum[awstypes.InternetGatewayBlockMode] `tfsdk:"internet_gateway_block_mode"`
	Timeouts                 timeouts.Value                                        `tfsdk:"timeouts"`
}
