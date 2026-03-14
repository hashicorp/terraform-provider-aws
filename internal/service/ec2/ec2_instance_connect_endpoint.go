// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_instance_connect_endpoint", name="Instance Connect Endpoint")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func newInstanceConnectEndpointResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &instanceConnectEndpointResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type instanceConnectEndpointResource struct {
	framework.ResourceWithModel[instanceConnectEndpointResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *instanceConnectEndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrAvailabilityZone: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDNSName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fips_dns_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrIPAddressType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IpAddressType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_interface_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"preserve_client_ip": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSecurityGroupIDs: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSubnetID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *instanceConnectEndpointResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data instanceConnectEndpointResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.CreateInstanceConnectEndpointInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, &data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeInstanceConnectEndpoint)

	output, err := conn.CreateInstanceConnectEndpoint(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 Instance Connect Endpoint", err.Error())

		return
	}

	data.InstanceConnectEndpointID = types.StringPointerValue(output.InstanceConnectEndpoint.InstanceConnectEndpointId)
	id := data.InstanceConnectEndpointID.ValueString()

	instanceConnectEndpoint, err := waitInstanceConnectEndpointCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Instance Connect Endpoint (%s) create", id), err.Error())

		return
	}

	// Fix missing FipsDnsName in regions without FIPS endpoint support.
	if instanceConnectEndpoint.FipsDnsName == nil {
		instanceConnectEndpoint.FipsDnsName = aws.String("")
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, instanceConnectEndpoint, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *instanceConnectEndpointResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data instanceConnectEndpointResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := data.InstanceConnectEndpointID.ValueString()
	instanceConnectEndpoint, err := findInstanceConnectEndpointByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Instance Connect Endpoint (%s)", id), err.Error())

		return
	}

	// Fix missing FipsDnsName in regions without FIPS endpoint support.
	if instanceConnectEndpoint.FipsDnsName == nil {
		instanceConnectEndpoint.FipsDnsName = aws.String("")
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, instanceConnectEndpoint, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, instanceConnectEndpoint.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *instanceConnectEndpointResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data instanceConnectEndpointResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.DeleteInstanceConnectEndpointInput{
		InstanceConnectEndpointId: fwflex.StringFromFramework(ctx, data.InstanceConnectEndpointID),
	}
	_, err := conn.DeleteInstanceConnectEndpoint(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceConnectEndpointIdNotFound) {
		return
	}

	id := data.InstanceConnectEndpointID.ValueString()

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Instance Connect Endpoint (%s)", id), err.Error())

		return
	}

	if _, err := waitInstanceConnectEndpointDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Instance Connect Endpoint (%s) delete", id), err.Error())

		return
	}
}

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Ec2InstanceConnectEndpoint.html.
type instanceConnectEndpointResourceModel struct {
	framework.WithRegionModel
	InstanceConnectEndpointARN types.String                               `tfsdk:"arn"`
	AvailabilityZone           types.String                               `tfsdk:"availability_zone"`
	DNSName                    types.String                               `tfsdk:"dns_name"`
	FipsDnsName                types.String                               `tfsdk:"fips_dns_name"`
	InstanceConnectEndpointID  types.String                               `tfsdk:"id"`
	IPAddressType              fwtypes.StringEnum[awstypes.IpAddressType] `tfsdk:"ip_address_type"`
	NetworkInterfaceIDs        fwtypes.ListOfString                       `tfsdk:"network_interface_ids"`
	OwnerID                    types.String                               `tfsdk:"owner_id"`
	PreserveClientIP           types.Bool                                 `tfsdk:"preserve_client_ip"`
	SecurityGroupIDs           fwtypes.SetOfString                        `tfsdk:"security_group_ids"`
	SubnetId                   types.String                               `tfsdk:"subnet_id"`
	Tags                       tftags.Map                                 `tfsdk:"tags"`
	TagsAll                    tftags.Map                                 `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                             `tfsdk:"timeouts"`
	VpcId                      types.String                               `tfsdk:"vpc_id"`
}
