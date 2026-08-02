// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
// @Testing(importStateIdFunc="testAccTransitGatewayPolicyTableEntryImportState")
// @Testing(importStateIdAttributes="transit_gateway_policy_table_id;policy_rule_number", importStateIdAttributesSep="intflex.ResourceIdSeparator")
func newTransitGatewayPolicyTableEntryResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &transitGatewayPolicyTableEntryResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type transitGatewayPolicyTableEntryResource struct {
	framework.ResourceWithModel[transitGatewayPolicyTableEntryResourceModel]
	framework.WithTimeouts
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
			names.AttrState: schema.StringAttribute{
				Computed: true,
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
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

	var input ec2.CreateTransitGatewayPolicyTableEntryInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateTransitGatewayPolicyTableEntry(ctx, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.TransitGatewayPolicyTableID.ValueString())
		return
	}

	planPolicyRuleNull := data.PolicyRule.IsNull()
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output.TransitGatewayPolicyTableEntry, &data))
	if response.Diagnostics.HasError() {
		return
	}
	if planPolicyRuleNull {
		data.PolicyRule = fwtypes.NewListNestedObjectValueOfNull[transitGatewayPolicyRuleModel](ctx)
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

	statePolicyRuleNull := data.PolicyRule.IsNull()
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, entry, &data))
	if response.Diagnostics.HasError() {
		return
	}
	if statePolicyRuleNull {
		data.PolicyRule = fwtypes.NewListNestedObjectValueOfNull[transitGatewayPolicyRuleModel](ctx)
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

	var input ec2.ModifyTransitGatewayPolicyTableEntryInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.ModifyTransitGatewayPolicyTableEntry(ctx, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.TransitGatewayPolicyTableID.ValueString())
		return
	}

	planPolicyRuleNull := data.PolicyRule.IsNull()
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output.TransitGatewayPolicyTableEntry, &data))
	if response.Diagnostics.HasError() {
		return
	}
	if planPolicyRuleNull {
		data.PolicyRule = fwtypes.NewListNestedObjectValueOfNull[transitGatewayPolicyRuleModel](ctx)
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

var _ inttypes.ImportIDParser = transitGatewayPolicyTableEntryImportID{}

type transitGatewayPolicyTableEntryImportID struct{}

func (transitGatewayPolicyTableEntryImportID) Parse(id string) (string, map[string]any, error) {
	policyTableID, ruleNumber, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id \"%s\" should be in the format <transit-gateway-policy-table-id>"+intflex.ResourceIdSeparator+"<policy-rule-number>", id)
	}

	result := map[string]any{
		"transit_gateway_policy_table_id": policyTableID,
		"policy_rule_number":              ruleNumber,
	}

	return id, result, nil
}

type transitGatewayPolicyTableEntryResourceModel struct {
	framework.WithRegionModel
	PolicyRule                  fwtypes.ListNestedObjectValueOf[transitGatewayPolicyRuleModel] `tfsdk:"policy_rule"`
	PolicyRuleNumber            types.String                                                   `tfsdk:"policy_rule_number"`
	State                       types.String                                                   `tfsdk:"state"`
	TargetRouteTableID          types.String                                                   `tfsdk:"target_route_table_id"`
	Timeouts                    timeouts.Value                                                 `tfsdk:"timeouts"`
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
