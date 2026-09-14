// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_transit_gateway_policy_table_entry", name="Transit Gateway Policy Table Entry")
// @IdentityAttribute("transit_gateway_policy_table_id")
// @IdentityAttribute("policy_rule_number")
// @ImportIDHandler("transitGatewayPolicyTableEntryImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.TransitGatewayPolicyTableEntry")
// @Testing(importStateIdFunc="testAccTransitGatewayPolicyTableEntryImportStateIDFunc")
// @Testing(importStateIdAttribute="transit_gateway_policy_table_id")
func newTransitGatewayPolicyTableEntryResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &transitGatewayPolicyTableEntryResource{}
	return r, nil
}

type transitGatewayPolicyTableEntryResource struct {
	framework.ResourceWithModel[transitGatewayPolicyTableEntryResourceModel]
	framework.WithImportByIdentity
}

func (r *transitGatewayPolicyTableEntryResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"policy_rule_number": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_route_table_id": schema.StringAttribute{
				Required: true,
			},
			"transit_gateway_policy_table_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"policy_rule": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[transitGatewayPolicyRuleModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"destination_cidr_block": schema.StringAttribute{
							Optional: true,
						},
						"destination_port_range": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						names.AttrProtocol: schema.StringAttribute{
							Optional: true,
						},
						"source_cidr_block": schema.StringAttribute{
							Optional: true,
						},
						"source_port_range": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"metadata": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[transitGatewayPolicyRuleMetaDataModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Optional: true,
									},
									names.AttrValue: schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *transitGatewayPolicyTableEntryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data transitGatewayPolicyTableEntryResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	policyTableID := fwflex.StringValueFromFramework(ctx, data.TransitGatewayPolicyTableID)
	var input ec2.CreateTransitGatewayPolicyTableEntryInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	var priorMetaData *awstypes.TransitGatewayPolicyRuleMetaData
	if v := input.PolicyRule; v != nil && v.MetaData != nil {
		priorMetaData = new(awstypes.TransitGatewayPolicyRuleMetaData{
			MetaDataKey:   v.MetaData.MetaDataKey,
			MetaDataValue: v.MetaData.MetaDataValue,
		})
	}

	output, err := conn.CreateTransitGatewayPolicyTableEntry(ctx, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyTableID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.TransitGatewayPolicyTableEntry, priorMetaData, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *transitGatewayPolicyTableEntryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data transitGatewayPolicyTableEntryResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	policyTableID, ruleNumber := fwflex.StringValueFromFramework(ctx, data.TransitGatewayPolicyTableID), fwflex.StringValueFromFramework(ctx, data.PolicyRuleNumber)
	entry, err := findTransitGatewayPolicyTableEntryByTwoPartKey(ctx, conn, policyTableID, ruleNumber)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyTableID)
		return
	}

	var priorMetaData *awstypes.TransitGatewayPolicyRuleMetaData
	if !data.PolicyRule.IsNull() {
		var rule awstypes.TransitGatewayPolicyRule
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data.PolicyRule, &rule))
		if response.Diagnostics.HasError() {
			return
		}
		priorMetaData = rule.MetaData
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, entry, priorMetaData, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *transitGatewayPolicyTableEntryResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data transitGatewayPolicyTableEntryResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	policyTableID := fwflex.StringValueFromFramework(ctx, data.TransitGatewayPolicyTableID)
	var input ec2.ModifyTransitGatewayPolicyTableEntryInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	var priorMetaData *awstypes.TransitGatewayPolicyRuleMetaData
	if v := input.PolicyRule; v != nil && v.MetaData != nil {
		priorMetaData = new(awstypes.TransitGatewayPolicyRuleMetaData{
			MetaDataKey:   v.MetaData.MetaDataKey,
			MetaDataValue: v.MetaData.MetaDataValue,
		})
	}

	output, err := conn.ModifyTransitGatewayPolicyTableEntry(ctx, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyTableID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.TransitGatewayPolicyTableEntry, priorMetaData, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *transitGatewayPolicyTableEntryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data transitGatewayPolicyTableEntryResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	policyTableID, ruleNumber := fwflex.StringValueFromFramework(ctx, data.TransitGatewayPolicyTableID), fwflex.StringValueFromFramework(ctx, data.PolicyRuleNumber)
	input := ec2.DeleteTransitGatewayPolicyTableEntryInput{
		PolicyRuleNumber:            aws.String(ruleNumber),
		TransitGatewayPolicyTableId: aws.String(policyTableID),
	}
	_, err := conn.DeleteTransitGatewayPolicyTableEntry(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound, errCodeInvalidTransitGatewayPolicyTableEntryNotFound) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, policyTableID)
		return
	}
}

func (r *transitGatewayPolicyTableEntryResource) flatten(ctx context.Context, entry *awstypes.TransitGatewayPolicyTableEntry, priorMetaData *awstypes.TransitGatewayPolicyRuleMetaData, data *transitGatewayPolicyTableEntryResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Corrects two AWS API quirks:
	//   - AWS always returns a PolicyRule, even when no policy_rule block was configured,
	//     reporting every unset match criterion as "*" (all) rather than omitting it. Such
	//     an all-wildcard rule carries no configuration, so it's restored to the prior
	//     value; returning it verbatim would make Terraform report "block count changed
	//     from 0 to 1" after apply. A rule that wildcards only some criteria (e.g.
	//     protocol = "*" alongside real CIDRs) is a real rule and is kept.
	//   - The API never returns the rule's metadata, so it's restored from the prior value
	//     (unavailable, e.g. on import, it's simply left null).
	if v := entry.PolicyRule; v != nil {
		// transitGatewayPolicyRuleWildcard is how AWS reports a policy rule match criterion
		// that matches everything, including one that was never configured.
		const transitGatewayPolicyRuleWildcard = "*"
		if aws.ToString(v.DestinationCidrBlock) == transitGatewayPolicyRuleWildcard &&
			aws.ToString(v.DestinationPortRange) == transitGatewayPolicyRuleWildcard &&
			aws.ToString(v.Protocol) == transitGatewayPolicyRuleWildcard &&
			aws.ToString(v.SourceCidrBlock) == transitGatewayPolicyRuleWildcard &&
			aws.ToString(v.SourcePortRange) == transitGatewayPolicyRuleWildcard {
			entry.PolicyRule = nil
		} else {
			entry.PolicyRule.MetaData = priorMetaData
		}
	}

	diags.Append(fwflex.Flatten(ctx, entry, data)...)

	return diags
}

const transitGatewayPolicyTableEntryImportIDSeparator = intflex.ResourceIdSeparator

func parseTransitGatewayPolicyTableEntryImportID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayPolicyTableEntryImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected transit-gateway-policy-table-id%[2]spolicy-rule-number", id, transitGatewayPolicyTableEntryImportIDSeparator)
}

var _ inttypes.ImportIDParser = transitGatewayPolicyTableEntryImportID{}

type transitGatewayPolicyTableEntryImportID struct{}

func (transitGatewayPolicyTableEntryImportID) Parse(id string) (string, map[string]any, error) {
	policyTableID, ruleNumber, err := parseTransitGatewayPolicyTableEntryImportID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"policy_rule_number":              ruleNumber,
		"transit_gateway_policy_table_id": policyTableID,
	}

	return id, result, nil
}

type transitGatewayPolicyTableEntryResourceModel struct {
	framework.WithRegionModel
	PolicyRule                  fwtypes.ListNestedObjectValueOf[transitGatewayPolicyRuleModel] `tfsdk:"policy_rule"`
	PolicyRuleNumber            types.String                                                   `tfsdk:"policy_rule_number"`
	TargetRouteTableID          types.String                                                   `tfsdk:"target_route_table_id"`
	TransitGatewayPolicyTableID types.String                                                   `tfsdk:"transit_gateway_policy_table_id"`
}

type transitGatewayPolicyRuleModel struct {
	DestinationCIDRBlock types.String                                                           `tfsdk:"destination_cidr_block"`
	DestinationPortRange types.String                                                           `tfsdk:"destination_port_range"`
	Metadata             fwtypes.ListNestedObjectValueOf[transitGatewayPolicyRuleMetaDataModel] `tfsdk:"metadata"`
	Protocol             types.String                                                           `tfsdk:"protocol"`
	SourceCIDRBlock      types.String                                                           `tfsdk:"source_cidr_block"`
	SourcePortRange      types.String                                                           `tfsdk:"source_port_range"`
}

type transitGatewayPolicyRuleMetaDataModel struct {
	MetaDataKey   types.String `tfsdk:"key"`
	MetaDataValue types.String `tfsdk:"value"`
}
