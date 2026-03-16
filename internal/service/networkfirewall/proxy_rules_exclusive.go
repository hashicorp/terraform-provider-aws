// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"errors"

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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_proxy_rules_exclusive", name="Proxy Rules Exclusive")
// @ArnIdentity("proxy_rule_group_arn",identityDuplicateAttributes="id")
// @ArnFormat("proxy-rule-group/{name}")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preIdentityVersion="v5.100.0")
func newResourceProxyRulesExclusive(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProxyRulesExclusive{}

	return r, nil
}

type resourceProxyRulesExclusive struct {
	framework.ResourceWithModel[resourceProxyRulesExclusiveModel]
	framework.WithImportByIdentity
}

func (r *resourceProxyRulesExclusive) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root("proxy_rule_group_arn")),
			"proxy_rule_group_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"post_response": proxyRuleSchemaBlock(ctx),
			"pre_dns":       proxyRuleSchemaBlock(ctx),
			"pre_request":   proxyRuleSchemaBlock(ctx),
		},
	}
}

func proxyRuleSchemaBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[proxyRuleModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrAction: schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.ProxyRulePhaseAction](),
					Required:   true,
				},
				names.AttrDescription: schema.StringAttribute{
					Optional: true,
				},
				"proxy_rule_name": schema.StringAttribute{
					Required: true,
				},
			},
			Blocks: map[string]schema.Block{
				"conditions": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[proxyRuleConditionModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"condition_key": schema.StringAttribute{
								Required: true,
							},
							"condition_operator": schema.StringAttribute{
								Required: true,
							},
							"condition_values": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								ElementType: types.StringType,
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceProxyRulesExclusive) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan resourceProxyRulesExclusiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := networkfirewall.CreateProxyRulesInput{
		ProxyRuleGroupArn: plan.ProxyRuleGroupArn.ValueStringPointer(),
	}

	var rulesByPhase awstypes.CreateProxyRulesByRequestPhase

	rulesByPhase.PostRESPONSE = proxyExpandRulesForPhase(ctx, plan.PostRESPONSE, &resp.Diagnostics)
	rulesByPhase.PreDNS = proxyExpandRulesForPhase(ctx, plan.PreDNS, &resp.Diagnostics)
	rulesByPhase.PreREQUEST = proxyExpandRulesForPhase(ctx, plan.PreREQUEST, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	input.Rules = &rulesByPhase

	out, err := conn.CreateProxyRules(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ProxyRuleGroupArn.String())
		return
	}
	if out == nil || out.ProxyRuleGroup == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ProxyRuleGroupArn.String())
		return
	}

	plan.setID()

	readOut, err := findProxyRulesByGroupARN(ctx, conn, plan.ProxyRuleGroupArn.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ProxyRuleGroupArn.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, setProxyRulesState(ctx, readOut, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceProxyRulesExclusive) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceProxyRulesExclusiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findProxyRulesByGroupARN(ctx, conn, state.ProxyRuleGroupArn.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ProxyRuleGroupArn.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, setProxyRulesState(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceProxyRulesExclusive) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan, state resourceProxyRulesExclusiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state to obtain update token and existing rules from AWS
	currentRules, err := findProxyRulesByGroupARN(ctx, conn, state.ProxyRuleGroupArn.ValueString())
	if err != nil && !retry.NotFound(err) {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	var updateToken *string

	// Extract rules from AWS for each phase
	var statePostRESPONSE, statePreDNS, statePreREQUEST []proxyRuleModel
	if currentRules != nil && currentRules.ProxyRuleGroup != nil && currentRules.ProxyRuleGroup.Rules != nil {
		rules := currentRules.ProxyRuleGroup.Rules
		statePostRESPONSE = proxyExtractRulesFromPhase(ctx, rules.PostRESPONSE, &resp.Diagnostics)
		statePreDNS = proxyExtractRulesFromPhase(ctx, rules.PreDNS, &resp.Diagnostics)
		statePreREQUEST = proxyExtractRulesFromPhase(ctx, rules.PreREQUEST, &resp.Diagnostics)
	}

	// Extract plan rules for each phase
	planPostRESPONSE := proxyExtractPlanRules(ctx, plan.PostRESPONSE, &resp.Diagnostics)
	planPreDNS := proxyExtractPlanRules(ctx, plan.PreDNS, &resp.Diagnostics)
	planPreREQUEST := proxyExtractPlanRules(ctx, plan.PreREQUEST, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	// Track rules to delete, update, and create
	// Using map to avoid duplicates since we track by name
	rulesToDelete := make(map[string]bool)
	rulesToUpdate := make(map[string]proxyRuleModel)
	rulesToRecreate := make(map[string]proxyRuleRecreateInfo)

	// Process each phase
	processProxyRulesPhaseChanges(ctx, statePostRESPONSE, planPostRESPONSE, "PostRESPONSE", rulesToDelete, rulesToUpdate, rulesToRecreate)
	processProxyRulesPhaseChanges(ctx, statePreDNS, planPreDNS, "PreDNS", rulesToDelete, rulesToUpdate, rulesToRecreate)
	processProxyRulesPhaseChanges(ctx, statePreREQUEST, planPreREQUEST, "PreREQUEST", rulesToDelete, rulesToUpdate, rulesToRecreate)

	// Remove any rules from the update list that are being recreated
	// This ensures we don't try to update a rule that's being deleted
	for name := range rulesToRecreate {
		delete(rulesToUpdate, name)
	}

	// Step 1: Delete rules (including those being recreated)
	// This must happen before Step 3 to satisfy proxy_rule_name uniqueness constraint
	var deleteList []string
	for name := range rulesToDelete {
		deleteList = append(deleteList, name)
	}

	if len(deleteList) > 0 {
		deleteInput := networkfirewall.DeleteProxyRulesInput{
			ProxyRuleGroupArn: plan.ProxyRuleGroupArn.ValueStringPointer(),
			Rules:             deleteList,
		}

		_, err = conn.DeleteProxyRules(ctx, &deleteInput)
		if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}

		// Refresh update token after deletion and verify rules were deleted
		currentRules, err = findProxyRulesByGroupARN(ctx, conn, plan.ProxyRuleGroupArn.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		// Verify that the deleted rules are actually gone
		if currentRules.ProxyRuleGroup != nil && currentRules.ProxyRuleGroup.Rules != nil {
			existingRuleNames := proxyCollectRuleNames(currentRules.ProxyRuleGroup.Rules)

			for _, deletedName := range deleteList {
				if existingRuleNames[deletedName] {
					smerr.AddError(ctx, &resp.Diagnostics, errors.New("rule deletion not yet complete"), smerr.ID, plan.ID.String())
					return
				}
			}
		}
	}

	// Step 2: Update rules (action, description, and conditions)
	for _, ruleModel := range rulesToUpdate {
		// Refresh the update token before each update
		currentRules, err = findProxyRulesByGroupARN(ctx, conn, plan.ProxyRuleGroupArn.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		updateToken = currentRules.UpdateToken

		// Build a map of current rules by name to get old conditions
		currentRulesByName := proxyBuildRuleMapByName(ctx, currentRules.ProxyRuleGroup, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		updateInput := networkfirewall.UpdateProxyRuleInput{
			ProxyRuleGroupArn: plan.ProxyRuleGroupArn.ValueStringPointer(),
			ProxyRuleName:     ruleModel.ProxyRuleName.ValueStringPointer(),
			UpdateToken:       updateToken,
		}

		// Update action if specified
		if !ruleModel.Action.IsNull() && !ruleModel.Action.IsUnknown() {
			updateInput.Action = ruleModel.Action.ValueEnum()
		}

		// Update description if specified
		if !ruleModel.Description.IsNull() && !ruleModel.Description.IsUnknown() {
			updateInput.Description = ruleModel.Description.ValueStringPointer()
		}

		// Remove old conditions
		if currentRule, exists := currentRulesByName[ruleModel.ProxyRuleName.ValueString()]; exists {
			if !currentRule.Conditions.IsNull() && !currentRule.Conditions.IsUnknown() {
				var oldConditions []proxyRuleConditionModel
				smerr.AddEnrich(ctx, &resp.Diagnostics, currentRule.Conditions.ElementsAs(ctx, &oldConditions, false))
				for _, cond := range oldConditions {
					var removeCondition awstypes.ProxyRuleCondition
					smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, cond, &removeCondition))
					updateInput.RemoveConditions = append(updateInput.RemoveConditions, removeCondition)
				}
			}
		}

		// Add new conditions
		if !ruleModel.Conditions.IsNull() && !ruleModel.Conditions.IsUnknown() {
			var newConditions []proxyRuleConditionModel
			smerr.AddEnrich(ctx, &resp.Diagnostics, ruleModel.Conditions.ElementsAs(ctx, &newConditions, false))
			for _, cond := range newConditions {
				var addCondition awstypes.ProxyRuleCondition
				smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, cond, &addCondition))
				updateInput.AddConditions = append(updateInput.AddConditions, addCondition)
			}
		}

		if resp.Diagnostics.HasError() {
			return
		}

		_, err = conn.UpdateProxyRule(ctx, &updateInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	// Step 3: Create/recreate rules
	// Safe to create rules now - all deletions completed in Step 1, ensuring proxy_rule_name uniqueness
	if len(rulesToRecreate) > 0 {
		var rulesByPhase awstypes.CreateProxyRulesByRequestPhase

		// Organize rules by phase
		for _, ruleData := range rulesToRecreate {
			var rule awstypes.CreateProxyRule
			smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, ruleData.rule, &rule))
			if resp.Diagnostics.HasError() {
				return
			}
			rule.InsertPosition = &ruleData.position

			switch ruleData.phase {
			case "PostRESPONSE":
				rulesByPhase.PostRESPONSE = append(rulesByPhase.PostRESPONSE, rule)
			case "PreDNS":
				rulesByPhase.PreDNS = append(rulesByPhase.PreDNS, rule)
			case "PreREQUEST":
				rulesByPhase.PreREQUEST = append(rulesByPhase.PreREQUEST, rule)
			}
		}

		createInput := networkfirewall.CreateProxyRulesInput{
			ProxyRuleGroupArn: plan.ProxyRuleGroupArn.ValueStringPointer(),
			Rules:             &rulesByPhase,
		}

		_, err = conn.CreateProxyRules(ctx, &createInput)

		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	// Read back to get full state
	readOut, err := findProxyRulesByGroupARN(ctx, conn, plan.ProxyRuleGroupArn.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ProxyRuleGroupArn.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, setProxyRulesState(ctx, readOut, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

// ruleNeedsRecreate determines if a rule needs to be deleted and recreated

func (r *resourceProxyRulesExclusive) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceProxyRulesExclusiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all rule names for this group
	out, err := findProxyRulesByGroupARN(ctx, conn, state.ID.ValueString())
	if err != nil && !retry.NotFound(err) {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	if out != nil && out.ProxyRuleGroup != nil && out.ProxyRuleGroup.Rules != nil {
		ruleNamesMap := proxyCollectRuleNames(out.ProxyRuleGroup.Rules)
		var ruleNames []string
		for name := range ruleNamesMap {
			ruleNames = append(ruleNames, name)
		}

		if len(ruleNames) > 0 {
			input := networkfirewall.DeleteProxyRulesInput{
				ProxyRuleGroupArn: state.ProxyRuleGroupArn.ValueStringPointer(),
				Rules:             ruleNames,
			}

			_, err = conn.DeleteProxyRules(ctx, &input)
			if err != nil {
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					return
				}

				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
				return
			}
		}
	}
}

func findProxyRulesByGroupARN(ctx context.Context, conn *networkfirewall.Client, groupARN string) (*networkfirewall.DescribeProxyRuleGroupOutput, error) {
	input := networkfirewall.DescribeProxyRuleGroupInput{
		ProxyRuleGroupArn: aws.String(groupARN),
	}

	out, err := conn.DescribeProxyRuleGroup(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil || out.ProxyRuleGroup == nil {
		return nil, &retry.NotFoundError{
			Message: "proxy rule group not found",
		}
	}

	return out, nil
}

func setProxyRulesState(ctx context.Context, out *networkfirewall.DescribeProxyRuleGroupOutput, model *resourceProxyRulesExclusiveModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if out.ProxyRuleGroup == nil || out.ProxyRuleGroup.Rules == nil {
		return diags
	}

	rules := out.ProxyRuleGroup.Rules

	model.PostRESPONSE = proxyFlattenRulesForPhase(ctx, rules.PostRESPONSE, &diags)
	model.PreDNS = proxyFlattenRulesForPhase(ctx, rules.PreDNS, &diags)
	model.PreREQUEST = proxyFlattenRulesForPhase(ctx, rules.PreREQUEST, &diags)

	if out.ProxyRuleGroup.ProxyRuleGroupArn != nil {
		model.ProxyRuleGroupArn = fwtypes.ARNValue(aws.ToString(out.ProxyRuleGroup.ProxyRuleGroupArn))
	}

	return diags
}

// conditionsEqual compares only the conditions of two rules
func proxyRuleConditionsEqual(ctx context.Context, a, b proxyRuleModel) bool {
	// Compare conditions count
	if a.Conditions.IsNull() != b.Conditions.IsNull() {
		return false
	}

	if a.Conditions.IsNull() {
		return true
	}

	var aConditions, bConditions []proxyRuleConditionModel
	a.Conditions.ElementsAs(ctx, &aConditions, false)
	b.Conditions.ElementsAs(ctx, &bConditions, false)

	if len(aConditions) != len(bConditions) {
		return false
	}

	// Compare each condition
	for i := range aConditions {
		if aConditions[i].ConditionKey.ValueString() != bConditions[i].ConditionKey.ValueString() {
			return false
		}
		if aConditions[i].ConditionOperator.ValueString() != bConditions[i].ConditionOperator.ValueString() {
			return false
		}

		var aValues, bValues []types.String
		aConditions[i].ConditionValues.ElementsAs(ctx, &aValues, false)
		bConditions[i].ConditionValues.ElementsAs(ctx, &bValues, false)

		if len(aValues) != len(bValues) {
			return false
		}

		for j := range aValues {
			if aValues[j].ValueString() != bValues[j].ValueString() {
				return false
			}
		}
	}

	return true
}

// processProxyRulesPhaseChanges compares state and plan rules for a given phase and populates
// the maps tracking which rules need to be deleted, updated, or recreated
func processProxyRulesPhaseChanges(
	ctx context.Context,
	stateRules, planRules []proxyRuleModel,
	phaseName string,
	rulesToDelete map[string]bool,
	rulesToUpdate map[string]proxyRuleModel,
	rulesToRecreate map[string]proxyRuleRecreateInfo,
) {
	// Compare position by position
	for i, planRule := range planRules {
		planName := planRule.ProxyRuleName.ValueString()

		// Check if a rule exists at this position in state
		if i < len(stateRules) {
			stateRule := stateRules[i]
			stateName := stateRule.ProxyRuleName.ValueString()

			// Check what kind of change this is
			if planName == stateName {
				// Same name at same position - check if attributes changed
				if !proxyRuleConditionsEqual(ctx, stateRule, planRule) ||
					stateRule.Action.ValueEnum() != planRule.Action.ValueEnum() ||
					stateRule.Description.ValueString() != planRule.Description.ValueString() {
					// Action, description, or conditions changed - can use UpdateProxyRule
					rulesToUpdate[planName] = planRule
				}
				// If no changes, do nothing
			} else {
				// Different names at same position
				// The old rule at this position needs to be deleted
				rulesToDelete[stateName] = true
				// The new rule needs to be created at this position
				rulesToRecreate[planName] = proxyRuleRecreateInfo{
					rule:     planRule,
					position: int32(i),
					phase:    phaseName,
				}
			}
		} else {
			// Plan has more rules than state - this is a new rule
			rulesToRecreate[planName] = proxyRuleRecreateInfo{
				rule:     planRule,
				position: int32(i),
				phase:    phaseName,
			}
		}
	}

	// If state has more rules than plan, mark extras for deletion
	for i := len(planRules); i < len(stateRules); i++ {
		rulesToDelete[stateRules[i].ProxyRuleName.ValueString()] = true
	}
}

// proxyExpandRulesForPhase converts plan rules to AWS CreateProxyRule format with positions
func proxyExpandRulesForPhase(ctx context.Context, rulesList fwtypes.ListNestedObjectValueOf[proxyRuleModel], diags *diag.Diagnostics) []awstypes.CreateProxyRule {
	if rulesList.IsNull() || rulesList.IsUnknown() {
		return nil
	}

	var ruleModels []proxyRuleModel
	smerr.AddEnrich(ctx, diags, rulesList.ElementsAs(ctx, &ruleModels, false))
	if diags.HasError() {
		return nil
	}

	var rules []awstypes.CreateProxyRule
	for i, ruleModel := range ruleModels {
		var rule awstypes.CreateProxyRule
		smerr.AddEnrich(ctx, diags, flex.Expand(ctx, ruleModel, &rule))
		if diags.HasError() {
			return nil
		}
		insertPos := int32(i)
		rule.InsertPosition = &insertPos
		rules = append(rules, rule)
	}

	return rules
}

// proxyExtractRulesFromPhase converts AWS ProxyRule format to proxyRuleModel
func proxyExtractRulesFromPhase(ctx context.Context, awsRules []awstypes.ProxyRule, diags *diag.Diagnostics) []proxyRuleModel {
	var ruleModels []proxyRuleModel
	for _, rule := range awsRules {
		var ruleModel proxyRuleModel
		smerr.AddEnrich(ctx, diags, flex.Flatten(ctx, &rule, &ruleModel))
		if diags.HasError() {
			return nil
		}
		ruleModels = append(ruleModels, ruleModel)
	}
	return ruleModels
}

// proxyExtractPlanRules extracts rules from plan's ListNestedObjectValue
func proxyExtractPlanRules(ctx context.Context, rulesList fwtypes.ListNestedObjectValueOf[proxyRuleModel], diags *diag.Diagnostics) []proxyRuleModel {
	if rulesList.IsNull() || rulesList.IsUnknown() {
		return nil
	}

	var ruleModels []proxyRuleModel
	smerr.AddEnrich(ctx, diags, rulesList.ElementsAs(ctx, &ruleModels, false))
	return ruleModels
}

// proxyFlattenRulesForPhase converts AWS rules to framework list format
func proxyFlattenRulesForPhase(ctx context.Context, awsRules []awstypes.ProxyRule, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[proxyRuleModel] {
	if len(awsRules) == 0 {
		return fwtypes.NewListNestedObjectValueOfNull[proxyRuleModel](ctx)
	}

	var ruleModels []proxyRuleModel
	for _, rule := range awsRules {
		var ruleModel proxyRuleModel
		diags.Append(flex.Flatten(ctx, &rule, &ruleModel)...)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfNull[proxyRuleModel](ctx)
		}
		ruleModels = append(ruleModels, ruleModel)
	}

	list, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, ruleModels)
	diags.Append(d...)
	return list
}

// proxyCollectRuleNames collects all rule names from all phases into a map
func proxyCollectRuleNames(rules *awstypes.ProxyRulesByRequestPhase) map[string]bool {
	ruleNames := make(map[string]bool)

	for _, rule := range rules.PostRESPONSE {
		if rule.ProxyRuleName != nil {
			ruleNames[*rule.ProxyRuleName] = true
		}
	}
	for _, rule := range rules.PreDNS {
		if rule.ProxyRuleName != nil {
			ruleNames[*rule.ProxyRuleName] = true
		}
	}
	for _, rule := range rules.PreREQUEST {
		if rule.ProxyRuleName != nil {
			ruleNames[*rule.ProxyRuleName] = true
		}
	}

	return ruleNames
}

// proxyBuildRuleMapByName builds a map of rules by name from current AWS state
func proxyBuildRuleMapByName(ctx context.Context, proxyRuleGroup *awstypes.ProxyRuleGroup, diags *diag.Diagnostics) map[string]proxyRuleModel {
	rulesByName := make(map[string]proxyRuleModel)

	if proxyRuleGroup == nil || proxyRuleGroup.Rules == nil {
		return rulesByName
	}

	rules := proxyRuleGroup.Rules

	for _, rule := range rules.PostRESPONSE {
		var rm proxyRuleModel
		smerr.AddEnrich(ctx, diags, flex.Flatten(ctx, &rule, &rm))
		if diags.HasError() {
			return nil
		}
		rulesByName[rm.ProxyRuleName.ValueString()] = rm
	}
	for _, rule := range rules.PreDNS {
		var rm proxyRuleModel
		smerr.AddEnrich(ctx, diags, flex.Flatten(ctx, &rule, &rm))
		if diags.HasError() {
			return nil
		}
		rulesByName[rm.ProxyRuleName.ValueString()] = rm
	}
	for _, rule := range rules.PreREQUEST {
		var rm proxyRuleModel
		smerr.AddEnrich(ctx, diags, flex.Flatten(ctx, &rule, &rm))
		if diags.HasError() {
			return nil
		}
		rulesByName[rm.ProxyRuleName.ValueString()] = rm
	}

	return rulesByName
}

type resourceProxyRulesExclusiveModel struct {
	framework.WithRegionModel
	ID                types.String                                    `tfsdk:"id"`
	PostRESPONSE      fwtypes.ListNestedObjectValueOf[proxyRuleModel] `tfsdk:"post_response"`
	PreDNS            fwtypes.ListNestedObjectValueOf[proxyRuleModel] `tfsdk:"pre_dns"`
	PreREQUEST        fwtypes.ListNestedObjectValueOf[proxyRuleModel] `tfsdk:"pre_request"`
	ProxyRuleGroupArn fwtypes.ARN                                     `tfsdk:"proxy_rule_group_arn"`
}

type proxyRuleModel struct {
	Action        fwtypes.StringEnum[awstypes.ProxyRulePhaseAction]        `tfsdk:"action"`
	Conditions    fwtypes.ListNestedObjectValueOf[proxyRuleConditionModel] `tfsdk:"conditions"`
	Description   types.String                                             `tfsdk:"description"`
	ProxyRuleName types.String                                             `tfsdk:"proxy_rule_name"`
}

type proxyRuleConditionModel struct {
	ConditionKey      types.String                      `tfsdk:"condition_key"`
	ConditionOperator types.String                      `tfsdk:"condition_operator"`
	ConditionValues   fwtypes.ListValueOf[types.String] `tfsdk:"condition_values"`
}

// proxyRuleRecreateInfo holds information about a rule that needs to be recreated
type proxyRuleRecreateInfo struct {
	rule     proxyRuleModel
	position int32
	phase    string
}

func (data *resourceProxyRulesExclusiveModel) setID() {
	data.ID = data.ProxyRuleGroupArn.StringValue
}
