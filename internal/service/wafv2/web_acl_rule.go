// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_wafv2_web_acl_rule", name="Web ACL Rule")
func newResourceWebACLRule(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceWebACLRule{}, nil
}

const (
	ResNameWebACLRule = "Web ACL Rule"
)

type resourceWebACLRule struct {
	framework.ResourceWithModel[webACLRuleModel]
	framework.WithImportByID
}

func (r *resourceWebACLRule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
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
					listvalidator.SizeAtLeast(1),
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
							CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleEmptyModel](ctx),
							Validators:   []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{},
						},
						"challenge": schema.ListNestedBlock{
							CustomType:   fwtypes.NewListNestedObjectTypeOf[webACLRuleEmptyModel](ctx),
							Validators:   []validator.List{listvalidator.SizeAtMost(1)},
							NestedObject: schema.NestedBlockObject{},
						},
					},
				},
				Description: "Action to take when the rule matches.",
			},
			"statement": statementBlock(ctx),
			"visibility_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleVisibilityConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
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

func statementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleStatementModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"ip_set_reference_statement": ipSetReferenceStatementBlock(ctx),
				"geo_match_statement":        geoMatchStatementBlock(ctx),
			},
		},
		Description: "Rule statement. Exactly one statement type must be specified.",
	}
}

func ipSetReferenceStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleIPSetReferenceStatementModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("geo_match_statement")),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrARN: schema.StringAttribute{
					Required:    true,
					Description: "ARN of the IP set to reference.",
				},
			},
			Blocks: map[string]schema.Block{
				"ip_set_forwarded_ip_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleIPSetForwardedIPConfigModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fallback_behavior": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf("MATCH", "NO_MATCH"),
								},
							},
							"header_name": schema.StringAttribute{
								Required: true,
							},
							"position": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf("FIRST", "LAST", "ANY"),
								},
							},
						},
					},
				},
			},
		},
		Description: "IP set reference statement.",
	}
}

func geoMatchStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleGeoMatchStatementModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("ip_set_reference_statement")),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"country_codes": schema.ListAttribute{
					ElementType: types.StringType,
					Required:    true,
					Description: "List of two-character country codes (e.g., US, CA).",
				},
			},
			Blocks: map[string]schema.Block{
				"forwarded_ip_config": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleForwardedIPConfigModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"fallback_behavior": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.OneOf("MATCH", "NO_MATCH"),
								},
							},
							"header_name": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
		},
		Description: "Geo match statement.",
	}
}

func customRequestHandlingBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCustomRequestHandlingModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"insert_header": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCustomHeaderModel](ctx),
					Validators: []validator.List{listvalidator.SizeAtLeast(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName:  schema.StringAttribute{Required: true},
							names.AttrValue: schema.StringAttribute{Required: true},
						},
					},
				},
			},
		},
	}
}

func customResponseBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCustomResponseModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"response_code": schema.Int32Attribute{
					Required: true,
				},
				"custom_response_body_key": schema.StringAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"response_header": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleCustomHeaderModel](ctx),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrName:  schema.StringAttribute{Required: true},
							names.AttrValue: schema.StringAttribute{Required: true},
						},
					},
				},
			},
		},
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

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.WebACLARN.ValueString(), plan.Name.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
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

func (r *resourceWebACLRule) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: web_acl_arn/rule_name
	parts := splitImportID(req.ID)
	if parts == nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: web_acl_arn/rule_name. Got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("web_acl_arn"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrName), parts[1])...)
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func splitImportID(id string) []string {
	// Find the last occurrence of "/" to split ARN from rule name
	lastSlash := -1
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '/' {
			lastSlash = i
			break
		}
	}
	if lastSlash == -1 || lastSlash == 0 || lastSlash == len(id)-1 {
		return nil
	}
	return []string{id[:lastSlash], id[lastSlash+1:]}
}

// Models

type webACLRuleModel struct {
	framework.WithRegionModel
	ID               types.String                                                     `tfsdk:"id"`
	Name             types.String                                                     `tfsdk:"name"`
	Priority         types.Int32                                                      `tfsdk:"priority"`
	WebACLARN        types.String                                                     `tfsdk:"web_acl_arn"`
	Action           fwtypes.ListNestedObjectValueOf[webACLRuleActionModel]           `tfsdk:"action"`
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
		return &awstypes.RuleAction{Allow: &awstypes.AllowAction{}}, diags
	case !m.Block.IsNull():
		return &awstypes.RuleAction{Block: &awstypes.BlockAction{}}, diags
	case !m.Count.IsNull():
		return &awstypes.RuleAction{Count: &awstypes.CountAction{}}, diags
	case !m.Captcha.IsNull():
		return &awstypes.RuleAction{Captcha: &awstypes.CaptchaAction{}}, diags
	case !m.Challenge.IsNull():
		return &awstypes.RuleAction{Challenge: &awstypes.ChallengeAction{}}, diags
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
		m.Allow, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleEmptyModel{emptyVal}, nil)
	case action.Block != nil:
		blockVal := &webACLRuleBlockActionModel{
			CustomResponse: fwtypes.NewListNestedObjectValueOfNull[webACLRuleCustomResponseModel](ctx),
		}
		m.Block, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleBlockActionModel{blockVal}, nil)
	case action.Count != nil:
		emptyVal := &webACLRuleEmptyModel{
			CustomRequestHandling: fwtypes.NewListNestedObjectValueOfNull[webACLRuleCustomRequestHandlingModel](ctx),
		}
		m.Count, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleEmptyModel{emptyVal}, nil)
	case action.Captcha != nil:
		emptyVal := &webACLRuleEmptyModel{
			CustomRequestHandling: fwtypes.NewListNestedObjectValueOfNull[webACLRuleCustomRequestHandlingModel](ctx),
		}
		m.Captcha, diags = fwtypes.NewListNestedObjectValueOfSlice(ctx, []*webACLRuleEmptyModel{emptyVal}, nil)
	case action.Challenge != nil:
		emptyVal := &webACLRuleEmptyModel{
			CustomRequestHandling: fwtypes.NewListNestedObjectValueOfNull[webACLRuleCustomRequestHandlingModel](ctx),
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
	IPSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel] `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement       fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]       `tfsdk:"geo_match_statement"`
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

	// Initialize both to null
	m.IPSetReferenceStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleIPSetReferenceStatementModel](ctx)
	m.GeoMatchStatement = fwtypes.NewListNestedObjectValueOfNull[webACLRuleGeoMatchStatementModel](ctx)

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

// Expand/Flatten helpers
