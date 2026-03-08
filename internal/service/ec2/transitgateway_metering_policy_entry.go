// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
			names.AttrID: framework.IDAttribute(),
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
			names.AttrState: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

	policyID := fwflex.StringValueFromFramework(ctx, data.TransitGatewayMeteringPolicyID)
	ruleNumber := data.PolicyRuleNumber.ValueInt64()

	input := ec2.CreateTransitGatewayMeteringPolicyEntryInput{
		TransitGatewayMeteringPolicyId:        aws.String(policyID),
		PolicyRuleNumber:                      aws.Int32(int32(ruleNumber)),
		MeteredAccount:                        awstypes.TransitGatewayMeteringPayerType(data.MeteredAccount.ValueString()),
		DestinationCidrBlock:                  fwflex.StringFromFramework(ctx, data.DestinationCidrBlock),
		DestinationPortRange:                  fwflex.StringFromFramework(ctx, data.DestinationPortRange),
		DestinationTransitGatewayAttachmentId: fwflex.StringFromFramework(ctx, data.DestinationTransitGatewayAttachmentID),
		Protocol:                              fwflex.StringFromFramework(ctx, data.Protocol),
		SourceCidrBlock:                       fwflex.StringFromFramework(ctx, data.SourceCidrBlock),
		SourcePortRange:                       fwflex.StringFromFramework(ctx, data.SourcePortRange),
		SourceTransitGatewayAttachmentId:      fwflex.StringFromFramework(ctx, data.SourceTransitGatewayAttachmentID),
	}

	if !data.DestinationTransitGatewayAttachmentType.IsNull() && !data.DestinationTransitGatewayAttachmentType.IsUnknown() {
		input.DestinationTransitGatewayAttachmentType = awstypes.TransitGatewayAttachmentResourceType(data.DestinationTransitGatewayAttachmentType.ValueString())
	}

	if !data.SourceTransitGatewayAttachmentType.IsNull() && !data.SourceTransitGatewayAttachmentType.IsUnknown() {
		input.SourceTransitGatewayAttachmentType = awstypes.TransitGatewayAttachmentResourceType(data.SourceTransitGatewayAttachmentType.ValueString())
	}

	output, err := conn.CreateTransitGatewayMeteringPolicyEntry(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating EC2 Transit Gateway Metering Policy Entry", err.Error())
		return
	}

	id := transitGatewayMeteringPolicyEntryCreateResourceID(policyID, strconv.FormatInt(ruleNumber, 10))
	data.ID = fwflex.StringValueToFramework(ctx, id)

	entry := output.TransitGatewayMeteringPolicyEntry
	data.State = fwflex.StringValueToFramework(ctx, string(entry.State))
	transitGatewayMeteringPolicyEntryFlattenRule(ctx, entry.MeteringPolicyRule, &data)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *transitGatewayMeteringPolicyEntryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data transitGatewayMeteringPolicyEntryResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	policyID, ruleNumberStr, err := transitGatewayMeteringPolicyEntryParseResourceID(data.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("parsing EC2 Transit Gateway Metering Policy Entry ID", err.Error())
		return
	}

	entry, err := findTransitGatewayMeteringPolicyEntryByTwoPartKey(ctx, conn, policyID, ruleNumberStr)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Transit Gateway Metering Policy Entry (%s)", data.ID.ValueString()), err.Error())
		return
	}

	data.MeteredAccount = fwtypes.StringEnumValue(entry.MeteredAccount)
	data.State = fwflex.StringValueToFramework(ctx, string(entry.State))
	transitGatewayMeteringPolicyEntryFlattenRule(ctx, entry.MeteringPolicyRule, &data)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *transitGatewayMeteringPolicyEntryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data transitGatewayMeteringPolicyEntryResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	policyID, ruleNumberStr, err := transitGatewayMeteringPolicyEntryParseResourceID(data.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("parsing EC2 Transit Gateway Metering Policy Entry ID", err.Error())
		return
	}

	ruleNumber, err := strconv.ParseInt(ruleNumberStr, 10, 32)
	if err != nil {
		response.Diagnostics.AddError("parsing EC2 Transit Gateway Metering Policy Entry rule number", err.Error())
		return
	}

	input := ec2.DeleteTransitGatewayMeteringPolicyEntryInput{
		TransitGatewayMeteringPolicyId: aws.String(policyID),
		PolicyRuleNumber:               aws.Int32(int32(ruleNumber)),
	}
	_, err = conn.DeleteTransitGatewayMeteringPolicyEntry(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMeteringPolicyIdNotFound, errCodeInvalidTransitGatewayMeteringPolicyEntryNotFound) ||
		tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "RuleNumber provided does not exists") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Transit Gateway Metering Policy Entry (%s)", data.ID.ValueString()), err.Error())
		return
	}
}

func (r *transitGatewayMeteringPolicyEntryResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	policyID, ruleNumberStr, err := transitGatewayMeteringPolicyEntryParseResourceID(request.ID)
	if err != nil {
		response.Diagnostics.AddError("importing EC2 Transit Gateway Metering Policy Entry", err.Error())
		return
	}

	ruleNumber, err := strconv.ParseInt(ruleNumberStr, 10, 64)
	if err != nil {
		response.Diagnostics.AddError("importing EC2 Transit Gateway Metering Policy Entry", fmt.Sprintf("invalid rule number %q: %s", ruleNumberStr, err))
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("transit_gateway_metering_policy_id"), policyID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("policy_rule_number"), ruleNumber)...)
}

const transitGatewayMeteringPolicyEntryIDSeparator = "_"

func transitGatewayMeteringPolicyEntryCreateResourceID(policyID, ruleNumber string) string {
	return strings.Join([]string{policyID, ruleNumber}, transitGatewayMeteringPolicyEntryIDSeparator)
}

func transitGatewayMeteringPolicyEntryParseResourceID(id string) (string, string, error) {
	// Policy IDs look like "tgw-mp-12345678" (hyphens, no underscores).
	// Rule numbers are integers. Separator is "_".
	idx := strings.LastIndex(id, transitGatewayMeteringPolicyEntryIDSeparator)
	if idx < 1 || idx >= len(id)-1 {
		return "", "", fmt.Errorf("invalid EC2 Transit Gateway Metering Policy Entry ID: %q, expected POLICY-ID_RULE-NUMBER", id)
	}
	return id[:idx], id[idx+1:], nil
}

// transitGatewayMeteringPolicyEntryFlattenRule sets the traffic-matching rule fields
// in the model from the AWS SDK MeteringPolicyRule struct.
func transitGatewayMeteringPolicyEntryFlattenRule(ctx context.Context, rule *awstypes.TransitGatewayMeteringPolicyRule, data *transitGatewayMeteringPolicyEntryResourceModel) {
	if rule == nil {
		return
	}
	data.DestinationCidrBlock = fwflex.StringToFramework(ctx, rule.DestinationCidrBlock)
	data.DestinationPortRange = fwflex.StringToFramework(ctx, rule.DestinationPortRange)
	data.DestinationTransitGatewayAttachmentID = fwflex.StringToFramework(ctx, rule.DestinationTransitGatewayAttachmentId)
	if rule.DestinationTransitGatewayAttachmentType != "" {
		data.DestinationTransitGatewayAttachmentType = fwtypes.StringEnumValue(rule.DestinationTransitGatewayAttachmentType)
	}
	data.Protocol = fwflex.StringToFramework(ctx, rule.Protocol)
	data.SourceCidrBlock = fwflex.StringToFramework(ctx, rule.SourceCidrBlock)
	data.SourcePortRange = fwflex.StringToFramework(ctx, rule.SourcePortRange)
	data.SourceTransitGatewayAttachmentID = fwflex.StringToFramework(ctx, rule.SourceTransitGatewayAttachmentId)
	if rule.SourceTransitGatewayAttachmentType != "" {
		data.SourceTransitGatewayAttachmentType = fwtypes.StringEnumValue(rule.SourceTransitGatewayAttachmentType)
	}
}

type transitGatewayMeteringPolicyEntryResourceModel struct {
	framework.WithRegionModel
	DestinationCidrBlock                    types.String                                                      `tfsdk:"destination_cidr_block"`
	DestinationPortRange                    types.String                                                      `tfsdk:"destination_port_range"`
	DestinationTransitGatewayAttachmentID   types.String                                                      `tfsdk:"destination_transit_gateway_attachment_id"`
	DestinationTransitGatewayAttachmentType fwtypes.StringEnum[awstypes.TransitGatewayAttachmentResourceType] `tfsdk:"destination_transit_gateway_attachment_type"`
	ID                                      types.String                                                      `tfsdk:"id"`
	MeteredAccount                          fwtypes.StringEnum[awstypes.TransitGatewayMeteringPayerType]      `tfsdk:"metered_account"`
	PolicyRuleNumber                        types.Int64                                                       `tfsdk:"policy_rule_number"`
	Protocol                                types.String                                                      `tfsdk:"protocol"`
	SourceCidrBlock                         types.String                                                      `tfsdk:"source_cidr_block"`
	SourcePortRange                         types.String                                                      `tfsdk:"source_port_range"`
	SourceTransitGatewayAttachmentID        types.String                                                      `tfsdk:"source_transit_gateway_attachment_id"`
	SourceTransitGatewayAttachmentType      fwtypes.StringEnum[awstypes.TransitGatewayAttachmentResourceType] `tfsdk:"source_transit_gateway_attachment_type"`
	State                                   types.String                                                      `tfsdk:"state"`
	Timeouts                                timeouts.Value                                                    `tfsdk:"timeouts"`
	TransitGatewayMeteringPolicyID          types.String                                                      `tfsdk:"transit_gateway_metering_policy_id"`
}
