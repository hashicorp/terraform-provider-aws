// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Instance Connect Endpoint")
// @Tags(identifierAttribute="id")
func newResourceInstanceConnectEndpoint(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceInstanceConnectEndpoint{}
	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type resourceInstanceConnectEndpoint struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceInstanceConnectEndpoint) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_ec2_instance_connect_endpoint"
}

func (r *resourceInstanceConnectEndpoint) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zone": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dns_name": schema.StringAttribute{
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
			"network_interface_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"owner_id": schema.StringAttribute{
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
			"security_group_ids": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"vpc_id": schema.StringAttribute{
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

func (r *resourceInstanceConnectEndpoint) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceInstanceConnectEndpointData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.CreateInstanceConnectEndpointInput{
		ClientToken:       aws.String(id.UniqueId()),
		PreserveClientIp:  aws.Bool(data.PreserveClientIp.ValueBool()),
		SecurityGroupIds:  flex.ExpandFrameworkStringValueSet(ctx, data.SecurityGroupIds),
		SubnetId:          aws.String(data.SubnetId.ValueString()),
		TagSpecifications: getTagSpecificationsInV2(ctx, awstypes.ResourceTypeInstanceConnectEndpoint),
	}

	output, err := conn.CreateInstanceConnectEndpoint(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 Instance Connect Endpoint", err.Error())

		return
	}

	data.InstanceConnectEndpointId = types.StringPointerValue(output.InstanceConnectEndpoint.InstanceConnectEndpointId)
	id := data.InstanceConnectEndpointId.ValueString()

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	instanceConnectEndpoint, err := WaitInstanceConnectEndpointCreated(ctx, conn, id, createTimeout)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Instance Connect Endpoint (%s) create", id), err.Error())

		return
	}

	// Set values for unknowns.
	if err := flex.Flatten(ctx, instanceConnectEndpoint, &data); err != nil {
		response.Diagnostics.AddError("flattening data", err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceInstanceConnectEndpoint) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceInstanceConnectEndpointData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := data.InstanceConnectEndpointId.ValueString()
	instanceConnectEndpoint, err := FindInstanceConnectEndpointByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Instance Connect Endpoint (%s)", id), err.Error())

		return
	}

	if err := flex.Flatten(ctx, instanceConnectEndpoint, &data); err != nil {
		response.Diagnostics.AddError("flattening data", err.Error())

		return
	}

	setTagsOutV2(ctx, instanceConnectEndpoint.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceInstanceConnectEndpoint) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Tags only.
}

func (r *resourceInstanceConnectEndpoint) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceInstanceConnectEndpointData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	_, err := conn.DeleteInstanceConnectEndpoint(ctx, &ec2.DeleteInstanceConnectEndpointInput{
		InstanceConnectEndpointId: flex.StringFromFramework(ctx, data.InstanceConnectEndpointId),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceConnectEndpointIdNotFound) {
		return
	}

	id := data.InstanceConnectEndpointId.ValueString()

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Instance Connect Endpoint (%s)", id), err.Error())

		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	if _, err := WaitInstanceConnectEndpointDeleted(ctx, conn, id, deleteTimeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Instance Connect Endpoint (%s) delete", id), err.Error())

		return
	}
}

func (r *resourceInstanceConnectEndpoint) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_Ec2InstanceConnectEndpoint.html.
type resourceInstanceConnectEndpointData struct {
	InstanceConnectEndpointArn types.String   `tfsdk:"arn"`
	AvailabilityZone           types.String   `tfsdk:"availability_zone"`
	DnsName                    types.String   `tfsdk:"dns_name"`
	FipsDnsName                types.String   `tfsdk:"fips_dns_name"`
	InstanceConnectEndpointId  types.String   `tfsdk:"id"`
	NetworkInterfaceIds        types.List     `tfsdk:"network_interface_ids"`
	OwnerId                    types.String   `tfsdk:"owner_id"`
	PreserveClientIp           types.Bool     `tfsdk:"preserve_client_ip"`
	SecurityGroupIds           types.Set      `tfsdk:"security_group_ids"`
	SubnetId                   types.String   `tfsdk:"subnet_id"`
	Tags                       types.Map      `tfsdk:"tags"`
	TagsAll                    types.Map      `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value `tfsdk:"timeouts"`
	VpcId                      types.String   `tfsdk:"vpc_id"`
}
