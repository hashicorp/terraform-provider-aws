// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_transit_gateway_metering_policy", name="Transit Gateway Metering Policy")
// @Tags(identifierAttribute="transit_gateway_metering_policy_id")
// @IdentityAttribute("transit_gateway_metering_policy_id")
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.TransitGatewayMeteringPolicy")
// @Testing(serialize=true)
func newTransitGatewayMeteringPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &transitGatewayMeteringPolicyResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type transitGatewayMeteringPolicyResource struct {
	framework.ResourceWithModel[transitGatewayMeteringPolicyResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *transitGatewayMeteringPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"middlebox_attachment_ids": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrTransitGatewayID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"transit_gateway_metering_policy_id": framework.IDAttribute(),
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

func (r *transitGatewayMeteringPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data transitGatewayMeteringPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	c := r.Meta()
	conn := c.EC2Client(ctx)

	input := ec2.CreateTransitGatewayMeteringPolicyInput{
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeTransitGatewayMeteringPolicy),
		TransitGatewayId:  fwflex.StringFromFramework(ctx, data.TransitGatewayID),
	}

	if !data.MiddleboxAttachmentIDs.IsNull() && !data.MiddleboxAttachmentIDs.IsUnknown() {
		input.MiddleboxAttachmentIds = fwflex.ExpandFrameworkStringValueSet(ctx, data.MiddleboxAttachmentIDs)
	}

	output, err := conn.CreateTransitGatewayMeteringPolicy(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating EC2 Transit Gateway Metering Policy", err.Error())
		return
	}

	id := aws.ToString(output.TransitGatewayMeteringPolicy.TransitGatewayMeteringPolicyId)
	data.ARN = fwflex.StringValueToFramework(ctx, transitGatewayMeteringPolicyARN(ctx, c, id))
	data.TransitGatewayMeteringPolicyID = fwflex.StringValueToFramework(ctx, id)

	if _, err := waitTransitGatewayMeteringPolicyCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Metering Policy (%s) create", id), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *transitGatewayMeteringPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data transitGatewayMeteringPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	c := r.Meta()
	conn := c.EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.TransitGatewayMeteringPolicyID)
	policy, err := findTransitGatewayMeteringPolicyByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Transit Gateway Metering Policy (%s)", id), err.Error())
		return
	}

	data.ARN = fwflex.StringValueToFramework(ctx, transitGatewayMeteringPolicyARN(ctx, c, id))
	data.MiddleboxAttachmentIDs = fwflex.FlattenFrameworkStringValueSetOfString(ctx, policy.MiddleboxAttachmentIds)
	data.TransitGatewayID = fwflex.StringToFramework(ctx, policy.TransitGatewayId)

	setTagsOut(ctx, policy.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *transitGatewayMeteringPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old transitGatewayMeteringPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	if !new.MiddleboxAttachmentIDs.Equal(old.MiddleboxAttachmentIDs) {
		id := fwflex.StringValueFromFramework(ctx, new.TransitGatewayMeteringPolicyID)

		oldIDs := inttypes.Set[string](fwflex.ExpandFrameworkStringValueSet(ctx, old.MiddleboxAttachmentIDs))
		newIDs := inttypes.Set[string](fwflex.ExpandFrameworkStringValueSet(ctx, new.MiddleboxAttachmentIDs))
		add := newIDs.Difference(oldIDs)
		remove := oldIDs.Difference(newIDs)

		input := ec2.ModifyTransitGatewayMeteringPolicyInput{
			AddMiddleboxAttachmentIds:      add,
			RemoveMiddleboxAttachmentIds:   remove,
			TransitGatewayMeteringPolicyId: aws.String(id),
		}

		_, err := conn.ModifyTransitGatewayMeteringPolicy(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating EC2 Transit Gateway Metering Policy (%s)", id), err.Error())
			return
		}

		if _, err := waitTransitGatewayMeteringPolicyUpdated(ctx, conn, id, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Metering Policy (%s) update", id), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *transitGatewayMeteringPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data transitGatewayMeteringPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.TransitGatewayMeteringPolicyID)
	input := ec2.DeleteTransitGatewayMeteringPolicyInput{
		TransitGatewayMeteringPolicyId: aws.String(id),
	}
	_, err := conn.DeleteTransitGatewayMeteringPolicy(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMeteringPolicyIdNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Transit Gateway Metering Policy (%s)", id), err.Error())
		return
	}

	if _, err := waitTransitGatewayMeteringPolicyDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Metering Policy (%s) delete", id), err.Error())
		return
	}
}

type transitGatewayMeteringPolicyResourceModel struct {
	framework.WithRegionModel
	ARN                            types.String        `tfsdk:"arn"`
	MiddleboxAttachmentIDs         fwtypes.SetOfString `tfsdk:"middlebox_attachment_ids"`
	Tags                           tftags.Map          `tfsdk:"tags"`
	TagsAll                        tftags.Map          `tfsdk:"tags_all"`
	Timeouts                       timeouts.Value      `tfsdk:"timeouts"`
	TransitGatewayID               types.String        `tfsdk:"transit_gateway_id"`
	TransitGatewayMeteringPolicyID types.String        `tfsdk:"transit_gateway_metering_policy_id"`
}

func transitGatewayMeteringPolicyARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, names.EC2, "transit-gateway-metering-policy/"+id)
}
