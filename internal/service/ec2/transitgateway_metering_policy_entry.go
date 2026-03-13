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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_transit_gateway_metering_policy_entry", name="Transit Gateway Metering Policy Entry")
func newTransitGatewayMeteringPolicyEntryResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &transitGatewayMeteringPolicyEntryResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type transitGatewayMeteringPolicyEntryResource struct {
	framework.ResourceWithModel[transitGatewayMeteringPolicyEntryResourceModel]
	framework.WithNoUpdate
	framework.WithTimeouts
}

func (r *transitGatewayMeteringPolicyEntryResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"destination_cidr_block": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_port_range": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_transit_gateway_attachment_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_transit_gateway_attachment_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TransitGatewayAttachmentResourceType](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metered_account": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TransitGatewayMeteringPayerType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_rule_number": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			names.AttrProtocol: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_cidr_block": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_port_range": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_transit_gateway_attachment_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_transit_gateway_attachment_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TransitGatewayAttachmentResourceType](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"transit_gateway_metering_policy_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

func (r *transitGatewayMeteringPolicyEntryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data transitGatewayMeteringPolicyEntryResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	var input ec2.CreateTransitGatewayMeteringPolicyEntryInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateTransitGatewayMeteringPolicyEntry(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating EC2 Transit Gateway Metering Policy Entry", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *transitGatewayMeteringPolicyEntryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data transitGatewayMeteringPolicyEntryResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	policyID, ruleNumber := fwflex.StringValueFromFramework(ctx, data.TransitGatewayMeteringPolicyID), intflex.Int32ValueToStringValue(fwflex.Int32ValueFromFrameworkInt64(ctx, data.PolicyRuleNumber))
	entry, err := findTransitGatewayMeteringPolicyEntryByTwoPartKey(ctx, conn, policyID, ruleNumber)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Transit Gateway Metering Policy (%s) Entry (%s)", policyID, ruleNumber), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, entry, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(fwflex.Flatten(ctx, entry.MeteringPolicyRule, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *transitGatewayMeteringPolicyEntryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data transitGatewayMeteringPolicyEntryResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	policyID, ruleNumber := fwflex.StringValueFromFramework(ctx, data.TransitGatewayMeteringPolicyID), fwflex.Int32ValueFromFrameworkInt64(ctx, data.PolicyRuleNumber)
	input := ec2.DeleteTransitGatewayMeteringPolicyEntryInput{
		PolicyRuleNumber:               aws.Int32(ruleNumber),
		TransitGatewayMeteringPolicyId: aws.String(policyID),
	}
	_, err := conn.DeleteTransitGatewayMeteringPolicyEntry(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMeteringPolicyIdNotFound, errCodeInvalidTransitGatewayMeteringPolicyEntryNotFound) ||
		tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "RuleNumber provided does not exists") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Transit Gateway Metering Policy (%s) Entry (%d)", policyID, ruleNumber), err.Error())
		return
	}
}

func (r *transitGatewayMeteringPolicyEntryResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		transitGatewayMeteringPolicyEntryIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, transitGatewayMeteringPolicyEntryIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("transit_gateway_metering_policy_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("policy_rule_number"), intflex.StringValueToInt64Value(parts[1]))...)
}

type transitGatewayMeteringPolicyEntryResourceModel struct {
	framework.WithRegionModel
	DestinationCIDRBlock                    types.String                                                      `tfsdk:"destination_cidr_block"`
	DestinationPortRange                    types.String                                                      `tfsdk:"destination_port_range"`
	DestinationTransitGatewayAttachmentID   types.String                                                      `tfsdk:"destination_transit_gateway_attachment_id"`
	DestinationTransitGatewayAttachmentType fwtypes.StringEnum[awstypes.TransitGatewayAttachmentResourceType] `tfsdk:"destination_transit_gateway_attachment_type"`
	MeteredAccount                          fwtypes.StringEnum[awstypes.TransitGatewayMeteringPayerType]      `tfsdk:"metered_account"`
	PolicyRuleNumber                        types.Int64                                                       `tfsdk:"policy_rule_number" autoflex:",noflatten"`
	Protocol                                types.String                                                      `tfsdk:"protocol"`
	SourceCIDRBlock                         types.String                                                      `tfsdk:"source_cidr_block"`
	SourcePortRange                         types.String                                                      `tfsdk:"source_port_range"`
	SourceTransitGatewayAttachmentID        types.String                                                      `tfsdk:"source_transit_gateway_attachment_id"`
	SourceTransitGatewayAttachmentType      fwtypes.StringEnum[awstypes.TransitGatewayAttachmentResourceType] `tfsdk:"source_transit_gateway_attachment_type"`
	Timeouts                                timeouts.Value                                                    `tfsdk:"timeouts"`
	TransitGatewayMeteringPolicyID          types.String                                                      `tfsdk:"transit_gateway_metering_policy_id"`
}
