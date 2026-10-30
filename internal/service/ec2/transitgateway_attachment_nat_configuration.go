// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_transit_gateway_attachment_nat_configuration", name="Transit Gateway Attachment NAT Configuration")
// @IdentityAttribute("transit_gateway_attachment_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2;ec2.GetTransitGatewayAttachmentNatConfigurationOutput")
// @Testing(importStateIdAttribute="transit_gateway_attachment_id")
// @Testing(preCheck="testAccPreCheckTransitGateway")
func newTransitGatewayAttachmentNATConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &transitGatewayAttachmentNATConfigurationResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type transitGatewayAttachmentNATConfigurationResource struct {
	framework.ResourceWithModel[transitGatewayAttachmentNATConfigurationResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *transitGatewayAttachmentNATConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"nat_gateway_support": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TransitGatewayAttachmentNatGatewaySupportValue](),
				Required:   true,
			},
			"nat_system_route_support": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TransitGatewayAttachmentNatSystemRouteSupportValue](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTransitGatewayAttachmentID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"nat_gateway_rule": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[transitGatewayNATRuleModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"match_source_cidr": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								fwvalidators.IPv4CIDRNetworkAddress(),
							},
						},
						"nat_cidr": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								fwvalidators.IPv4CIDRNetworkAddress(),
							},
						},
						"rule_number": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.Between(1, 32766),
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

func (r *transitGatewayAttachmentNATConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data transitGatewayAttachmentNATConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.TransitGatewayAttachmentID)
	var input ec2.ModifyTransitGatewayAttachmentNatConfigurationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.ModifyTransitGatewayAttachmentNatConfiguration(ctx, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	output, err := waitTransitGatewayAttachmentNATConfigurationUpdated(ctx, conn, id, input.NatGatewaySupport, input.NatGatewayRules, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *transitGatewayAttachmentNATConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data transitGatewayAttachmentNATConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.TransitGatewayAttachmentID)
	output, err := findTransitGatewayAttachmentNATConfigurationByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
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

func (r *transitGatewayAttachmentNATConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old transitGatewayAttachmentNATConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		id := fwflex.StringValueFromFramework(ctx, new.TransitGatewayAttachmentID)

		// Send only the changed fields. In particular, nat_system_route_support can only be
		// changed on some attachment types, so an unchanged value must not be sent.
		input := ec2.ModifyTransitGatewayAttachmentNatConfigurationInput{
			TransitGatewayAttachmentId: aws.String(id),
		}
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input, diff.IgnoredFieldNamesOpts()...))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.ModifyTransitGatewayAttachmentNatConfiguration(ctx, &input)

		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		var rules []awstypes.TransitGatewayNatRuleRequest
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new.NATGatewayRules, &rules))
		if response.Diagnostics.HasError() {
			return
		}

		output, err := waitTransitGatewayAttachmentNATConfigurationUpdated(ctx, conn, id, new.NATGatewaySupport.ValueEnum(), rules, r.UpdateTimeout(ctx, new.Timeouts))

		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &new))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *transitGatewayAttachmentNATConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data transitGatewayAttachmentNATConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	// nat_system_route_support is left unchanged, as in CloudFormation.
	id := fwflex.StringValueFromFramework(ctx, data.TransitGatewayAttachmentID)
	input := ec2.ModifyTransitGatewayAttachmentNatConfigurationInput{
		NatGatewaySupport:          awstypes.TransitGatewayAttachmentNatGatewaySupportValueDisable,
		ResetNatGatewayRules:       aws.Bool(true),
		TransitGatewayAttachmentId: aws.String(id),
	}
	_, err := conn.ModifyTransitGatewayAttachmentNatConfiguration(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if _, err := waitTransitGatewayAttachmentNATConfigurationDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

func (r *transitGatewayAttachmentNATConfigurationResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var rules fwtypes.SetNestedObjectValueOf[transitGatewayNATRuleModel]
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.GetAttribute(ctx, path.Root("nat_gateway_rule"), &rules))
	if response.Diagnostics.HasError() || rules.IsNull() || rules.IsUnknown() {
		return
	}

	models, d := rules.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if err := validateTransitGatewayNATRules(models); err != nil {
		response.Diagnostics.AddAttributeError(path.Root("nat_gateway_rule"), "Invalid NAT gateway rule", err.Error())
	}
}

func (r *transitGatewayAttachmentNATConfigurationResource) flatten(ctx context.Context, output *ec2.GetTransitGatewayAttachmentNatConfigurationOutput, data *transitGatewayAttachmentNATConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, output, data, fwflex.WithIgnoredFieldNamesAppend("NatGatewaySupport"))...)
	if diags.HasError() {
		return diags
	}

	// The API returns a support state (enabled, enabling, ...), while the argument is a value (enable or disable).
	switch output.NatGatewaySupport {
	case awstypes.TransitGatewayAttachmentNatGatewaySupportStateEnabled, awstypes.TransitGatewayAttachmentNatGatewaySupportStateEnabling:
		data.NATGatewaySupport = fwtypes.StringEnumValue(awstypes.TransitGatewayAttachmentNatGatewaySupportValueEnable)
	default:
		data.NATGatewaySupport = fwtypes.StringEnumValue(awstypes.TransitGatewayAttachmentNatGatewaySupportValueDisable)
	}

	return diags
}

// validateTransitGatewayNATRules checks constraints across NAT gateway rules that attribute validators can't express.
// Unknown, null, and unparsable values are skipped; attribute validators and the API report those.
func validateTransitGatewayNATRules(rules []*transitGatewayNATRuleModel) error {
	var errs []error
	ruleNumbers := make(map[int32]bool, len(rules))

	for _, rule := range rules {
		if !rule.RuleNumber.IsNull() && !rule.RuleNumber.IsUnknown() {
			n := rule.RuleNumber.ValueInt32()
			if ruleNumbers[n] {
				errs = append(errs, fmt.Errorf("rule_number %d is used by more than one rule", n))
			}
			ruleNumbers[n] = true
		}

		if rule.MatchSourceCIDR.IsNull() || rule.MatchSourceCIDR.IsUnknown() || rule.NATCIDR.IsNull() || rule.NATCIDR.IsUnknown() {
			continue
		}

		source, nat := rule.MatchSourceCIDR.ValueString(), rule.NATCIDR.ValueString()
		sourcePrefix, err := netip.ParsePrefix(source)
		if err != nil {
			continue
		}
		natPrefix, err := netip.ParsePrefix(nat)
		if err != nil {
			continue
		}

		if sourcePrefix.Bits() != natPrefix.Bits() {
			errs = append(errs, fmt.Errorf("nat_cidr (%s) must have the same prefix length as match_source_cidr (%s)", nat, source))
		}
	}

	return errors.Join(errs...)
}

type transitGatewayAttachmentNATConfigurationResourceModel struct {
	framework.WithRegionModel
	NATGatewayRules            fwtypes.SetNestedObjectValueOf[transitGatewayNATRuleModel]                      `tfsdk:"nat_gateway_rule"`
	NATGatewaySupport          fwtypes.StringEnum[awstypes.TransitGatewayAttachmentNatGatewaySupportValue]     `tfsdk:"nat_gateway_support"`
	NATSystemRouteSupport      fwtypes.StringEnum[awstypes.TransitGatewayAttachmentNatSystemRouteSupportValue] `tfsdk:"nat_system_route_support"`
	Timeouts                   timeouts.Value                                                                  `tfsdk:"timeouts"`
	TransitGatewayAttachmentID types.String                                                                    `tfsdk:"transit_gateway_attachment_id"`
}

type transitGatewayNATRuleModel struct {
	MatchSourceCIDR types.String `tfsdk:"match_source_cidr"`
	NATCIDR         types.String `tfsdk:"nat_cidr"`
	RuleNumber      types.Int32  `tfsdk:"rule_number"`
}
