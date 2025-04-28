// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpclattice_resource_gateway", name="Resource Gateway")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/vpclattice;vpclattice.GetResourceGatewayOutput")
func newResourceGatewayResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGatewayResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type resourceGatewayResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceGatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrIPAddressType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ResourceGatewayIpAddressType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 40),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ResourceGatewayStatus](),
				Computed:   true,
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
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

func (r *resourceGatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceGatewayResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	input := vpclattice.CreateResourceGatewayInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)
	input.VpcIdentifier = fwflex.StringFromFramework(ctx, data.VPCID)

	outputCRG, err := conn.CreateResourceGateway(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating VPCLattice Resource Gateway (%s)", data.Name.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, outputCRG.Id)

	outputGRG, err := waitResourceGatewayActive(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPCLattice Resource Gateway (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGRG, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceGatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceGatewayResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	output, err := findResourceGatewayByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPCLattice Resource Gateway (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceGatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceGatewayResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	// Only security group IDs can be updated.
	if !new.SecurityGroupIDs.Equal(old.SecurityGroupIDs) {
		input := vpclattice.UpdateResourceGatewayInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ResourceGatewayIdentifier = fwflex.StringFromFramework(ctx, new.ID)

		_, err := conn.UpdateResourceGateway(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating VPCLattice Resource Gateway (%s)", new.ID.ValueString()), err.Error())

			return
		}

		outputGRG, err := waitResourceGatewayActive(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for VPCLattice Resource Gateway (%s) update", new.ID.ValueString()), err.Error())

			return
		}

		new.Status = fwtypes.StringEnumValue(outputGRG.Status)
	} else {
		new.Status = old.Status
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceGatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceGatewayResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	input := vpclattice.DeleteResourceGatewayInput{
		ResourceGatewayIdentifier: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteResourceGateway(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPCLattice Resource Gateway (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitResourceGatewayDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPCLattice Resource Gateway (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findResourceGatewayByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetResourceGatewayOutput, error) {
	input := vpclattice.GetResourceGatewayInput{
		ResourceGatewayIdentifier: aws.String(id),
	}

	output, err := conn.GetResourceGateway(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusResourceGateway(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findResourceGatewayByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitResourceGatewayActive(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetResourceGatewayOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceGatewayStatusCreateInProgress, awstypes.ResourceGatewayStatusUpdateInProgress),
		Target:                    enum.Slice(awstypes.ResourceGatewayStatusActive),
		Refresh:                   statusResourceGateway(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetResourceGatewayOutput); ok {
		return output, err
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

	if output, ok := outputRaw.(*vpclattice.GetResourceGatewayOutput); ok {
		return output, err
	}

	return nil, err
}

type resourceGatewayResourceModel struct {
	ARN              types.String                                              `tfsdk:"arn"`
	ID               types.String                                              `tfsdk:"id"`
	IPAddressType    fwtypes.StringEnum[awstypes.ResourceGatewayIpAddressType] `tfsdk:"ip_address_type"`
	Name             types.String                                              `tfsdk:"name"`
	SecurityGroupIDs fwtypes.SetOfString                                       `tfsdk:"security_group_ids"`
	Status           fwtypes.StringEnum[awstypes.ResourceGatewayStatus]        `tfsdk:"status"`
	SubnetIDs        fwtypes.SetOfString                                       `tfsdk:"subnet_ids"`
	Tags             tftags.Map                                                `tfsdk:"tags"`
	TagsAll          tftags.Map                                                `tfsdk:"tags_all"`
	Timeouts         timeouts.Value                                            `tfsdk:"timeouts"`
	VPCID            types.String                                              `tfsdk:"vpc_id"`
}
