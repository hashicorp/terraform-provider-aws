// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying state resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_secondary_network", name="SecondaryNetwork")
// @Tags(identifierAttribute="id")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(tagsTest=false)
// @Testing(serialize=true)
// @Testing(generator=false)
// @Testing(existsTakesT=false, destroyTakesT=false)
func newSecondaryNetworkResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &secondaryNetworkResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type secondaryNetworkResource struct {
	framework.ResourceWithModel[secondaryNetworkResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *secondaryNetworkResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"ipv4_cidr_block": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ipv4_cidr_block_associations": framework.ResourceComputedListOfObjectsAttribute[ipv4CidrBlockAssociationModel](ctx, listplanmodifier.UseStateForUnknown()),
			"network_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(string(awstypes.SecondaryNetworkTypeRdma)),
				},
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secondary_network_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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

func (r *secondaryNetworkResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data secondaryNetworkResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	var input ec2.CreateSecondaryNetworkInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeSecondaryNetwork)

	output, err := conn.CreateSecondaryNetwork(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating EC2 Secondary Network", err.Error())
		return
	}
	data.ID = fwflex.StringToFramework(ctx, output.SecondaryNetwork.SecondaryNetworkId)

	waitOutput, err := waitSecondaryNetworkCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Secondary Network (%s) create", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, waitOutput, &data, fwflex.WithFieldNamePrefix("SecondaryNetwork"))...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, waitOutput.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *secondaryNetworkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data secondaryNetworkResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)
	id := data.ID.ValueString()

	output, err := findSecondaryNetworkByID(ctx, conn, id)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Secondary Network (%s)", id), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix("SecondaryNetwork"))...)
	if response.Diagnostics.HasError() {
		return
	}
	data.NetworkType = types.StringValue(string(output.Type))

	// Try to get IPv4 CIDR block from associations if available
	if len(output.Ipv4CidrBlockAssociations) > 0 && output.Ipv4CidrBlockAssociations[0].CidrBlock != nil {
		data.IPv4CidrBlock = types.StringValue(aws.ToString(output.Ipv4CidrBlockAssociations[0].CidrBlock))
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *secondaryNetworkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data secondaryNetworkResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)
	id := data.ID.ValueString()

	input := ec2.DeleteSecondaryNetworkInput{
		SecondaryNetworkId: aws.String(id),
	}

	_, err := conn.DeleteSecondaryNetwork(ctx, &input)
	if tfawserr.ErrCodeEquals(err, errCodeInvalidSecondaryNetworkIdNotFound) {
		return
	}
	if tfawserr.ErrMessageContains(err, errCodeInvalidState, "is not in a modifiable state") {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Secondary Network (%s)", id), err.Error())
		return
	}

	if _, err := waitSecondaryNetworkDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Secondary Network (%s) delete", id), err.Error())
		return
	}
}

type secondaryNetworkResourceModel struct {
	framework.WithRegionModel
	ARN                       types.String                                                   `tfsdk:"arn"`
	ID                        types.String                                                   `tfsdk:"id"`
	IPv4CidrBlock             types.String                                                   `tfsdk:"ipv4_cidr_block"`
	IPv4CidrBlockAssociations fwtypes.ListNestedObjectValueOf[ipv4CidrBlockAssociationModel] `tfsdk:"ipv4_cidr_block_associations"`
	NetworkType               types.String                                                   `tfsdk:"network_type"`
	OwnerID                   types.String                                                   `tfsdk:"owner_id"`
	SecondaryNetworkId        types.String                                                   `tfsdk:"secondary_network_id"`
	State                     types.String                                                   `tfsdk:"state"`
	Tags                      tftags.Map                                                     `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                     `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                 `tfsdk:"timeouts"`
}

type ipv4CidrBlockAssociationModel struct {
	AssociationID types.String `tfsdk:"association_id"`
	CidrBlock     types.String `tfsdk:"cidr_block"`
	State         types.String `tfsdk:"state"`
}
