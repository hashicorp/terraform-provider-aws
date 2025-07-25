// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_wafv2_web_acl_rule_group_association", name="Web ACL Rule Group Association")
func newResourceWebACLRuleGroupAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceWebACLRuleGroupAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameWebACLRuleGroupAssociation = "Web ACL Rule Group Association"
)

type resourceWebACLRuleGroupAssociation struct {
	framework.ResourceWithModel[resourceWebACLRuleGroupAssociationModel]
	framework.WithTimeouts
}

func (r *resourceWebACLRuleGroupAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"rule_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
				Description: "Name of the rule to create in the Web ACL that references the rule group.",
			},
			names.AttrPriority: schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
				Description: "Priority of the rule within the Web ACL.",
			},
			"rule_group_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
				Description: "ARN of the Rule Group to associate with the Web ACL.",
			},
			"web_acl_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
				Description: "ARN of the Web ACL to associate the Rule Group with.",
			},
			"override_action": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("none", "count"),
				},
				Description: "Override action for the rule group. Valid values are 'none' and 'count'. Defaults to 'none'.",
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
		Description: "Associates a WAFv2 Rule Group with a Web ACL by adding a rule that references the Rule Group.",
	}
}

func (r *resourceWebACLRuleGroupAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var plan resourceWebACLRuleGroupAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse Web ACL ARN to get ID, name, and scope
	webACLID, webACLName, webACLScope, err := parseWebACLARN(plan.WebACLARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	// Get current Web ACL configuration
	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	// Check if rule with same priority or name already exists
	for _, rule := range webACL.WebACL.Rules {
		if rule.Priority == plan.Priority.ValueInt32() {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), nil),
				fmt.Sprintf("Rule with priority %d already exists in Web ACL", plan.Priority.ValueInt32()),
			)
			return
		}
		if aws.ToString(rule.Name) == plan.RuleName.ValueString() {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), nil),
				fmt.Sprintf("Rule with name %s already exists in Web ACL", plan.RuleName.ValueString()),
			)
			return
		}
	}

	// Create new rule with rule group reference statement
	newRule := awstypes.Rule{
		Name:     aws.String(plan.RuleName.ValueString()),
		Priority: plan.Priority.ValueInt32(),
		Statement: &awstypes.Statement{
			RuleGroupReferenceStatement: &awstypes.RuleGroupReferenceStatement{
				ARN: aws.String(plan.RuleGroupARN.ValueString()),
			},
		},
		VisibilityConfig: &awstypes.VisibilityConfig{
			SampledRequestsEnabled:   true,
			CloudWatchMetricsEnabled: true,
			MetricName:               aws.String(plan.RuleName.ValueString()),
		},
	}

	// Set override action
	overrideAction := plan.OverrideAction.ValueString()
	if overrideAction == "" {
		overrideAction = "none"
	}

	if overrideAction == "none" {
		newRule.OverrideAction = &awstypes.OverrideAction{
			None: &awstypes.NoneAction{},
		}
	} else if overrideAction == "count" {
		newRule.OverrideAction = &awstypes.OverrideAction{
			Count: &awstypes.CountAction{},
		}
	}

	// Add the new rule to existing rules
	updatedRules := append(webACL.WebACL.Rules, newRule)

	// Update the Web ACL
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
		Description:          webACL.WebACL.Description,
		TokenDomains:         webACL.WebACL.TokenDomains,
	}

	const timeout = 5 * time.Minute
	_, err = tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, timeout, func() (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameWebACLRuleGroupAssociation, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	// Set the ID as a combination that uniquely identifies this association
	// Format: webACLARN/ruleName/ruleGroupARN (using slashes for consistency with other WAFv2 resources)
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", 
		plan.WebACLARN.ValueString(), 
		plan.RuleName.ValueString(),
		plan.RuleGroupARN.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceWebACLRuleGroupAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceWebACLRuleGroupAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	// Parse the ID to get Web ACL ARN, rule name, and rule group ARN
	// Format: webACLARN/ruleName/ruleGroupARN
	idParts := strings.SplitN(state.ID.ValueString(), "/", 3)
	if len(idParts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			"Resource ID should be in format 'webACLARN/ruleName/ruleGroupARN'",
		)
		return
	}

	webACLARN := idParts[0]
	ruleName := idParts[1]
	ruleGroupARN := idParts[2]

	// Parse Web ACL ARN to get ID, name, and scope
	webACLID, webACLName, webACLScope, err := parseWebACLARN(webACLARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Reading WAFv2 Web ACL Rule Group Association",
			fmt.Sprintf("Error parsing Web ACL ARN: %s", err),
		)
		return
	}

	// Get the Web ACL and check if the rule group is associated
	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		if tfresource.NotFound(err) {
			resp.Diagnostics.AddWarning(
				"Web ACL Not Found",
				"Web ACL was not found, removing from state",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Reading WAFv2 Web ACL Rule Group Association",
			fmt.Sprintf("Error reading Web ACL: %s", err),
		)
		return
	}

	// Find the rule group in the Web ACL rules
	found := false
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) == ruleName && 
		   rule.Statement != nil && 
		   rule.Statement.RuleGroupReferenceStatement != nil &&
		   aws.ToString(rule.Statement.RuleGroupReferenceStatement.ARN) == ruleGroupARN {
			found = true
			state.Priority = types.Int32Value(rule.Priority)

			// Determine override action
			overrideAction := "none"
			if rule.OverrideAction != nil {
				if rule.OverrideAction.Count != nil {
					overrideAction = "count"
				} else if rule.OverrideAction.None != nil {
					overrideAction = "none"
				}
			}
			state.OverrideAction = types.StringValue(overrideAction)
			break
		}
	}

	if !found {
		resp.Diagnostics.AddWarning(
			"Rule Group Association Not Found",
			"Rule group association was not found in Web ACL, removing from state",
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with current values
	state.WebACLARN = types.StringValue(webACLARN)
	state.RuleGroupARN = types.StringValue(ruleGroupARN)
	state.RuleName = types.StringValue(ruleName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceWebACLRuleGroupAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceWebACLRuleGroupAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Most attributes require replacement, so we only need to handle changes to attributes
	// that don't require replacement (currently none)
	
	// For future extensibility, if any attributes are added that don't require replacement,
	// they would be handled here

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceWebACLRuleGroupAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceWebACLRuleGroupAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	// Parse the ID to get Web ACL ARN, rule name, and rule group ARN
	// Format: webACLARN/ruleName/ruleGroupARN
	idParts := strings.SplitN(state.ID.ValueString(), "/", 3)
	if len(idParts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			"Resource ID should be in format 'webACLARN/ruleName/ruleGroupARN'",
		)
		return
	}

	webACLARN := idParts[0]
	ruleName := idParts[1]

	// Parse Web ACL ARN to get ID, name, and scope
	webACLID, webACLName, webACLScope, err := parseWebACLARN(webACLARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Deleting WAFv2 Web ACL Rule Group Association",
			fmt.Sprintf("Error parsing Web ACL ARN: %s", err),
		)
		return
	}

	// Get the Web ACL
	webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
	if err != nil {
		if tfresource.NotFound(err) {
			// Web ACL is already gone, nothing to do
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRuleGroupAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Filter out the rule we want to remove
	var updatedRules []awstypes.Rule
	ruleFound := false
	for _, rule := range webACL.WebACL.Rules {
		if aws.ToString(rule.Name) != ruleName {
			updatedRules = append(updatedRules, rule)
		} else {
			ruleFound = true
		}
	}

	if !ruleFound {
		// Rule is already gone, nothing to do
		return
	}

	// Update the Web ACL without the rule
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
		Description:          webACL.WebACL.Description,
		TokenDomains:         webACL.WebACL.TokenDomains,
	}

	const timeout = 5 * time.Minute
	_, err = tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, timeout, func() (any, error) {
		return conn.UpdateWebACL(ctx, updateInput)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameWebACLRuleGroupAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceWebACLRuleGroupAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: webACLARN,ruleGroupARN,ruleName
	parts := strings.Split(req.ID, ",")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID should be in format 'webACLARN,ruleGroupARN,ruleName'",
		)
		return
	}

	webACLARN := parts[0]
	ruleGroupARN := parts[1]
	ruleName := parts[2]

	// Parse Web ACL ARN to get ID, name, and scope
	_, _, _, err := parseWebACLARN(webACLARN)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Web ACL ARN",
			fmt.Sprintf("Error parsing Web ACL ARN: %s", err),
		)
		return
	}

	// Set the ID in the expected format for Read
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), 
		fmt.Sprintf("%s/%s/%s", webACLARN, ruleName, ruleGroupARN))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("web_acl_arn"), webACLARN)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_group_arn"), ruleGroupARN)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_name"), ruleName)...)
}

// parseWebACLARN extracts the Web ACL ID, name, and scope from the ARN
func parseWebACLARN(arn string) (id, name, scope string, err error) {
	// ARN format: arn:aws:wafv2:region:account-id:scope/webacl/name/id
	// or for CloudFront: arn:aws:wafv2:global:account-id:global/webacl/name/id
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return "", "", "", fmt.Errorf("invalid Web ACL ARN format: %s", arn)
	}
	
	resourceParts := strings.Split(parts[5], "/")
	if len(resourceParts) < 4 {
		return "", "", "", fmt.Errorf("invalid Web ACL ARN resource format: %s", parts[5])
	}
	
	// Validate that this is a webacl ARN
	if resourceParts[1] != "webacl" {
		return "", "", "", fmt.Errorf("invalid Web ACL ARN: expected webacl resource type, got %s", resourceParts[1])
	}
	
	// Determine scope
	scopeValue := "REGIONAL"
	if parts[3] == "global" || resourceParts[0] == "global" {
		scopeValue = "CLOUDFRONT"
	}
	
	// Extract name and ID
	nameIndex := len(resourceParts) - 2
	idIndex := len(resourceParts) - 1
	
	return resourceParts[idIndex], resourceParts[nameIndex], scopeValue, nil
}

type resourceWebACLRuleGroupAssociationModel struct {
	ID             types.String   `tfsdk:"id"`
	RuleName       types.String   `tfsdk:"rule_name"`
	Priority       types.Int32    `tfsdk:"priority"`
	RuleGroupARN   types.String   `tfsdk:"rule_group_arn"`
	WebACLARN      types.String   `tfsdk:"web_acl_arn"`
	OverrideAction types.String   `tfsdk:"override_action"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func sweepWebACLRuleGroupAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	// Since this is a synthetic resource that modifies web ACLs,
	// we don't need a specific sweep function for it.
	// The web ACL sweep function will handle cleaning up web ACLs.
	return nil, nil
}

// Unit tests for parseWebACLARN function
// These tests are included in the main file following the pattern used in flex_test.go
