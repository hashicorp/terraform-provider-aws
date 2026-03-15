// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfretry "github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive", name="Proxy Configuration Rule Group Attachments Exclusive")
// @ArnIdentity("proxy_configuration_arn",identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preIdentityVersion="v5.100.0")
func newResourceProxyConfigurationRuleGroupAttachmentsExclusive(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProxyConfigurationRuleGroupAttachmentsExclusive{}

	return r, nil
}

type resourceProxyConfigurationRuleGroupAttachmentsExclusive struct {
	framework.ResourceWithModel[proxyConfigurationRuleGroupAttachmentModel]
	framework.WithImportByIdentity
}

func (r *resourceProxyConfigurationRuleGroupAttachmentsExclusive) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			"proxy_configuration_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"update_token": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"rule_group": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[RuleGroupAttachmentModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"proxy_rule_group_name": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceProxyConfigurationRuleGroupAttachmentsExclusive) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan proxyConfigurationRuleGroupAttachmentModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	proxyConfigArn := plan.ProxyConfigurationArn.ValueString()

	// First, get the current update token
	out, err := findProxyConfigurationByARN(ctx, conn, proxyConfigArn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
		return
	}

	// Build the list of rule groups to attach with InsertPosition based on order
	var ruleGroups []awstypes.ProxyRuleGroupAttachment
	planRuleGroups, d := plan.RuleGroups.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, rg := range planRuleGroups {
		ruleGroups = append(ruleGroups, awstypes.ProxyRuleGroupAttachment{
			ProxyRuleGroupName: aws.String(rg.ProxyRuleGroupName.ValueString()),
			InsertPosition:     aws.Int32(int32(i)),
		})
	}

	input := &networkfirewall.AttachRuleGroupsToProxyConfigurationInput{
		ProxyConfigurationArn: aws.String(proxyConfigArn),
		RuleGroups:            ruleGroups,
		UpdateToken:           out.UpdateToken,
	}

	_, err = conn.AttachRuleGroupsToProxyConfiguration(ctx, input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
		return
	}

	plan.setID()

	// Read back the update token from AWS
	refreshedOut, err := findProxyConfigurationByARN(ctx, conn, proxyConfigArn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
		return
	}

	// Set the update token from the refreshed state, but keep the plan's rule group order
	// since we've just applied that order via the API
	plan.UpdateToken = flex.StringToFramework(ctx, refreshedOut.UpdateToken)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceProxyConfigurationRuleGroupAttachmentsExclusive) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state proxyConfigurationRuleGroupAttachmentModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	proxyConfigArn := state.ID.ValueString()

	out, err := findProxyConfigurationByARN(ctx, conn, proxyConfigArn)
	if tfretry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
		return
	}

	// If there are no rule groups attached, the attachment resource no longer exists
	if out.ProxyConfiguration == nil || out.ProxyConfiguration.RuleGroups == nil || len(out.ProxyConfiguration.RuleGroups) == 0 {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(errors.New("no rule groups attached to proxy configuration")))
		resp.State.RemoveResource(ctx)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenProxyConfigurationRuleGroups(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceProxyConfigurationRuleGroupAttachmentsExclusive) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan, state proxyConfigurationRuleGroupAttachmentModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	proxyConfigArn := state.ID.ValueString()

	// Get current update token from AWS (required for attach/detach operations)
	out, err := findProxyConfigurationByARN(ctx, conn, proxyConfigArn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
		return
	}

	updateToken := out.UpdateToken

	// Get the current rule groups from Terraform state
	stateRuleGroups, d := state.RuleGroups.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the planned rule groups
	planRuleGroups, d := plan.RuleGroups.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build maps for comparison
	stateRuleGroupNames := make(map[string]bool)
	for _, rg := range stateRuleGroups {
		stateRuleGroupNames[rg.ProxyRuleGroupName.ValueString()] = true
	}

	planRuleGroupNames := make(map[string]bool)
	for _, rg := range planRuleGroups {
		planRuleGroupNames[rg.ProxyRuleGroupName.ValueString()] = true
	}

	// Detach rule groups that are in state but not in plan
	var ruleGroupsToDetach []string
	for name := range stateRuleGroupNames {
		if !planRuleGroupNames[name] {
			ruleGroupsToDetach = append(ruleGroupsToDetach, name)
		}
	}

	if len(ruleGroupsToDetach) > 0 {
		detachInput := &networkfirewall.DetachRuleGroupsFromProxyConfigurationInput{
			ProxyConfigurationArn: aws.String(proxyConfigArn),
			RuleGroupNames:        ruleGroupsToDetach,
			UpdateToken:           updateToken,
		}

		detachOut, err := conn.DetachRuleGroupsFromProxyConfiguration(ctx, detachInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
			return
		}
		updateToken = detachOut.UpdateToken
	}

	// Attach rule groups that are in plan but not in state
	var ruleGroupsToAttach []awstypes.ProxyRuleGroupAttachment
	for i, rg := range planRuleGroups {
		if !stateRuleGroupNames[rg.ProxyRuleGroupName.ValueString()] {
			ruleGroupsToAttach = append(ruleGroupsToAttach, awstypes.ProxyRuleGroupAttachment{
				ProxyRuleGroupName: aws.String(rg.ProxyRuleGroupName.ValueString()),
				InsertPosition:     aws.Int32(int32(i)),
			})
		}
	}

	if len(ruleGroupsToAttach) > 0 {
		attachInput := &networkfirewall.AttachRuleGroupsToProxyConfigurationInput{
			ProxyConfigurationArn: aws.String(proxyConfigArn),
			RuleGroups:            ruleGroupsToAttach,
			UpdateToken:           updateToken,
		}

		attachOut, err := conn.AttachRuleGroupsToProxyConfiguration(ctx, attachInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
			return
		}
		updateToken = attachOut.UpdateToken
	}

	// Check if reordering is needed (same groups but different order)
	needsReorder := false
	if len(stateRuleGroups) == len(planRuleGroups) && len(ruleGroupsToAttach) == 0 && len(ruleGroupsToDetach) == 0 {
		// Same groups exist, check if order has changed
		for i := range planRuleGroups {
			if planRuleGroups[i].ProxyRuleGroupName.ValueString() != stateRuleGroups[i].ProxyRuleGroupName.ValueString() {
				needsReorder = true
				break
			}
		}
	}

	if needsReorder {
		var ruleGroupPriorities []awstypes.ProxyRuleGroupPriority
		for i, rg := range planRuleGroups {
			ruleGroupPriorities = append(ruleGroupPriorities, awstypes.ProxyRuleGroupPriority{
				ProxyRuleGroupName: aws.String(rg.ProxyRuleGroupName.ValueString()),
				NewPosition:        aws.Int32(int32(i)),
			})
		}

		priorityInput := &networkfirewall.UpdateProxyRuleGroupPrioritiesInput{
			ProxyConfigurationArn: aws.String(proxyConfigArn),
			RuleGroups:            ruleGroupPriorities,
			UpdateToken:           updateToken,
		}

		priorityOut, err := conn.UpdateProxyRuleGroupPriorities(ctx, priorityInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
			return
		}
		updateToken = priorityOut.UpdateToken
	}

	// Read back the update token from AWS
	refreshedOut, err := findProxyConfigurationByARN(ctx, conn, proxyConfigArn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
		return
	}

	// Set the update token from the refreshed state, but keep the plan's rule group order
	// since we've just applied that order via the API
	plan.UpdateToken = flex.StringToFramework(ctx, refreshedOut.UpdateToken)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceProxyConfigurationRuleGroupAttachmentsExclusive) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state proxyConfigurationRuleGroupAttachmentModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	proxyConfigArn := state.ID.ValueString()

	// Get current state to retrieve update token
	out, err := findProxyConfigurationByARN(ctx, conn, proxyConfigArn)
	if tfretry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
		return
	}

	// Get all rule groups to detach
	stateRuleGroups, d := state.RuleGroups.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(stateRuleGroups) == 0 {
		return
	}

	var ruleGroupNames []string
	for _, rg := range stateRuleGroups {
		ruleGroupNames = append(ruleGroupNames, rg.ProxyRuleGroupName.ValueString())
	}

	input := &networkfirewall.DetachRuleGroupsFromProxyConfigurationInput{
		ProxyConfigurationArn: aws.String(proxyConfigArn),
		RuleGroupNames:        ruleGroupNames,
		UpdateToken:           out.UpdateToken,
	}

	_, err = conn.DetachRuleGroupsFromProxyConfiguration(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		// Ignore error if rule groups are already detached (specific InvalidRequestException message)
		var invalidRequestErr *awstypes.InvalidRequestException
		if errors.As(err, &invalidRequestErr) && invalidRequestErr.Message != nil {
			if strings.Contains(*invalidRequestErr.Message, "not currently attached") {
				return
			}
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, proxyConfigArn)
		return
	}
}

type proxyConfigurationRuleGroupAttachmentModel struct {
	framework.WithRegionModel
	ID                    types.String                                              `tfsdk:"id"`
	ProxyConfigurationArn fwtypes.ARN                                               `tfsdk:"proxy_configuration_arn"`
	RuleGroups            fwtypes.ListNestedObjectValueOf[RuleGroupAttachmentModel] `tfsdk:"rule_group"`
	UpdateToken           types.String                                              `tfsdk:"update_token"`
}

// RuleGroupAttachmentModel is exported for use in tests
type RuleGroupAttachmentModel struct {
	ProxyRuleGroupName types.String `tfsdk:"proxy_rule_group_name"`
}

func (data *proxyConfigurationRuleGroupAttachmentModel) setID() {
	data.ID = data.ProxyConfigurationArn.StringValue
}

func flattenProxyConfigurationRuleGroups(ctx context.Context, out *networkfirewall.DescribeProxyConfigurationOutput, model *proxyConfigurationRuleGroupAttachmentModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if out.ProxyConfiguration == nil || out.ProxyConfiguration.RuleGroups == nil {
		model.RuleGroups = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []RuleGroupAttachmentModel{})
		model.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)
		return diags
	}

	// Sort by Priority to maintain order (lower priority number = higher priority = first in list)
	sortedRuleGroups := make([]awstypes.ProxyConfigRuleGroup, len(out.ProxyConfiguration.RuleGroups))
	copy(sortedRuleGroups, out.ProxyConfiguration.RuleGroups)
	sort.SliceStable(sortedRuleGroups, func(i, j int) bool {
		return aws.ToInt32(sortedRuleGroups[i].Priority) < aws.ToInt32(sortedRuleGroups[j].Priority)
	})

	var ruleGroups []RuleGroupAttachmentModel
	for _, rg := range sortedRuleGroups {
		ruleGroups = append(ruleGroups, RuleGroupAttachmentModel{
			ProxyRuleGroupName: flex.StringToFramework(ctx, rg.ProxyRuleGroupName),
		})
	}

	model.RuleGroups = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, ruleGroups)
	model.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)

	return diags
}
