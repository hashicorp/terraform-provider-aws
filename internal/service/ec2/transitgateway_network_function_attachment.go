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
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_transit_gateway_network_function_attachment", name="Transit Gateway Network Function Attachment")
// @Tags(identifierAttribute="id")
// @IdentityAttribute("id")
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.TransitGatewayNetworkFunctionAttachment")
// @Testing(serialize=true)
// @Testing(preCheck="testAccPreCheckTransitGateway")
func newTransitGatewayNetworkFunctionAttachmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &transitGatewayNetworkFunctionAttachmentResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type transitGatewayNetworkFunctionAttachmentResource struct {
	framework.ResourceWithModel[transitGatewayNetworkFunctionAttachmentResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *transitGatewayNetworkFunctionAttachmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"availability_zone_ids": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 16),
					setvalidator.ExactlyOneOf(
						path.MatchRoot("availability_zone_ids"),
						path.MatchRoot(names.AttrAvailabilityZones),
					),
				},
			},
			names.AttrAvailabilityZones: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 16),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrServiceName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrTransitGatewayID: schema.StringAttribute{
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

// ModifyPlan keeps the two views of the Availability Zones (names and IDs) consistent.
// Only one of them is configured; the other is computed by the API. When the configured
// one changes, the computed one is only known after apply. When it does not change,
// the computed one keeps its value, so that an unrelated update (tags) shows no diff on it.
func (r *transitGatewayNetworkFunctionAttachmentResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	var config, plan, state transitGatewayNetworkFunctionAttachmentResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if config.AvailabilityZoneIDs.IsNull() {
		computed := state.AvailabilityZoneIDs
		if !plan.AvailabilityZones.Equal(state.AvailabilityZones) {
			computed = fwtypes.NewSetValueOfUnknown[types.String](ctx)
		}
		response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root("availability_zone_ids"), computed)...)
	}

	if config.AvailabilityZones.IsNull() {
		computed := state.AvailabilityZones
		if !plan.AvailabilityZoneIDs.Equal(state.AvailabilityZoneIDs) {
			computed = fwtypes.NewSetValueOfUnknown[types.String](ctx)
		}
		response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrAvailabilityZones), computed)...)
	}
}

func (r *transitGatewayNetworkFunctionAttachmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data transitGatewayNetworkFunctionAttachmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	c := r.Meta()
	conn := c.EC2Client(ctx)

	serviceName := fwflex.StringValueFromFramework(ctx, data.ServiceName)
	input := ec2.CreateTransitGatewayNetworkFunctionAttachmentInput{
		ServiceName:       aws.String(serviceName),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeTransitGatewayAttachment),
		TransitGatewayId:  fwflex.StringFromFramework(ctx, data.TransitGatewayID),
	}

	if !data.AvailabilityZoneIDs.IsNull() && !data.AvailabilityZoneIDs.IsUnknown() {
		input.AvailabilityZoneIds = fwflex.ExpandFrameworkStringValueSet(ctx, data.AvailabilityZoneIDs)
	}

	if !data.AvailabilityZones.IsNull() && !data.AvailabilityZones.IsUnknown() {
		input.AvailabilityZones = fwflex.ExpandFrameworkStringValueSet(ctx, data.AvailabilityZones)
	}

	output, err := conn.CreateTransitGatewayNetworkFunctionAttachment(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating EC2 Transit Gateway Network Function Attachment", err.Error())
		return
	}

	id := aws.ToString(output.TransitGatewayNetworkFunctionAttachment.TransitGatewayAttachmentId)

	// From here on the attachment exists: keep its ID even if waiting fails.
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), id)...)
	if response.Diagnostics.HasError() {
		return
	}

	attachment, err := waitTransitGatewayNetworkFunctionAttachmentCreated(ctx, conn, id, r.acceptanceRequired(ctx, serviceName), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Network Function Attachment (%s) create", id), err.Error())
		return
	}

	response.Diagnostics.Append(r.flatten(ctx, c, attachment, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *transitGatewayNetworkFunctionAttachmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data transitGatewayNetworkFunctionAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	c := r.Meta()
	conn := c.EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	attachment, err := findTransitGatewayNetworkFunctionAttachmentByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Transit Gateway Network Function Attachment (%s)", id), err.Error())
		return
	}

	response.Diagnostics.Append(r.flatten(ctx, c, attachment, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, attachment.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *transitGatewayNetworkFunctionAttachmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old transitGatewayNetworkFunctionAttachmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	c := r.Meta()
	conn := c.EC2Client(ctx)

	if !new.AvailabilityZoneIDs.Equal(old.AvailabilityZoneIDs) || !new.AvailabilityZones.Equal(old.AvailabilityZones) {
		id := fwflex.StringValueFromFramework(ctx, new.ID)
		input := ec2.ModifyTransitGatewayNetworkFunctionAttachmentInput{
			TransitGatewayAttachmentId: aws.String(id),
		}

		// Only the configured view of the zones is known in the plan; the other one is computed.
		var zonesAdded bool
		if !new.AvailabilityZoneIDs.IsUnknown() {
			oldIDs, newIDs := fwflex.ExpandFrameworkStringValueSet(ctx, old.AvailabilityZoneIDs), fwflex.ExpandFrameworkStringValueSet(ctx, new.AvailabilityZoneIDs)
			input.AddAvailabilityZoneIds = newIDs.Difference(oldIDs)
			input.RemoveAvailabilityZoneIds = oldIDs.Difference(newIDs)
			zonesAdded = len(input.AddAvailabilityZoneIds) > 0
		} else {
			oldNames, newNames := fwflex.ExpandFrameworkStringValueSet(ctx, old.AvailabilityZones), fwflex.ExpandFrameworkStringValueSet(ctx, new.AvailabilityZones)
			input.AddAvailabilityZones = newNames.Difference(oldNames)
			input.RemoveAvailabilityZones = oldNames.Difference(newNames)
			zonesAdded = len(input.AddAvailabilityZones) > 0
		}

		_, err := conn.ModifyTransitGatewayNetworkFunctionAttachment(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating EC2 Transit Gateway Network Function Attachment (%s)", id), err.Error())
			return
		}

		// Each added zone opens a new connection to the endpoint service. When the service
		// requires acceptance, the attachment stays in "modifying" until its owner accepts.
		if zonesAdded && r.acceptanceRequired(ctx, fwflex.StringValueFromFramework(ctx, new.ServiceName)) {
			tflog.Info(ctx, "EC2 Transit Gateway Network Function Attachment is waiting for the endpoint service owner to accept the new connection", map[string]any{
				names.AttrID: id,
			})
		} else if _, err := waitTransitGatewayNetworkFunctionAttachmentUpdated(ctx, conn, id, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Network Function Attachment (%s) update", id), err.Error())
			return
		}

		attachment, err := findTransitGatewayNetworkFunctionAttachmentByID(ctx, conn, id)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Transit Gateway Network Function Attachment (%s)", id), err.Error())
			return
		}

		response.Diagnostics.Append(r.flatten(ctx, c, attachment, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *transitGatewayNetworkFunctionAttachmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data transitGatewayNetworkFunctionAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	input := ec2.DeleteTransitGatewayNetworkFunctionAttachmentInput{
		TransitGatewayAttachmentId: aws.String(id),
	}
	_, err := conn.DeleteTransitGatewayNetworkFunctionAttachment(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return
	}

	// Deleting an attachment that is already being deleted, or was deleted, is answered with
	// IncorrectState and not with a not found error. Only fail if it is in another state.
	if tfawserr.ErrCodeEquals(err, errCodeIncorrectState) {
		attachment, findErr := findTransitGatewayNetworkFunctionAttachmentByID(ctx, conn, id)

		if retry.NotFound(findErr) {
			return
		}

		if findErr == nil && attachment.State == awstypes.TransitGatewayAttachmentStateDeleting {
			err = nil
		}
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Transit Gateway Network Function Attachment (%s)", id), err.Error())
		return
	}

	if _, err := waitTransitGatewayNetworkFunctionAttachmentDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Transit Gateway Network Function Attachment (%s) delete", id), err.Error())
		return
	}
}

// acceptanceRequired reports whether the Gateway Load Balancer endpoint service requires
// its owner to accept connections.
//
// It decides what "created" means. An attachment to a service that accepts connections
// automatically still goes through "pendingAcceptance" for a few minutes before it becomes
// "available", so that state cannot be taken as final unless acceptance is really required.
// If the service cannot be described, acceptance is assumed not to be required and the
// attachment is waited on until it is "available".
func (r *transitGatewayNetworkFunctionAttachmentResource) acceptanceRequired(ctx context.Context, serviceName string) bool {
	conn := r.Meta().EC2Client(ctx)

	input := ec2.DescribeVpcEndpointServicesInput{
		ServiceNames: []string{serviceName},
	}
	serviceDetails, _, err := findVPCEndpointServices(ctx, conn, &input)

	if err != nil {
		tflog.Warn(ctx, "reading EC2 VPC Endpoint Service to determine whether acceptance is required", map[string]any{
			names.AttrServiceName: serviceName,
			"error":               err.Error(),
		})
		return false
	}

	for _, v := range serviceDetails {
		if aws.ToString(v.ServiceName) == serviceName {
			return aws.ToBool(v.AcceptanceRequired)
		}
	}

	return false
}

func (r *transitGatewayNetworkFunctionAttachmentResource) flatten(ctx context.Context, c *conns.AWSClient, attachment *awstypes.TransitGatewayNetworkFunctionAttachment, data *transitGatewayNetworkFunctionAttachmentResourceModel) diag.Diagnostics { //nolint:unparam
	var diags diag.Diagnostics
	id := aws.ToString(attachment.TransitGatewayAttachmentId)
	data.ARN = fwflex.StringValueToFramework(ctx, transitGatewayAttachmentARN(ctx, c, aws.ToString(attachment.TransitGatewayOwnerId), id))
	data.AvailabilityZoneIDs = fwflex.FlattenFrameworkStringValueSetOfString(ctx, attachment.AvailabilityZoneIds)
	data.AvailabilityZones = fwflex.FlattenFrameworkStringValueSetOfString(ctx, attachment.AvailabilityZones)
	data.ID = fwflex.StringValueToFramework(ctx, id)
	// The endpoint service name is returned as the attachment's resource ID.
	data.ServiceName = fwflex.StringToFramework(ctx, attachment.ResourceId)
	data.TransitGatewayID = fwflex.StringToFramework(ctx, attachment.TransitGatewayId)
	return diags
}

type transitGatewayNetworkFunctionAttachmentResourceModel struct {
	framework.WithRegionModel
	ARN                 types.String        `tfsdk:"arn"`
	AvailabilityZoneIDs fwtypes.SetOfString `tfsdk:"availability_zone_ids"`
	AvailabilityZones   fwtypes.SetOfString `tfsdk:"availability_zones"`
	ID                  types.String        `tfsdk:"id"`
	ServiceName         types.String        `tfsdk:"service_name"`
	Tags                tftags.Map          `tfsdk:"tags"`
	TagsAll             tftags.Map          `tfsdk:"tags_all"`
	Timeouts            timeouts.Value      `tfsdk:"timeouts"`
	TransitGatewayID    types.String        `tfsdk:"transit_gateway_id"`
}
