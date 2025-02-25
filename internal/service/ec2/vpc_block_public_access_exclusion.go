// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_block_public_access_exclusion", name="VPC Block Public Access Exclusion")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=true)
// @Testing(generator=false)
// @Testing(name="BlockPublicAccessExclusion")
func newVPCBlockPublicAccessExclusionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcBlockPublicAccessExclusionResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type vpcBlockPublicAccessExclusionResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *vpcBlockPublicAccessExclusionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"internet_gateway_exclusion_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InternetGatewayExclusionMode](),
				Required:   true,
			},
			names.AttrResourceARN: framework.ARNAttributeComputedOnly(),
			names.AttrSubnetID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Optional: true,
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

func (r *vpcBlockPublicAccessExclusionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceVPCBlockPublicAccessExclusionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.CreateVpcBlockPublicAccessExclusionInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpcBlockPublicAccessExclusion)

	output, err := conn.CreateVpcBlockPublicAccessExclusion(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating VPC Block Public Access Exclusion", err.Error())

		return
	}

	// Set values for unknowns.
	data.ExclusionID = fwflex.StringToFramework(ctx, output.VpcBlockPublicAccessExclusion.ExclusionId)
	data.ResourceARN = fwflex.StringToFramework(ctx, output.VpcBlockPublicAccessExclusion.ResourceArn)

	if _, err := waitVPCBlockPublicAccessExclusionCreated(ctx, conn, data.ExclusionID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ExclusionID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Block Public Access Exclusion (%s) create", data.ExclusionID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *vpcBlockPublicAccessExclusionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceVPCBlockPublicAccessExclusionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	output, err := findVPCBlockPublicAccessExclusionByID(ctx, conn, data.ExclusionID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Block Public Access Exclusion (%s)", data.ExclusionID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Extract VPC ID and Subnet ID.
	resourceARN, err := arn.Parse(data.ResourceARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError("parsing Resource ARN", err.Error())

		return
	}

	if resource := resourceARN.Resource; strings.HasPrefix(resource, "vpc/") {
		data.VPCID = types.StringValue(strings.TrimPrefix(resource, "vpc/"))
	} else if strings.HasPrefix(resource, "subnet/") {
		data.SubnetID = types.StringValue(strings.TrimPrefix(resource, "subnet/"))
	} else {
		response.Diagnostics.AddError("parsing Resource_ARN", fmt.Sprintf("unknown resource type: %s", resource))

		return
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcBlockPublicAccessExclusionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old resourceVPCBlockPublicAccessExclusionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	if !new.InternetGatewayExclusionMode.Equal(old.InternetGatewayExclusionMode) {
		input := &ec2.ModifyVpcBlockPublicAccessExclusionInput{
			ExclusionId:                  fwflex.StringFromFramework(ctx, new.ExclusionID),
			InternetGatewayExclusionMode: new.InternetGatewayExclusionMode.ValueEnum(),
		}

		_, err := conn.ModifyVpcBlockPublicAccessExclusion(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating VPC Block Public Access Exclusion (%s)", new.ExclusionID.ValueString()), err.Error())

			return
		}

		if _, err := waitVPCBlockPublicAccessExclusionUpdated(ctx, conn, new.ExclusionID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Block Public Access Exclusion (%s) update", new.ExclusionID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *vpcBlockPublicAccessExclusionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceVPCBlockPublicAccessExclusionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := &ec2.DeleteVpcBlockPublicAccessExclusionInput{
		ExclusionId: fwflex.StringFromFramework(ctx, data.ExclusionID),
	}

	_, err := conn.DeleteVpcBlockPublicAccessExclusion(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCBlockPublicAccessExclusionIdNotFound) {
		return
	}

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "is in delete-complete state and cannot be deleted") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Block Public Access Exclusion (%s)", data.ExclusionID.ValueString()), err.Error())

		return
	}

	if _, err := waitVPCBlockPublicAccessExclusionDeleted(ctx, conn, data.ExclusionID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPC Block Public Access Exclusion (%s) delete", data.ExclusionID.ValueString()), err.Error())

		return
	}
}

func (r *vpcBlockPublicAccessExclusionResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrSubnetID),
			path.MatchRoot(names.AttrVPCID),
		),
	}
}

type resourceVPCBlockPublicAccessExclusionModel struct {
	ExclusionID                  types.String                                              `tfsdk:"id"`
	InternetGatewayExclusionMode fwtypes.StringEnum[awstypes.InternetGatewayExclusionMode] `tfsdk:"internet_gateway_exclusion_mode"`
	ResourceARN                  types.String                                              `tfsdk:"resource_arn"`
	SubnetID                     types.String                                              `tfsdk:"subnet_id"`
	Tags                         tftags.Map                                                `tfsdk:"tags"`
	TagsAll                      tftags.Map                                                `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                            `tfsdk:"timeouts"`
	VPCID                        types.String                                              `tfsdk:"vpc_id"`
}
