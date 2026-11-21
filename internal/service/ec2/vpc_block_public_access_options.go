// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
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
	framework.ResourceWithModel[vpcBlockPublicAccessOptionsResourceModel]
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
		// Preserve regional identity so recovery of the tainted resource cannot use the provider's default Region.
		smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID))
		smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrRegion), r.Meta().Region(ctx)))
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

	if err := r.checkRegion(ctx, data.ID); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	conn := r.Meta().EC2Client(ctx)

	options, err := findVPCBlockPublicAccessOptions(ctx, conn)

	if retry.NotFound(err) {
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

	if err := r.checkRegion(ctx, new.ID); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ID.ValueString())
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

	if err := r.checkRegion(ctx, data.ID); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
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

func (r *vpcBlockPublicAccessOptionsResource) checkRegion(ctx context.Context, id types.String) error {
	if id.IsNull() || id.IsUnknown() || id.ValueString() == "" {
		return errors.New("refusing to access VPC Block Public Access Options without a known regional resource ID")
	}

	// Older failed creates can have an ID but no region. Never refresh those options or disable BPA in another Region.
	if region := r.Meta().Region(ctx); region != id.ValueString() {
		return fmt.Errorf("refusing to access VPC Block Public Access Options: resource ID identifies Region %q but the selected Region is %q; restore the resource's region in state to match its ID before retrying", id.ValueString(), region)
	}

	return nil
}

type vpcBlockPublicAccessOptionsResourceModel struct {
	framework.WithRegionModel
	AWSAccountID             types.String                                          `tfsdk:"aws_account_id"`
	AWSRegion                types.String                                          `tfsdk:"aws_region"`
	ID                       types.String                                          `tfsdk:"id"`
	InternetGatewayBlockMode fwtypes.StringEnum[awstypes.InternetGatewayBlockMode] `tfsdk:"internet_gateway_block_mode"`
	Timeouts                 timeouts.Value                                        `tfsdk:"timeouts"`
}
