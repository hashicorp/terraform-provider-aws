// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_wafv2_web_acl_rule", name="Web ACL Rule")
// @IdentityAttribute("web_acl_arn")
// @IdentityAttribute("name")
// @ImportIDHandler("webACLRuleImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="web_acl_arn;name", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(importStateIdAttribute="web_acl_arn")
func newResourceWebACLRule(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceWebACLRule{}, nil
}

const (
	ResNameWebACLRule = "Web ACL Rule"
)

type resourceWebACLRule struct {
	framework.ResourceWithModel[webACLRuleModel]
	framework.WithImportByIdentity
}

func (r *resourceWebACLRule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Rule name, unique within the Web ACL.",
			},
			names.AttrPriority: schema.Int32Attribute{
				Required:    true,
				Description: "Rule priority. Rules with lower priority are evaluated first.",
			},
			"web_acl_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "ARN of the Web ACL to add the rule to.",
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrAction: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleActionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"allow": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleEmptyModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"custom_request_handling": customRequestHandlingBlock(ctx),
								},
							},
						},
						"block": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleBlockActionModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"custom_response": customResponseBlock(ctx),
								},
							},
						},
						"count": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleEmptyModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"custom_request_handling": customRequestHandlingBlock(ctx),
								},
							},
						},
						"captcha": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleEmptyModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"custom_request_handling": customRequestHandlingBlock(ctx),
								},
							},
						},
						"challenge": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleEmptyModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"custom_request_handling": customRequestHandlingBlock(ctx),
								},
							},
						},
					},
				},
				Description: "Action to take when the rule matches. Specify exactly one of action or override_action.",
			},
			"captcha_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleImmunityTimeModel](ctx),
				Validators: []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"immunity_time_property": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleImmunityTimePropertyModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"immunity_time": schema.Int64Attribute{Optional: true},
								},
							},
						},
					},
				},
				Description: "Specifies how WAF should handle CAPTCHA evaluations. Overrides the web ACL level setting.",
			},
			"challenge_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleImmunityTimeModel](ctx),
				Validators: []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"immunity_time_property": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleImmunityTimePropertyModel](ctx),
							Validators: []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"immunity_time": schema.Int64Attribute{Optional: true},
								},
							},
						},
					},
				},
				Description: "Specifies how WAF should handle Challenge evaluations. Overrides the web ACL level setting.",
			},
			"override_action": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleOverrideActionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(names.AttrAction)),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"count": schema.ListNestedBlock{
							CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleOverrideActionEmptyModel](ctx),
							Validators:   []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{},
						},
						"none": schema.ListNestedBlock{
							CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleOverrideActionEmptyModel](ctx),
							Validators:   []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{},
						},
					},
				},
				Description: "Override action for rule group and managed rule group statements. Use instead of action.",
			},
			"rule_label": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleLabelModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required:    true,
							Description: "Label string.",
						},
					},
				},
				Description: "Labels to apply to matching web requests.",
			},
			"statement": statementBlock(ctx),
			"visibility_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleVisibilityConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cloudwatch_metrics_enabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
						names.AttrMetricName: schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"sampled_requests_enabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
				},
				Description: "CloudWatch metrics configuration.",
			},
		},
		Description: "Manages an individual rule within a WAFv2 Web ACL. Creates proper Terraform dependencies for safe deletion of referenced resources like IP sets.",
	}
}

func (r *resourceWebACLRule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var plan webACLRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webACLID, webACLName, webACLScope, err := parseWebACLARN(plan.WebACLARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRule, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRule, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Check for duplicate priority or name
	for _, rule := range webACL.WebACL.Rules {
		if rule.Priority == plan.Priority.ValueInt32() {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRule, plan.Name.String(), nil),
				fmt.Sprintf("Rule with priority %d already exists in Web ACL", plan.Priority.ValueInt32()),
			)
			return
		}
		if aws.ToString(rule.Name) == plan.Name.ValueString() {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRule, plan.Name.String(), nil),
				fmt.Sprintf("Rule with name %s already exists in Web ACL", plan.Name.ValueString()),
			)
			return
		}
	}

	// Build the rule
	var newRule awstypes.Rule
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, &newRule)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webACL.WebACL.Rules = append(webACL.WebACL.Rules, newRule)

	updateInput := &wafv2.UpdateWebACLInput{
		Id:                   aws.String(webACLID),
		Name:                 aws.String(webACLName),
		Scope:                awstypes.Scope(webACLScope),
		DefaultAction:        webACL.WebACL.DefaultAction,
		Rules:                webACL.WebACL.Rules,
		VisibilityConfig:     webACL.WebACL.VisibilityConfig,
		LockToken:            webACL.LockToken,
		AssociationConfig:    webACL.WebACL.AssociationConfig,
		CaptchaConfig:        webACL.WebACL.CaptchaConfig,
		ChallengeConfig:      webACL.WebACL.ChallengeConfig,
		CustomResponseBodies: webACL.WebACL.CustomResponseBodies,
		TokenDomains:         webACL.WebACL.TokenDomains,
	}

	if webACL.WebACL.Description != nil && aws.ToString(webACL.WebACL.Description) != "" {
		updateInput.Description = webACL.WebACL.Description
	}

	const timeout = 5 * time.Minute
	_, err = tfresource.RetryWhenIsA[any, *awstypes.WAFUnavailableEntityException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRule, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Read back the WebACL to get computed values
	webACL, err = findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRule, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Find the created rule
	var createdRule *awstypes.Rule
	for i := range webACL.WebACL.Rules {
		if aws.ToString(webACL.WebACL.Rules[i].Name) == plan.Name.ValueString() {
			createdRule = &webACL.WebACL.Rules[i]
			break
		}
	}

	if createdRule == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRule, plan.Name.String(), nil),
			"Created rule not found in Web ACL after update",
		)
		return
	}

	// Flatten the created rule to get computed values
	var state webACLRuleModel
	resp.Diagnostics.Append(flex.Flatten(ctx, createdRule, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve fields from plan that aren't in the Rule
	state.WebACLARN = plan.WebACLARN
	state.WithRegionModel = plan.WithRegionModel

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceWebACLRule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var state webACLRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webACLID, webACLName, webACLScope, err := parseWebACLARN(state.WebACLARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionReading, ResNameWebACLRule, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionReading, ResNameWebACLRule, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Find the rule
	var foundRule *awstypes.Rule
	for i := range webACL.WebACL.Rules {
		if aws.ToString(webACL.WebACL.Rules[i].Name) == state.Name.ValueString() {
			foundRule = &webACL.WebACL.Rules[i]
			break
		}
	}

	if foundRule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Flatten the rule back to state
	resp.Diagnostics.Append(flex.Flatten(ctx, foundRule, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceWebACLRule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var plan, state webACLRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webACLID, webACLName, webACLScope, err := parseWebACLARN(plan.WebACLARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRule, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRule, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Build updated rule
	var updatedRule awstypes.Rule
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, &updatedRule)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Replace the rule in the list
	var updatedRules []awstypes.Rule
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) == state.Name.ValueString() {
			updatedRules = append(updatedRules, updatedRule)
		} else {
			updatedRules = append(updatedRules, rule)
		}
	}

	updateInput := &wafv2.UpdateWebACLInput{
		Id:                   aws.String(webACLID),
		Name:                 aws.String(webACLName),
		Scope:                awstypes.Scope(webACLScope),
		DefaultAction:        webACL.WebACL.DefaultAction,
		Rules:                updatedRules,
		VisibilityConfig:     webACL.WebACL.VisibilityConfig,
		LockToken:            webACL.LockToken,
		AssociationConfig:    webACL.WebACL.AssociationConfig,
		CaptchaConfig:        webACL.WebACL.CaptchaConfig,
		ChallengeConfig:      webACL.WebACL.ChallengeConfig,
		CustomResponseBodies: webACL.WebACL.CustomResponseBodies,
		TokenDomains:         webACL.WebACL.TokenDomains,
	}

	if webACL.WebACL.Description != nil && aws.ToString(webACL.WebACL.Description) != "" {
		updateInput.Description = webACL.WebACL.Description
	}

	const timeout = 5 * time.Minute
	_, err = tfresource.RetryWhenIsA[any, *awstypes.WAFUnavailableEntityException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionUpdating, ResNameWebACLRule, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceWebACLRule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var state webACLRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webACLID, webACLName, webACLScope, err := parseWebACLARN(state.WebACLARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRule, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRule, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Filter out the rule
	var updatedRules []awstypes.Rule
	ruleFound := false
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) != state.Name.ValueString() {
			updatedRules = append(updatedRules, rule)
		} else {
			ruleFound = true
		}
	}

	if !ruleFound {
		return
	}

	updateInput := &wafv2.UpdateWebACLInput{
		Id:                   aws.String(webACLID),
		Name:                 aws.String(webACLName),
		Scope:                awstypes.Scope(webACLScope),
		DefaultAction:        webACL.WebACL.DefaultAction,
		Rules:                updatedRules,
		VisibilityConfig:     webACL.WebACL.VisibilityConfig,
		LockToken:            webACL.LockToken,
		AssociationConfig:    webACL.WebACL.AssociationConfig,
		CaptchaConfig:        webACL.WebACL.CaptchaConfig,
		ChallengeConfig:      webACL.WebACL.ChallengeConfig,
		CustomResponseBodies: webACL.WebACL.CustomResponseBodies,
		TokenDomains:         webACL.WebACL.TokenDomains,
	}

	if webACL.WebACL.Description != nil && aws.ToString(webACL.WebACL.Description) != "" {
		updateInput.Description = webACL.WebACL.Description
	}

	const timeout = 5 * time.Minute
	_, err = tfresource.RetryWhenIsA[any, *awstypes.WAFUnavailableEntityException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRule, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

type webACLRuleImportID struct{}

func (webACLRuleImportID) Parse(id string) (string, map[string]any, error) {
	webACLARN, ruleName, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <web-acl-arn>%s<rule-name>", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		"web_acl_arn":  webACLARN,
		names.AttrName: ruleName,
	}

	return id, result, nil
}

// Models

type webACLRuleModel struct {
	framework.WithRegionModel
	Name             types.String                                                     `tfsdk:"name"`
	Priority         types.Int32                                                      `tfsdk:"priority"`
	WebACLARN        types.String                                                     `tfsdk:"web_acl_arn"`
	Action           fwtypes.ListNestedObjectValueOf[webACLRuleActionModel]           `tfsdk:"action"`
	CaptchaConfig    fwtypes.ListNestedObjectValueOf[webACLRuleImmunityTimeModel]     `tfsdk:"captcha_config"`
	ChallengeConfig  fwtypes.ListNestedObjectValueOf[webACLRuleImmunityTimeModel]     `tfsdk:"challenge_config"`
	OverrideAction   fwtypes.ListNestedObjectValueOf[webACLRuleOverrideActionModel]   `tfsdk:"override_action"`
	RuleLabel        fwtypes.ListNestedObjectValueOf[webACLRuleLabelModel]            `tfsdk:"rule_label"`
	Statement        fwtypes.ListNestedObjectValueOf[webACLRuleStatementModel]        `tfsdk:"statement"`
	VisibilityConfig fwtypes.ListNestedObjectValueOf[webACLRuleVisibilityConfigModel] `tfsdk:"visibility_config"`
}

func (m webACLRuleModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	rule := &awstypes.Rule{
		Name:     m.Name.ValueStringPointer(),
		Priority: m.Priority.ValueInt32(),
	}

	// Expand action
	if !m.Action.IsNull() {
		actionData, d := m.Action.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		action, d := actionData.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		rule.Action = action.(*awstypes.RuleAction)
	}

	// Expand override_action
	if !m.OverrideAction.IsNull() {
		oaData, d := m.OverrideAction.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		oa, d := oaData.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		rule.OverrideAction = oa.(*awstypes.OverrideAction)
	}

	// Expand captcha_config
	if !m.CaptchaConfig.IsNull() {
		ccData, d := m.CaptchaConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		cc, d := ccData.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		rule.CaptchaConfig = cc.(*awstypes.CaptchaConfig)
	}

	// Expand challenge_config
	if !m.ChallengeConfig.IsNull() {
		ccData, d := m.ChallengeConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		cc, d := ccData.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		rule.ChallengeConfig = cc.(*awstypes.ChallengeConfig)
	}

	// Expand rule_label
	if !m.RuleLabel.IsNull() {
		labels, d := m.RuleLabel.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		for _, label := range labels {
			rule.RuleLabels = append(rule.RuleLabels, awstypes.Label{Name: label.Name.ValueStringPointer()})
		}
	}

	// Expand statement
	if !m.Statement.IsNull() {
		stmtData, d := m.Statement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		stmt, d := stmtData.Expand(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		rule.Statement = stmt.(*awstypes.Statement)
	}

	// Expand visibility config
	if !m.VisibilityConfig.IsNull() {
		vcData, d := m.VisibilityConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var vc awstypes.VisibilityConfig
		diags.Append(flex.Expand(ctx, vcData, &vc)...)
		if diags.HasError() {
			return nil, diags
		}
		rule.VisibilityConfig = &vc
	}

	return rule, diags
}

type webACLRuleVisibilityConfigModel struct {
	CloudWatchMetricsEnabled types.Bool   `tfsdk:"cloudwatch_metrics_enabled"`
	MetricName               types.String `tfsdk:"metric_name"`
	SampledRequestsEnabled   types.Bool   `tfsdk:"sampled_requests_enabled"`
}

type webACLRuleActionModel struct {
	Allow     fwtypes.ListNestedObjectValueOf[webACLRuleEmptyModel]       `tfsdk:"allow"`
	Block     fwtypes.ListNestedObjectValueOf[webACLRuleBlockActionModel] `tfsdk:"block"`
	Count     fwtypes.ListNestedObjectValueOf[webACLRuleEmptyModel]       `tfsdk:"count"`
	Captcha   fwtypes.ListNestedObjectValueOf[webACLRuleEmptyModel]       `tfsdk:"captcha"`
	Challenge fwtypes.ListNestedObjectValueOf[webACLRuleEmptyModel]       `tfsdk:"challenge"`
}

func (m webACLRuleActionModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.Allow.IsNull():
		allowData, d := m.Allow.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		allowAction := &awstypes.AllowAction{}
		if !allowData.CustomRequestHandling.IsNull() {
			crhData, d := allowData.CustomRequestHandling.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}

			var crh awstypes.CustomRequestHandling
			diags.Append(flex.Expand(ctx, crhData, &crh)...)
			if diags.HasError() {
				return nil, diags
			}
			allowAction.CustomRequestHandling = &crh
		}
		return &awstypes.RuleAction{Allow: allowAction}, diags

	case !m.Block.IsNull():
		blockData, d := m.Block.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		blockAction := &awstypes.BlockAction{}
		if !blockData.CustomResponse.IsNull() {
			crData, d := blockData.CustomResponse.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}

			var cr awstypes.CustomResponse
			diags.Append(flex.Expand(ctx, crData, &cr)...)
			if diags.HasError() {
				return nil, diags
			}
			blockAction.CustomResponse = &cr
		}
		return &awstypes.RuleAction{Block: blockAction}, diags

	case !m.Count.IsNull():
		countData, d := m.Count.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		countAction := &awstypes.CountAction{}
		if !countData.CustomRequestHandling.IsNull() {
			crhData, d := countData.CustomRequestHandling.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}

			var crh awstypes.CustomRequestHandling
			diags.Append(flex.Expand(ctx, crhData, &crh)...)
			if diags.HasError() {
				return nil, diags
			}
			countAction.CustomRequestHandling = &crh
		}
		return &awstypes.RuleAction{Count: countAction}, diags

	case !m.Captcha.IsNull():
		captchaData, d := m.Captcha.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		captchaAction := &awstypes.CaptchaAction{}
		if !captchaData.CustomRequestHandling.IsNull() {
			crhData, d := captchaData.CustomRequestHandling.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}

			var crh awstypes.CustomRequestHandling
			diags.Append(flex.Expand(ctx, crhData, &crh)...)
			if diags.HasError() {
				return nil, diags
			}
			captchaAction.CustomRequestHandling = &crh
		}
		return &awstypes.RuleAction{Captcha: captchaAction}, diags

	case !m.Challenge.IsNull():
		challengeData, d := m.Challenge.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		challengeAction := &awstypes.ChallengeAction{}
		if !challengeData.CustomRequestHandling.IsNull() {
			crhData, d := challengeData.CustomRequestHandling.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}

			var crh awstypes.CustomRequestHandling
			diags.Append(flex.Expand(ctx, crhData, &crh)...)
			if diags.HasError() {
				return nil, diags
			}
			challengeAction.CustomRequestHandling = &crh
		}
		return &awstypes.RuleAction{Challenge: challengeAction}, diags
	}
	return nil, diags
}

func (m *webACLRuleActionModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch action := v.(type) {
	case *awstypes.RuleAction:
		return m.flattenAction(ctx, action)
	case awstypes.RuleAction:
		return m.flattenAction(ctx, &action)
	}
	return diags
}

func (m *webACLRuleActionModel) flattenAction(ctx context.Context, action *awstypes.RuleAction) diag.Diagnostics {
	var diags diag.Diagnostics

	// Initialize all to null
	m.Allow = fwtypes.NewListNestedObjectValueOfNull[webACLRuleEmptyModel](ctx)
	m.Block = fwtypes.NewListNestedObjectValueOfNull[webACLRuleBlockActionModel](ctx)
	m.Count = fwtypes.NewListNestedObjectValueOfNull[webACLRuleEmptyModel](ctx)
	m.Captcha = fwtypes.NewListNestedObjectValueOfNull[webACLRuleEmptyModel](ctx)
	m.Challenge = fwtypes.NewListNestedObjectValueOfNull[webACLRuleEmptyModel](ctx)

	switch {
	case action.Allow != nil:
		emptyVal := &webACLRuleEmptyModel{
			CustomRequestHandling: fwtypes.NewListNestedObjectValueOfNull[webACLRuleCustomRequestHandlingModel](ctx),
		}
		if action.Allow.CustomRequestHandling != nil {
			var crhModel webACLRuleCustomRequestHandlingModel
			diags.Append(flex.Flatten(ctx, action.Allow.CustomRequestHandling, &crhModel)...)
			if !diags.HasError() {
				emptyVal.CustomRequestHandling, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleCustomRequestHandlingModel{&crhModel}, nil)
			}
		}
		m.Allow, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleEmptyModel{emptyVal}, nil)

	case action.Block != nil:
		blockVal := &webACLRuleBlockActionModel{
			CustomResponse: fwtypes.NewListNestedObjectValueOfNull[webACLRuleCustomResponseModel](ctx),
		}
		if action.Block.CustomResponse != nil {
			var crModel webACLRuleCustomResponseModel
			diags.Append(flex.Flatten(ctx, action.Block.CustomResponse, &crModel)...)
			if !diags.HasError() {
				blockVal.CustomResponse, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleCustomResponseModel{&crModel}, nil)
			}
		}
		m.Block, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleBlockActionModel{blockVal}, nil)

	case action.Count != nil:
		emptyVal := &webACLRuleEmptyModel{
			CustomRequestHandling: fwtypes.NewListNestedObjectValueOfNull[webACLRuleCustomRequestHandlingModel](ctx),
		}
		if action.Count.CustomRequestHandling != nil {
			var crhModel webACLRuleCustomRequestHandlingModel
			diags.Append(flex.Flatten(ctx, action.Count.CustomRequestHandling, &crhModel)...)
			if !diags.HasError() {
				emptyVal.CustomRequestHandling, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleCustomRequestHandlingModel{&crhModel}, nil)
			}
		}
		m.Count, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleEmptyModel{emptyVal}, nil)

	case action.Captcha != nil:
		emptyVal := &webACLRuleEmptyModel{
			CustomRequestHandling: fwtypes.NewListNestedObjectValueOfNull[webACLRuleCustomRequestHandlingModel](ctx),
		}
		if action.Captcha.CustomRequestHandling != nil {
			var crhModel webACLRuleCustomRequestHandlingModel
			diags.Append(flex.Flatten(ctx, action.Captcha.CustomRequestHandling, &crhModel)...)
			if !diags.HasError() {
				emptyVal.CustomRequestHandling, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleCustomRequestHandlingModel{&crhModel}, nil)
			}
		}
		m.Captcha, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleEmptyModel{emptyVal}, nil)

	case action.Challenge != nil:
		emptyVal := &webACLRuleEmptyModel{
			CustomRequestHandling: fwtypes.NewListNestedObjectValueOfNull[webACLRuleCustomRequestHandlingModel](ctx),
		}
		if action.Challenge.CustomRequestHandling != nil {
			var crhModel webACLRuleCustomRequestHandlingModel
			diags.Append(flex.Flatten(ctx, action.Challenge.CustomRequestHandling, &crhModel)...)
			if !diags.HasError() {
				emptyVal.CustomRequestHandling, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleCustomRequestHandlingModel{&crhModel}, nil)
			}
		}
		m.Challenge, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleEmptyModel{emptyVal}, nil)
	}
	return diags
}

type webACLRuleEmptyModel struct {
	CustomRequestHandling fwtypes.ListNestedObjectValueOf[webACLRuleCustomRequestHandlingModel] `tfsdk:"custom_request_handling"`
}

type webACLRuleBlockActionModel struct {
	CustomResponse fwtypes.ListNestedObjectValueOf[webACLRuleCustomResponseModel] `tfsdk:"custom_response"`
}

type webACLRuleCustomRequestHandlingModel struct {
	InsertHeader fwtypes.ListNestedObjectValueOf[webACLRuleCustomHeaderModel] `tfsdk:"insert_header"`
}

type webACLRuleCustomResponseModel struct {
	ResponseCode          types.Int32                                                  `tfsdk:"response_code"`
	CustomResponseBodyKey types.String                                                 `tfsdk:"custom_response_body_key"`
	ResponseHeader        fwtypes.ListNestedObjectValueOf[webACLRuleCustomHeaderModel] `tfsdk:"response_header"`
}

type webACLRuleCustomHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type webACLRuleStatementModel struct {
	IPSetReferenceStatement           fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel]           `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]                 `tfsdk:"geo_match_statement"`
	RuleGroupReferenceStatement       fwtypes.ListNestedObjectValueOf[webACLRuleRuleGroupReferenceStatementModel]       `tfsdk:"rule_group_reference_statement"`
	ManagedRuleGroupStatement         fwtypes.ListNestedObjectValueOf[webACLRuleManagedRuleGroupStatementModel]         `tfsdk:"managed_rule_group_statement"`
	RegexPatternSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleRegexPatternSetReferenceStatementModel] `tfsdk:"regex_pattern_set_reference_statement"`
	RateBasedStatement                fwtypes.ListNestedObjectValueOf[webACLRuleRateBasedStatementModel]                `tfsdk:"rate_based_statement"`
	ByteMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleByteMatchStatementModel]                `tfsdk:"byte_match_statement"`
	SqliMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleSqliMatchStatementModel]                `tfsdk:"sqli_match_statement"`
	XssMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleXssMatchStatementModel]                 `tfsdk:"xss_match_statement"`
	SizeConstraintStatement           fwtypes.ListNestedObjectValueOf[webACLRuleSizeConstraintStatementModel]           `tfsdk:"size_constraint_statement"`
	RegexMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleRegexMatchStatementModel]               `tfsdk:"regex_match_statement"`
	LabelMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleLabelMatchStatementModel]               `tfsdk:"label_match_statement"`
	AsnMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleAsnMatchStatementModel]                 `tfsdk:"asn_match_statement"`
}

func (m webACLRuleStatementModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.IPSetReferenceStatement.IsNull():
		ipSetData, d := m.IPSetReferenceStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var ipSetStmt awstypes.IPSetReferenceStatement
		diags.Append(flex.Expand(ctx, ipSetData, &ipSetStmt)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.Statement{IPSetReferenceStatement: &ipSetStmt}, diags

	case !m.GeoMatchStatement.IsNull():
		geoData, d := m.GeoMatchStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var geoStmt awstypes.GeoMatchStatement
		diags.Append(flex.Expand(ctx, geoData, &geoStmt)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.Statement{GeoMatchStatement: &geoStmt}, diags

	case !m.RuleGroupReferenceStatement.IsNull():
		rgData, d := m.RuleGroupReferenceStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var rg awstypes.RuleGroupReferenceStatement
		diags.Append(flex.Expand(ctx, rgData, &rg)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.Statement{RuleGroupReferenceStatement: &rg}, diags

	case !m.ManagedRuleGroupStatement.IsNull():
		mrgData, d := m.ManagedRuleGroupStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		mrg := &awstypes.ManagedRuleGroupStatement{
			Name:       mrgData.Name.ValueStringPointer(),
			VendorName: mrgData.VendorName.ValueStringPointer(),
		}

		if !mrgData.Version.IsNull() {
			mrg.Version = mrgData.Version.ValueStringPointer()
		}

		return &awstypes.Statement{ManagedRuleGroupStatement: mrg}, diags

	case !m.RegexPatternSetReferenceStatement.IsNull():
		rpsData, d := m.RegexPatternSetReferenceStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var rps awstypes.RegexPatternSetReferenceStatement
		diags.Append(flex.Expand(ctx, rpsData, &rps)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.Statement{RegexPatternSetReferenceStatement: &rps}, diags

	case !m.RateBasedStatement.IsNull():
		rbsData, d := m.RateBasedStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var rbs awstypes.RateBasedStatement
		diags.Append(flex.Expand(ctx, rbsData, &rbs)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.Statement{RateBasedStatement: &rbs}, diags

	case !m.ByteMatchStatement.IsNull():
		bmsData, d := m.ByteMatchStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var bms awstypes.ByteMatchStatement
		diags.Append(flex.Expand(ctx, bmsData, &bms)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.Statement{ByteMatchStatement: &bms}, diags

	case !m.SqliMatchStatement.IsNull():
		return &awstypes.Statement{SqliMatchStatement: &awstypes.SqliMatchStatement{}}, diags

	case !m.XssMatchStatement.IsNull():
		return &awstypes.Statement{XssMatchStatement: &awstypes.XssMatchStatement{}}, diags

	case !m.SizeConstraintStatement.IsNull():
		return &awstypes.Statement{SizeConstraintStatement: &awstypes.SizeConstraintStatement{}}, diags

	case !m.RegexMatchStatement.IsNull():
		rmsData, d := m.RegexMatchStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var rms awstypes.RegexMatchStatement
		diags.Append(flex.Expand(ctx, rmsData, &rms)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.Statement{RegexMatchStatement: &rms}, diags

	case !m.LabelMatchStatement.IsNull():
		labelData, d := m.LabelMatchStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var labelStmt awstypes.LabelMatchStatement
		diags.Append(flex.Expand(ctx, labelData, &labelStmt)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.Statement{LabelMatchStatement: &labelStmt}, diags

	case !m.AsnMatchStatement.IsNull():
		asnData, d := m.AsnMatchStatement.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var asnStmt awstypes.AsnMatchStatement
		diags.Append(flex.Expand(ctx, asnData, &asnStmt)...)
		if diags.HasError() {
			return nil, diags
		}

		return &awstypes.Statement{AsnMatchStatement: &asnStmt}, diags
	}
	return nil, diags
}

func (m *webACLRuleStatementModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch stmt := v.(type) {
	case *awstypes.Statement:
		// Handle pointer case
		return m.flattenStatement(ctx, stmt)
	case awstypes.Statement:
		// Handle value case
		return m.flattenStatement(ctx, &stmt)
	}
	return diags
}

func (m *webACLRuleStatementModel) flattenStatement(ctx context.Context, stmt *awstypes.Statement) diag.Diagnostics {
	var diags diag.Diagnostics

	m.IPSetReferenceStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleIPSetReferenceStatementModel](ctx)
	m.GeoMatchStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleGeoMatchStatementModel](ctx)
	m.RuleGroupReferenceStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleRuleGroupReferenceStatementModel](ctx)
	m.ManagedRuleGroupStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleManagedRuleGroupStatementModel](ctx)
	m.RegexPatternSetReferenceStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleRegexPatternSetReferenceStatementModel](ctx)
	m.RateBasedStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleRateBasedStatementModel](ctx)
	m.ByteMatchStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleByteMatchStatementModel](ctx)
	m.SqliMatchStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleSqliMatchStatementModel](ctx)
	m.XssMatchStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleXssMatchStatementModel](ctx)
	m.SizeConstraintStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleSizeConstraintStatementModel](ctx)
	m.RegexMatchStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleRegexMatchStatementModel](ctx)
	m.LabelMatchStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleLabelMatchStatementModel](ctx)
	m.AsnMatchStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleAsnMatchStatementModel](ctx)

	switch {
	case stmt.IPSetReferenceStatement != nil:
		var ipSetModel webACLRuleIPSetReferenceStatementModel
		diags.Append(flex.Flatten(ctx, stmt.IPSetReferenceStatement, &ipSetModel)...)
		if diags.HasError() {
			return diags
		}
		m.IPSetReferenceStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleIPSetReferenceStatementModel{&ipSetModel}, nil)

	case stmt.GeoMatchStatement != nil:
		var geoModel webACLRuleGeoMatchStatementModel
		diags.Append(flex.Flatten(ctx, stmt.GeoMatchStatement, &geoModel)...)
		if diags.HasError() {
			return diags
		}
		m.GeoMatchStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleGeoMatchStatementModel{&geoModel}, nil)

	case stmt.RuleGroupReferenceStatement != nil:
		var rgModel webACLRuleRuleGroupReferenceStatementModel
		diags.Append(flex.Flatten(ctx, stmt.RuleGroupReferenceStatement, &rgModel)...)
		if !diags.HasError() {
			m.RuleGroupReferenceStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleRuleGroupReferenceStatementModel{&rgModel}, nil)
		}

	case stmt.ManagedRuleGroupStatement != nil:
		mrgModel := webACLRuleManagedRuleGroupStatementModel{
			Name:                    types.StringPointerValue(stmt.ManagedRuleGroupStatement.Name),
			VendorName:              types.StringPointerValue(stmt.ManagedRuleGroupStatement.VendorName),
			Version:                 types.StringPointerValue(stmt.ManagedRuleGroupStatement.Version),
			ManagedRuleGroupConfigs: fwtypes.NewListNestedObjectValueOfNull[webACLRuleManagedRuleGroupConfigModel](ctx),
			RuleActionOverrides:     fwtypes.NewListNestedObjectValueOfNull[webACLRuleRuleActionOverrideModel](ctx),
			ScopeDownStatement:      fwtypes.NewListNestedObjectValueOfNull[webACLRuleScopeDownStatementModel](ctx),
		}
		m.ManagedRuleGroupStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleManagedRuleGroupStatementModel{&mrgModel}, nil)

	case stmt.RegexPatternSetReferenceStatement != nil:
		var rpsModel webACLRuleRegexPatternSetReferenceStatementModel
		diags.Append(flex.Flatten(ctx, stmt.RegexPatternSetReferenceStatement, &rpsModel)...)
		if !diags.HasError() {
			m.RegexPatternSetReferenceStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleRegexPatternSetReferenceStatementModel{&rpsModel}, nil)
		}

	case stmt.RateBasedStatement != nil:
		var rbsModel webACLRuleRateBasedStatementModel
		diags.Append(flex.Flatten(ctx, stmt.RateBasedStatement, &rbsModel)...)
		if !diags.HasError() {
			m.RateBasedStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleRateBasedStatementModel{&rbsModel}, nil)
		}

	case stmt.ByteMatchStatement != nil:
		var bmsModel webACLRuleByteMatchStatementModel
		diags.Append(flex.Flatten(ctx, stmt.ByteMatchStatement, &bmsModel)...)
		if !diags.HasError() {
			m.ByteMatchStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleByteMatchStatementModel{&bmsModel}, nil)
		}

	case stmt.SqliMatchStatement != nil:
		sqliModel := webACLRuleSqliMatchStatementModel{}
		m.SqliMatchStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleSqliMatchStatementModel{&sqliModel}, nil)

	case stmt.XssMatchStatement != nil:
		xssModel := webACLRuleXssMatchStatementModel{}
		m.XssMatchStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleXssMatchStatementModel{&xssModel}, nil)

	case stmt.SizeConstraintStatement != nil:
		sizeModel := webACLRuleSizeConstraintStatementModel{}
		m.SizeConstraintStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleSizeConstraintStatementModel{&sizeModel}, nil)

	case stmt.RegexMatchStatement != nil:
		var rmsModel webACLRuleRegexMatchStatementModel
		diags.Append(flex.Flatten(ctx, stmt.RegexMatchStatement, &rmsModel)...)
		if !diags.HasError() {
			m.RegexMatchStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleRegexMatchStatementModel{&rmsModel}, nil)
		}

	case stmt.LabelMatchStatement != nil:
		var labelModel webACLRuleLabelMatchStatementModel
		diags.Append(flex.Flatten(ctx, stmt.LabelMatchStatement, &labelModel)...)
		if !diags.HasError() {
			m.LabelMatchStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleLabelMatchStatementModel{&labelModel}, nil)
		}

	case stmt.AsnMatchStatement != nil:
		var asnModel webACLRuleAsnMatchStatementModel
		diags.Append(flex.Flatten(ctx, stmt.AsnMatchStatement, &asnModel)...)
		if !diags.HasError() {
			m.AsnMatchStatement, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleAsnMatchStatementModel{&asnModel}, nil)
		}
	}
	return diags
}

type webACLRuleIPSetReferenceStatementModel struct {
	ARN                    types.String                                                           `tfsdk:"arn"`
	IPSetForwardedIPConfig fwtypes.ListNestedObjectValueOf[webACLRuleIPSetForwardedIPConfigModel] `tfsdk:"ip_set_forwarded_ip_config"`
}

type webACLRuleIPSetForwardedIPConfigModel struct {
	FallbackBehavior fwtypes.StringEnum[awstypes.FallbackBehavior]    `tfsdk:"fallback_behavior"`
	HeaderName       types.String                                     `tfsdk:"header_name"`
	Position         fwtypes.StringEnum[awstypes.ForwardedIPPosition] `tfsdk:"position"`
}

type webACLRuleGeoMatchStatementModel struct {
	CountryCodes      fwtypes.ListValueOf[fwtypes.StringEnum[awstypes.CountryCode]]     `tfsdk:"country_codes"`
	ForwardedIPConfig fwtypes.ListNestedObjectValueOf[webACLRuleForwardedIPConfigModel] `tfsdk:"forwarded_ip_config"`
}

type webACLRuleForwardedIPConfigModel struct {
	FallbackBehavior fwtypes.StringEnum[awstypes.FallbackBehavior] `tfsdk:"fallback_behavior"`
	HeaderName       types.String                                  `tfsdk:"header_name"`
}

type webACLRuleRuleGroupReferenceStatementModel struct {
	ARN                 types.String                                                       `tfsdk:"arn"`
	ExcludedRules       fwtypes.ListNestedObjectValueOf[webACLRuleExcludedRuleModel]       `tfsdk:"excluded_rule"`
	RuleActionOverrides fwtypes.ListNestedObjectValueOf[webACLRuleRuleActionOverrideModel] `tfsdk:"rule_action_override"`
}

type webACLRuleExcludedRuleModel struct {
	Name types.String `tfsdk:"name"`
}

type webACLRuleManagedRuleGroupStatementModel struct {
	Name                    types.String                                                           `tfsdk:"name"`
	VendorName              types.String                                                           `tfsdk:"vendor_name"`
	Version                 types.String                                                           `tfsdk:"version"`
	ManagedRuleGroupConfigs fwtypes.ListNestedObjectValueOf[webACLRuleManagedRuleGroupConfigModel] `tfsdk:"managed_rule_group_configs"`
	RuleActionOverrides     fwtypes.ListNestedObjectValueOf[webACLRuleRuleActionOverrideModel]     `tfsdk:"rule_action_override"`
	ScopeDownStatement      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownStatementModel]     `tfsdk:"scope_down_statement"`
}

type webACLRuleManagedRuleGroupConfigModel struct {
}

type webACLRuleRuleActionOverrideModel struct {
	Name        types.String                                           `tfsdk:"name"`
	ActionToUse fwtypes.ListNestedObjectValueOf[webACLRuleActionModel] `tfsdk:"action_to_use"`
}

type webACLRuleScopeDownStatementModel struct {
	IPSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel] `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement       fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]       `tfsdk:"geo_match_statement"`
	ByteMatchStatement      fwtypes.ListNestedObjectValueOf[webACLRuleByteMatchStatementModel]      `tfsdk:"byte_match_statement"`
	SqliMatchStatement      fwtypes.ListNestedObjectValueOf[webACLRuleSqliMatchStatementModel]      `tfsdk:"sqli_match_statement"`
	XssMatchStatement       fwtypes.ListNestedObjectValueOf[webACLRuleXssMatchStatementModel]       `tfsdk:"xss_match_statement"`
	SizeConstraintStatement fwtypes.ListNestedObjectValueOf[webACLRuleSizeConstraintStatementModel] `tfsdk:"size_constraint_statement"`
	RegexMatchStatement     fwtypes.ListNestedObjectValueOf[webACLRuleRegexMatchStatementModel]     `tfsdk:"regex_match_statement"`
	LabelMatchStatement     fwtypes.ListNestedObjectValueOf[webACLRuleLabelMatchStatementModel]     `tfsdk:"label_match_statement"`
	AsnMatchStatement       fwtypes.ListNestedObjectValueOf[webACLRuleAsnMatchStatementModel]       `tfsdk:"asn_match_statement"`
}

type webACLRuleRegexPatternSetReferenceStatementModel struct {
	ARN                 types.String                                                  `tfsdk:"arn"`
	FieldToMatch        fwtypes.ListNestedObjectValueOf[webACLRuleFieldToMatchModel]  `tfsdk:"field_to_match"`
	TextTransformations fwtypes.ListNestedObjectValueOf[webACLRuleTextTransformModel] `tfsdk:"text_transformation"`
}

type webACLRuleRateBasedStatementCustomKeyModel struct {
	Cookie         fwtypes.ListNestedObjectValueOf[webACLRuleRateBasedStatementCustomKeyCookieModel]         `tfsdk:"cookie"`
	ForwardedIP    fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]                                `tfsdk:"forwarded_ip"`
	Header         fwtypes.ListNestedObjectValueOf[webACLRuleRateBasedStatementCustomKeyHeaderModel]         `tfsdk:"header"`
	HTTPMethod     fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]                                `tfsdk:"http_method"`
	IP             fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]                                `tfsdk:"ip"`
	LabelNamespace fwtypes.ListNestedObjectValueOf[webACLRuleRateBasedStatementCustomKeyLabelNamespaceModel] `tfsdk:"label_namespace"`
	QueryArgument  fwtypes.ListNestedObjectValueOf[webACLRuleRateBasedStatementCustomKeyQueryArgumentModel]  `tfsdk:"query_argument"`
	QueryString    fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]                                `tfsdk:"query_string"`
	UriPath        fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]                                `tfsdk:"uri_path"`
}

type webACLRuleRateBasedStatementCustomKeyCookieModel struct {
	Name                types.String                                                  `tfsdk:"name"`
	TextTransformations fwtypes.ListNestedObjectValueOf[webACLRuleTextTransformModel] `tfsdk:"text_transformation"`
}

type webACLRuleRateBasedStatementCustomKeyHeaderModel struct {
	Name                types.String                                                  `tfsdk:"name"`
	TextTransformations fwtypes.ListNestedObjectValueOf[webACLRuleTextTransformModel] `tfsdk:"text_transformation"`
}

type webACLRuleRateBasedStatementCustomKeyLabelNamespaceModel struct {
	Namespace types.String `tfsdk:"namespace"`
}

type webACLRuleRateBasedStatementCustomKeyQueryArgumentModel struct {
	Name                types.String                                                  `tfsdk:"name"`
	TextTransformations fwtypes.ListNestedObjectValueOf[webACLRuleTextTransformModel] `tfsdk:"text_transformation"`
}

type webACLRuleRateBasedStatementModel struct {
	Limit               types.Int64                                                                 `tfsdk:"limit"`
	AggregateKeyType    fwtypes.StringEnum[awstypes.RateBasedStatementAggregateKeyType]             `tfsdk:"aggregate_key_type"`
	CustomKeys          fwtypes.ListNestedObjectValueOf[webACLRuleRateBasedStatementCustomKeyModel] `tfsdk:"custom_keys"`
	EvaluationWindowSec types.Int64                                                                 `tfsdk:"evaluation_window_sec"`
	ForwardedIPConfig   fwtypes.ListNestedObjectValueOf[webACLRuleForwardedIPConfigModel]           `tfsdk:"forwarded_ip_config"`
	ScopeDownStatement  fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownStatementModel]          `tfsdk:"scope_down_statement"`
}

func (m webACLRuleRateBasedStatementModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	rbs := awstypes.RateBasedStatement{
		AggregateKeyType: m.AggregateKeyType.ValueEnum(),
	}

	diags.Append(flex.Expand(ctx, m.Limit, &rbs.Limit)...)
	diags.Append(flex.Expand(ctx, m.CustomKeys, &rbs.CustomKeys)...)
	diags.Append(flex.Expand(ctx, m.ForwardedIPConfig, &rbs.ForwardedIPConfig)...)
	diags.Append(flex.Expand(ctx, m.ScopeDownStatement, &rbs.ScopeDownStatement)...)

	// EvaluationWindowSec: only set if explicitly provided (not null)
	if !m.EvaluationWindowSec.IsNull() {
		rbs.EvaluationWindowSec = m.EvaluationWindowSec.ValueInt64()
	}

	return &rbs, diags
}

type webACLRuleByteMatchStatementModel struct {
	SearchString         types.String                                                  `tfsdk:"search_string"`
	PositionalConstraint fwtypes.StringEnum[awstypes.PositionalConstraint]             `tfsdk:"positional_constraint"`
	FieldToMatch         fwtypes.ListNestedObjectValueOf[webACLRuleFieldToMatchModel]  `tfsdk:"field_to_match"`
	TextTransformations  fwtypes.ListNestedObjectValueOf[webACLRuleTextTransformModel] `tfsdk:"text_transformation"`
}

type webACLRuleSqliMatchStatementModel struct {
}

type webACLRuleXssMatchStatementModel struct {
}

type webACLRuleSizeConstraintStatementModel struct {
}

type webACLRuleRegexMatchStatementModel struct {
	RegexString         types.String                                                  `tfsdk:"regex_string"`
	FieldToMatch        fwtypes.ListNestedObjectValueOf[webACLRuleFieldToMatchModel]  `tfsdk:"field_to_match"`
	TextTransformations fwtypes.ListNestedObjectValueOf[webACLRuleTextTransformModel] `tfsdk:"text_transformation"`
}

type webACLRuleLabelMatchStatementModel struct {
	Key   types.String                                 `tfsdk:"key"`
	Scope fwtypes.StringEnum[awstypes.LabelMatchScope] `tfsdk:"scope"`
}

type webACLRuleAsnMatchStatementModel struct {
	AsnList           fwtypes.ListValueOf[types.Int64]                                  `tfsdk:"asn_list"`
	ForwardedIPConfig fwtypes.ListNestedObjectValueOf[webACLRuleForwardedIPConfigModel] `tfsdk:"forwarded_ip_config"`
}

type webACLRuleLabelModel struct {
	Name types.String `tfsdk:"name"`
}

type webACLRuleImmunityTimeModel struct {
	ImmunityTimeProperty fwtypes.ListNestedObjectValueOf[webACLRuleImmunityTimePropertyModel] `tfsdk:"immunity_time_property"`
}

func (m webACLRuleImmunityTimeModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	cfg := &awstypes.CaptchaConfig{}
	if !m.ImmunityTimeProperty.IsNull() {
		itpData, d := m.ImmunityTimeProperty.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		cfg.ImmunityTimeProperty = &awstypes.ImmunityTimeProperty{
			ImmunityTime: itpData.ImmunityTime.ValueInt64Pointer(),
		}
	}
	return cfg, diags
}

func (m *webACLRuleImmunityTimeModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	m.ImmunityTimeProperty = fwtypes.NewListNestedObjectValueOfNull[webACLRuleImmunityTimePropertyModel](ctx)

	var itp *awstypes.ImmunityTimeProperty
	switch cfg := v.(type) {
	case *awstypes.CaptchaConfig:
		if cfg != nil {
			itp = cfg.ImmunityTimeProperty
		}
	case *awstypes.ChallengeConfig:
		if cfg != nil {
			itp = cfg.ImmunityTimeProperty
		}
	}

	if itp != nil {
		itpModel := webACLRuleImmunityTimePropertyModel{ImmunityTime: types.Int64PointerValue(itp.ImmunityTime)}
		m.ImmunityTimeProperty, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleImmunityTimePropertyModel{&itpModel}, nil)
	}
	return diags
}

type webACLRuleImmunityTimePropertyModel struct {
	ImmunityTime types.Int64 `tfsdk:"immunity_time"`
}

type webACLRuleOverrideActionModel struct {
	Count fwtypes.ListNestedObjectValueOf[webACLRuleOverrideActionEmptyModel] `tfsdk:"count"`
	None  fwtypes.ListNestedObjectValueOf[webACLRuleOverrideActionEmptyModel] `tfsdk:"none"`
}

func (m webACLRuleOverrideActionModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	oa := &awstypes.OverrideAction{}
	if !m.Count.IsNull() {
		oa.Count = &awstypes.CountAction{}
	} else if !m.None.IsNull() {
		oa.None = &awstypes.NoneAction{}
	} else {
		// Default to None if override_action block is present but no specific action is set
		oa.None = &awstypes.NoneAction{}
	}
	return oa, diags
}

func (m *webACLRuleOverrideActionModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	m.Count = fwtypes.NewListNestedObjectValueOfNull[webACLRuleOverrideActionEmptyModel](ctx)
	m.None = fwtypes.NewListNestedObjectValueOfNull[webACLRuleOverrideActionEmptyModel](ctx)

	oa, ok := v.(*awstypes.OverrideAction)
	if !ok || oa == nil {
		return diags
	}

	empty := webACLRuleOverrideActionEmptyModel{}
	if oa.Count != nil {
		m.Count, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleOverrideActionEmptyModel{&empty}, nil)
	} else if oa.None != nil {
		m.None, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleOverrideActionEmptyModel{&empty}, nil)
	}
	return diags
}

type webACLRuleOverrideActionEmptyModel struct{}

type webACLRuleTrulyEmptyModel struct{}

type webACLRuleFieldToMatchModel struct {
	AllQueryArguments   fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]          `tfsdk:"all_query_arguments"`
	Body                fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]          `tfsdk:"body"`
	Method              fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]          `tfsdk:"method"`
	QueryString         fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]          `tfsdk:"query_string"`
	SingleHeader        fwtypes.ListNestedObjectValueOf[webACLRuleSingleHeaderModel]        `tfsdk:"single_header"`
	SingleQueryArgument fwtypes.ListNestedObjectValueOf[webACLRuleSingleQueryArgumentModel] `tfsdk:"single_query_argument"`
	UriPath             fwtypes.ListNestedObjectValueOf[webACLRuleTrulyEmptyModel]          `tfsdk:"uri_path"`
}

type webACLRuleSingleHeaderModel struct {
	Name types.String `tfsdk:"name"`
}

type webACLRuleSingleQueryArgumentModel struct {
	Name types.String `tfsdk:"name"`
}

type webACLRuleTextTransformModel struct {
	Priority types.Int32                                         `tfsdk:"priority"`
	Type     fwtypes.StringEnum[awstypes.TextTransformationType] `tfsdk:"type"`
}

// Expand/Flatten helpers
