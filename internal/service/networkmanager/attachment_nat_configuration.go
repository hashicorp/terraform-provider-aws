// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkmanager_attachment_nat_configuration", name="Attachment NAT Configuration")
// @IdentityAttribute("core_network_id")
// @IdentityAttribute("attachment_id")
// @ImportIDHandler("attachmentNATConfigurationImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networkmanager/types;awstypes;awstypes.AttachmentNatConfiguration")
// @Testing(importStateIdFunc=testAccAttachmentNATConfigurationImportStateIdFunc)
// @Testing(importStateIdAttribute="attachment_id")
func newAttachmentNATConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &attachmentNATConfigurationResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type attachmentNATConfigurationResource struct {
	framework.ResourceWithModel[attachmentNATConfigurationResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

type attachmentNATConfigurationResourceModel struct {
	AttachmentID          types.String                                           `tfsdk:"attachment_id"`
	CoreNetworkID         types.String                                           `tfsdk:"core_network_id"`
	NATGatewayRules       fwtypes.SetNestedObjectValueOf[attachmentNATRuleModel] `tfsdk:"nat_gateway_rule"`
	NATGatewaySupport     fwtypes.StringEnum[awstypes.NatGatewaySupportValue]    `tfsdk:"nat_gateway_support"`
	NATSystemRouteSupport types.Bool                                             `tfsdk:"nat_system_route_support"`
	Timeouts              timeouts.Value                                         `tfsdk:"timeouts"`
}

type attachmentNATRuleModel struct {
	MatchSourceCIDR types.String `tfsdk:"match_source_cidr"`
	NATCIDR         types.String `tfsdk:"nat_cidr"`
	RuleNumber      types.Int32  `tfsdk:"rule_number"`
}

func (r *attachmentNATConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"attachment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"core_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nat_gateway_support": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.NatGatewaySupportValue](),
				Required:   true,
			},
			"nat_system_route_support": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"nat_gateway_rule": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[attachmentNATRuleModel](ctx),
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

func (r *attachmentNATConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data attachmentNATConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	coreNetworkID, attachmentID := fwflex.StringValueFromFramework(ctx, data.CoreNetworkID), fwflex.StringValueFromFramework(ctx, data.AttachmentID)
	timeout := r.CreateTimeout(ctx, data.Timeouts)

	if _, err := waitAttachmentAvailable(ctx, conn, coreNetworkID, attachmentID, timeout); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
		return
	}

	var input networkmanager.UpdateAttachmentNatConfigurationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateAttachmentNatConfiguration(ctx, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
		return
	}

	output, err := waitAttachmentNATConfigurationUpdated(ctx, conn, coreNetworkID, attachmentID, input.NatGatewaySupport, input.NatGatewayRules, timeout)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *attachmentNATConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data attachmentNATConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	attachmentID := fwflex.StringValueFromFramework(ctx, data.AttachmentID)
	output, err := findAttachmentNATConfigurationByID(ctx, conn, attachmentID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *attachmentNATConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old attachmentNATConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		coreNetworkID, attachmentID := fwflex.StringValueFromFramework(ctx, new.CoreNetworkID), fwflex.StringValueFromFramework(ctx, new.AttachmentID)
		timeout := r.UpdateTimeout(ctx, new.Timeouts)

		if _, err := waitAttachmentAvailable(ctx, conn, coreNetworkID, attachmentID, timeout); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
			return
		}

		// Send only the changed fields. nat_system_route_support can only be changed on some
		// attachment types, so an unchanged value must not be sent.
		input := networkmanager.UpdateAttachmentNatConfigurationInput{
			AttachmentId: aws.String(attachmentID),
		}
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input, diff.IgnoredFieldNamesOpts()...))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateAttachmentNatConfiguration(ctx, &input)

		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
			return
		}

		var rules []awstypes.NatRuleRequest
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new.NATGatewayRules, &rules))
		if response.Diagnostics.HasError() {
			return
		}

		output, err := waitAttachmentNATConfigurationUpdated(ctx, conn, coreNetworkID, attachmentID, new.NATGatewaySupport.ValueEnum(), rules, timeout)

		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &new))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *attachmentNATConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data attachmentNATConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	coreNetworkID, attachmentID := fwflex.StringValueFromFramework(ctx, data.CoreNetworkID), fwflex.StringValueFromFramework(ctx, data.AttachmentID)
	timeout := r.DeleteTimeout(ctx, data.Timeouts)

	_, err := waitAttachmentAvailable(ctx, conn, coreNetworkID, attachmentID, timeout)

	if retry.NotFound(err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
		return
	}

	// nat_system_route_support isn't sent, so the attachment keeps its last value.
	input := networkmanager.UpdateAttachmentNatConfigurationInput{
		AttachmentId: aws.String(attachmentID),
		// A non-nil empty list removes all rules. A nil list leaves them unchanged.
		NatGatewayRules:   []awstypes.NatRuleRequest{},
		NatGatewaySupport: awstypes.NatGatewaySupportValueDisable,
	}
	_, err = conn.UpdateAttachmentNatConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
		return
	}

	if _, err := waitAttachmentNATConfigurationDeleted(ctx, conn, coreNetworkID, attachmentID, timeout); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, attachmentID)
		return
	}
}

func (r *attachmentNATConfigurationResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var rules fwtypes.SetNestedObjectValueOf[attachmentNATRuleModel]
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.GetAttribute(ctx, path.Root("nat_gateway_rule"), &rules))
	if response.Diagnostics.HasError() || rules.IsNull() || rules.IsUnknown() {
		return
	}

	models, d := rules.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if err := validateAttachmentNATRules(models); err != nil {
		response.Diagnostics.AddAttributeError(path.Root("nat_gateway_rule"), "Invalid NAT gateway rule", err.Error())
	}
}

func (r *attachmentNATConfigurationResource) flatten(ctx context.Context, output *awstypes.AttachmentNatConfiguration, data *attachmentNATConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, output, data, fwflex.WithIgnoredFieldNamesAppend("NatGatewaySupport"))...)
	if diags.HasError() {
		return diags
	}

	// The API returns a support state (ENABLED, ENABLING, ...), while the argument is a value (ENABLE or DISABLE).
	switch output.NatGatewaySupport {
	case awstypes.NatGatewaySupportStateEnabled, awstypes.NatGatewaySupportStateEnabling, awstypes.NatGatewaySupportStatePendingAttachmentAcceptance:
		data.NATGatewaySupport = fwtypes.StringEnumValue(awstypes.NatGatewaySupportValueEnable)
	default:
		data.NATGatewaySupport = fwtypes.StringEnumValue(awstypes.NatGatewaySupportValueDisable)
	}

	return diags
}

var (
	_ inttypes.ImportIDParser = attachmentNATConfigurationImportID{}
)

type attachmentNATConfigurationImportID struct{}

const (
	attachmentNATConfigurationIDParts = 2
)

func (attachmentNATConfigurationImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := intflex.ExpandResourceId(id, attachmentNATConfigurationIDParts, false)
	if err != nil {
		return "", nil, smarterr.NewError(fmt.Errorf("import ID (%s) must be in the format core_network_id,attachment_id: %w", id, err))
	}

	result := map[string]any{
		"attachment_id":   parts[1],
		"core_network_id": parts[0],
	}

	return id, result, nil
}

func findAttachmentNATConfigurationByID(ctx context.Context, conn *networkmanager.Client, attachmentID string) (*awstypes.AttachmentNatConfiguration, error) {
	input := networkmanager.ListAttachmentNatConfigurationsInput{
		AttachmentId: aws.String(attachmentID),
	}
	output, err := findAttachmentNATConfiguration(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	// An attachment that supports NAT can report a configuration even when NAT was never set,
	// so one that is disabled and has no rules is treated as absent.
	if output.NatGatewaySupport == awstypes.NatGatewaySupportStateDisabled && len(output.NatGatewayRules) == 0 {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: "NAT gateway support is disabled and there are no NAT gateway rules",
		})
	}

	return output, nil
}

func findAttachmentNATConfiguration(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListAttachmentNatConfigurationsInput) (*awstypes.AttachmentNatConfiguration, error) {
	output, err := findAttachmentNATConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Only keep the requested attachment.
	if input.AttachmentId != nil {
		output = tfslices.Filter(output, func(v awstypes.AttachmentNatConfiguration) bool {
			return aws.ToString(v.AttachmentId) == aws.ToString(input.AttachmentId)
		})
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findAttachmentNATConfigurations(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListAttachmentNatConfigurationsInput) ([]awstypes.AttachmentNatConfiguration, error) {
	var output []awstypes.AttachmentNatConfiguration

	pages := networkmanager.NewListAttachmentNatConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.AttachmentNatConfigurations...)
	}

	return output, nil
}

// statusAttachmentNATConfiguration reports the attachment state while the attachment isn't settled,
// then the state of any rule in transition, then the NAT gateway support state.
func statusAttachmentNATConfiguration(conn *networkmanager.Client, coreNetworkID, attachmentID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		attachment, err := findAttachmentByTwoPartKey(ctx, conn, coreNetworkID, attachmentID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		input := networkmanager.ListAttachmentNatConfigurationsInput{
			AttachmentId: aws.String(attachmentID),
		}
		output, err := findAttachmentNATConfiguration(ctx, conn, &input)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, rule := range output.NatGatewayRules {
			if rule.State == awstypes.NatRuleStateFailed {
				return nil, "", smarterr.NewError(errors.Join(fmt.Errorf("NAT gateway rule (%d) failed", aws.ToInt32(rule.RuleNumber)), attachmentsError(attachment.LastModificationErrors)))
			}
		}

		switch attachment.State {
		case awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingAttachmentAcceptance, awstypes.AttachmentStatePendingTagAcceptance:
		default:
			return output, string(attachment.State), nil
		}

		for _, rule := range output.NatGatewayRules {
			switch rule.State {
			case awstypes.NatRuleStateCreating, awstypes.NatRuleStateModifying, awstypes.NatRuleStateDeleting, awstypes.NatRuleStateActivating, awstypes.NatRuleStateDeactivating:
				return output, string(rule.State), nil
			}
		}

		return output, string(output.NatGatewaySupport), nil
	}
}

func waitAttachmentNATConfigurationUpdated(ctx context.Context, conn *networkmanager.Client, coreNetworkID, attachmentID string, support awstypes.NatGatewaySupportValue, rules []awstypes.NatRuleRequest, timeout time.Duration) (*awstypes.AttachmentNatConfiguration, error) {
	target, other := awstypes.NatGatewaySupportStateEnabled, awstypes.NatGatewaySupportStateDisabled
	if support == awstypes.NatGatewaySupportValueDisable {
		target, other = other, target
	}
	targets := enum.Slice(target, awstypes.NatGatewaySupportStatePendingAttachmentAcceptance)

	refresh := statusAttachmentNATConfiguration(conn, coreNetworkID, attachmentID)
	stateConf := &retry.StateChangeConf{
		Pending: attachmentNATConfigurationPendingStates(other, awstypes.NatGatewaySupportStateEnabling, awstypes.NatGatewaySupportStateDisabling),
		Target:  targets,
		Refresh: func(ctx context.Context) (any, string, error) {
			outputRaw, status, err := refresh(ctx)

			// Updates are asynchronous and List can briefly report the previous rule set in a stable state.
			if output, ok := outputRaw.(*awstypes.AttachmentNatConfiguration); ok && slices.Contains(targets, status) && !attachmentNATRulesMatch(output.NatGatewayRules, rules) {
				return output, string(awstypes.NatRuleStateModifying), nil
			}

			return outputRaw, status, err
		},
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AttachmentNatConfiguration); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAttachmentNATConfigurationDeleted(ctx context.Context, conn *networkmanager.Client, coreNetworkID, attachmentID string, timeout time.Duration) (*awstypes.AttachmentNatConfiguration, error) {
	refresh := statusAttachmentNATConfiguration(conn, coreNetworkID, attachmentID)
	stateConf := &retry.StateChangeConf{
		Pending: attachmentNATConfigurationPendingStates(enum.EnumValues[awstypes.NatGatewaySupportState]()...),
		Target:  []string{},
		Refresh: func(ctx context.Context) (any, string, error) {
			outputRaw, status, err := refresh(ctx)

			if output, ok := outputRaw.(*awstypes.AttachmentNatConfiguration); ok && status == string(awstypes.NatGatewaySupportStateDisabled) && len(output.NatGatewayRules) == 0 {
				return nil, "", nil
			}

			return outputRaw, status, err
		},
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AttachmentNatConfiguration); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// attachmentNATConfigurationPendingStates returns the given support states plus the attachment and rule states in which a NAT change is still being applied.
func attachmentNATConfigurationPendingStates(support ...awstypes.NatGatewaySupportState) []string {
	return slices.Concat(
		enum.Slice(support...),
		enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStateUpdating, awstypes.AttachmentStatePendingNetworkUpdate),
		enum.Slice(awstypes.NatRuleStateCreating, awstypes.NatRuleStateModifying, awstypes.NatRuleStateDeleting, awstypes.NatRuleStateActivating, awstypes.NatRuleStateDeactivating),
	)
}

// attachmentNATRulesMatch reports whether got contains exactly the rules in want, ignoring rule state.
func attachmentNATRulesMatch(got []awstypes.NatRule, want []awstypes.NatRuleRequest) bool {
	if len(got) != len(want) {
		return false
	}

	rules := make(map[int32]awstypes.NatRule, len(got))
	for _, rule := range got {
		rules[aws.ToInt32(rule.RuleNumber)] = rule
	}

	for _, w := range want {
		rule, ok := rules[aws.ToInt32(w.RuleNumber)]
		if !ok || aws.ToString(rule.MatchSourceCidr) != aws.ToString(w.MatchSourceCidr) || aws.ToString(rule.NatCidr) != aws.ToString(w.NatCidr) {
			return false
		}
	}

	return true
}

// validateAttachmentNATRules checks constraints across NAT gateway rules that attribute validators can't express.
// Unknown, null, and unparsable values are skipped; attribute validators and the API report those.
func validateAttachmentNATRules(rules []*attachmentNATRuleModel) error {
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
