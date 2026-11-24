// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_public_ipv4_pool", name="Public IPv4 Pool")
// @Tags(identifierAttribute="pool_id")
// @IdentityAttribute("pool_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.PublicIpv4Pool")
// @Testing(generator=false)
// @Testing(importStateIdAttribute="pool_id")
func newPublicIPv4PoolResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &publicIPv4PoolResource{}

	return r, nil
}

type publicIPv4PoolResource struct {
	framework.ResourceWithModel[publicIPv4PoolResourceModel]
	framework.WithImportByIdentity
}

func (r *publicIPv4PoolResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Computed:    true,
				Description: "ARN, arn:<partition>:ec2:<region>:<account>:ipv4pool-ec2/<pool_id>.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Computed:    true,
				Description: "Description of the address pool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_border_group": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Availability Zone (AZ) or Local Zone (LZ) network border group that the resource that the IP address is assigned to is in. Defaults to an AZ network border group. For more information on available Local Zones, see Local Zone availability (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-byoip.html#byoip-zone-avail) in the Amazon EC2 User Guide.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pool_address_ranges": framework.ResourceComputedListOfObjectsAttribute[publicIpv4PoolRangeModel](ctx, listplanmodifier.UseStateForUnknown()),
			"pool_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the address pool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"total_address_count": schema.Int32Attribute{
				Computed:    true,
				Description: "Total number of addresses.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"total_available_address_count": schema.Int32Attribute{
				Computed:    true,
				Description: "Total number of available addresses.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *publicIPv4PoolResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan publicIPv4PoolResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	var input ec2.CreatePublicIpv4PoolInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeIpv4poolEc2)

	out, err := conn.CreatePublicIpv4Pool(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	// The ARN no API returns, which buildARN composes from the identity.
	plan.ARN = fwtypes.ARNValue(r.buildARN(ctx, &plan))

	// Every identity attribute is in state before the next call, so a failure from here on taints the resource.
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("pool_id"), plan.PoolID))
	if response.Diagnostics.HasError() {
		return
	}

	id := fwflex.StringValueFromFramework(ctx, plan.PoolID)

	// Describe after Create is eventually consistent: the finder is retried while it reports not found.
	output, err := tfresource.RetryWhenNotFound(ctx, ec2PropagationTimeout, func(ctx context.Context) (*awstypes.PublicIpv4Pool, error) {
		return findPublicIPv4PoolByID(ctx, conn, id)
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *publicIPv4PoolResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data publicIPv4PoolResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.PoolID)
	output, err := findPublicIPv4PoolByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

// flatten sets the model from the finder's value, for Read and for an item the list resource includes.
func (r *publicIPv4PoolResource) flatten(ctx context.Context, output *awstypes.PublicIpv4Pool, data *publicIPv4PoolResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, output, data)...)
	if diags.HasError() {
		return diags
	}

	data.ARN = fwtypes.ARNValue(r.buildARN(ctx, data))

	setTagsOut(ctx, output.Tags)

	return diags
}

func (r *publicIPv4PoolResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data publicIPv4PoolResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.PoolID)
	input := ec2.DeletePublicIpv4PoolInput{
		PoolId:             aws.String(id),
		NetworkBorderGroup: fwflex.StringFromFramework(ctx, data.NetworkBorderGroup),
	}
	_, err := conn.DeletePublicIpv4Pool(ctx, &input)
	if tfawserr.ErrCodeEquals(err, errCodeInvalidPublicIpv4PoolNotFound) || tfawserr.ErrCodeEquals(err, errCodeInvalidPublicIpv4PoolIDNotFound) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

// buildARN composes the ARN no API returns from the partition, the Region, the account and the identity attribute pool_id.
func (r *publicIPv4PoolResource) buildARN(ctx context.Context, data *publicIPv4PoolResourceModel) string {
	return r.Meta().RegionalARN(ctx, "ec2", "ipv4pool-ec2/"+data.PoolID.ValueString())
}

type publicIPv4PoolResourceModel struct {
	framework.WithRegionModel
	ARN                        fwtypes.ARN                                               `tfsdk:"arn" autoflex:"-"`
	Description                types.String                                              `tfsdk:"description"`
	NetworkBorderGroup         types.String                                              `tfsdk:"network_border_group"`
	PoolAddressRanges          fwtypes.ListNestedObjectValueOf[publicIpv4PoolRangeModel] `tfsdk:"pool_address_ranges"`
	PoolID                     types.String                                              `tfsdk:"pool_id"`
	Tags                       tftags.Map                                                `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                `tfsdk:"tags_all"`
	TotalAddressCount          types.Int32                                               `tfsdk:"total_address_count"`
	TotalAvailableAddressCount types.Int32                                               `tfsdk:"total_available_address_count"`
}

type publicIpv4PoolRangeModel struct {
	AddressCount          types.Int32  `tfsdk:"address_count"`
	AvailableAddressCount types.Int32  `tfsdk:"available_address_count"`
	FirstAddress          types.String `tfsdk:"first_address"`
	LastAddress           types.String `tfsdk:"last_address"`
}
