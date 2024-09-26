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

// @FrameworkResource("aws_eip_domain_name", name="EIP Domain Name")
func newEIPDomainNameResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &eipDomainNameResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type eipDomainNameResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (*eipDomainNameResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_eip_domain_name"
}

func (r *eipDomainNameResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allocation_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
			"ptr_record": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *eipDomainNameResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data eipDomainNameResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.ModifyAddressAttributeInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.ModifyAddressAttribute(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 EIP Domain Name", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.Address.AllocationId)

	v, err := waitEIPDomainNameAttributeUpdated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 EIP Domain Name (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	data.PTRRecord = fwflex.StringToFramework(ctx, v.PtrRecord)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *eipDomainNameResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data eipDomainNameResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	output, err := findEIPDomainNameAttributeByAllocationID(ctx, conn, data.AllocationID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 EIP Domain Name (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *eipDomainNameResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new eipDomainNameResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	if !new.DomainName.Equal(old.DomainName) {
		input := &ec2.ModifyAddressAttributeInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.ModifyAddressAttribute(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating EC2 EIP Domain Name (%s)", new.ID.ValueString()), err.Error())

			return
		}

		if _, err := waitEIPDomainNameAttributeUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 EIP Domain Name (%s) update", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *eipDomainNameResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data eipDomainNameResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	_, err := conn.ResetAddressAttribute(ctx, &ec2.ResetAddressAttributeInput{
		AllocationId: fwflex.StringFromFramework(ctx, data.ID),
		Attribute:    awstypes.AddressAttributeNameDomainName,
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAllocationIDNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 EIP Domain Name (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitEIPDomainNameAttributeDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 EIP Domain Name (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

type eipDomainNameResourceModel struct {
	AllocationID types.String   `tfsdk:"allocation_id"`
	ID           types.String   `tfsdk:"id"`
	DomainName   types.String   `tfsdk:"domain_name"`
	PTRRecord    types.String   `tfsdk:"ptr_record"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}
