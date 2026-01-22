// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_security_group_rules_exclusive", name="Security Group Rules Exclusive")
func newResourceSecurityGroupRulesExclusive(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceSecurityGroupRulesExclusive{}, nil
}

const (
	ResNameSecurityGroupRulesExclusive = "Security Group Rules Exclusive"
)

type resourceSecurityGroupRulesExclusive struct {
	framework.ResourceWithModel[resourceSecurityGroupRulesExclusiveData]
	framework.WithNoOpDelete
}

func (r *resourceSecurityGroupRulesExclusive) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"egress_rule_ids": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Required:    true,
				ElementType: types.StringType,
			},
			"ingress_rule_ids": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Required:    true,
				ElementType: types.StringType,
			},
			"security_group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceSecurityGroupRulesExclusive) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceSecurityGroupRulesExclusiveData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ingressRuleIDs, egressRuleIDs []string
	resp.Diagnostics.Append(plan.IngressRuleIDs.ElementsAs(ctx, &ingressRuleIDs, false)...)
	resp.Diagnostics.Append(plan.EgressRuleIDs.ElementsAs(ctx, &egressRuleIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.syncRules(ctx, &resp.Diagnostics, plan.SecurityGroupID.ValueString(), ingressRuleIDs, egressRuleIDs)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameSecurityGroupRulesExclusive, plan.SecurityGroupID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSecurityGroupRulesExclusive) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceSecurityGroupRulesExclusiveData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The ingress/egress rule finder below will simply return empty arrays for deleted
	// security groups. To determine whether this resource should be removed from state,
	// check for presence of the security group first.
	if _, err := findSecurityGroupByID(ctx, conn, state.SecurityGroupID.ValueString()); retry.NotFound(err) { // nosemgrep:ci.semgrep.errors.notfound-without-err-checks
		resp.State.RemoveResource(ctx)
		return
	}

	ingressRuleIDs, egressRuleIDs, err := findSecurityGroupRuleIDsBySecurityGroupID(ctx, conn, state.SecurityGroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, ResNameSecurityGroupRulesExclusive, state.SecurityGroupID.String(), err),
			err.Error(),
		)
		return
	}

	state.IngressRuleIDs = fwflex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, ingressRuleIDs)
	state.EgressRuleIDs = fwflex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, egressRuleIDs)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSecurityGroupRulesExclusive) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceSecurityGroupRulesExclusiveData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.IngressRuleIDs.Equal(state.IngressRuleIDs) || !plan.EgressRuleIDs.Equal(state.EgressRuleIDs) {
		var ingressRuleIDs, egressRuleIDs []string
		resp.Diagnostics.Append(plan.IngressRuleIDs.ElementsAs(ctx, &ingressRuleIDs, false)...)
		resp.Diagnostics.Append(plan.EgressRuleIDs.ElementsAs(ctx, &egressRuleIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		err := r.syncRules(ctx, &resp.Diagnostics, plan.SecurityGroupID.ValueString(), ingressRuleIDs, egressRuleIDs)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.EC2, create.ErrActionUpdating, ResNameSecurityGroupRulesExclusive, plan.SecurityGroupID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// syncRules handles keeping the configured security group rules in sync with
// the remote resource.
//
// Rules defined on this resource but not present in the security group will
// generate warnings directing users to create them. Rules present in the security
// group but not configured on this resource will be removed.
func (r *resourceSecurityGroupRulesExclusive) syncRules(ctx context.Context, diags *diag.Diagnostics, securityGroupID string, wantIngress, wantEgress []string) error {
	conn := r.Meta().EC2Client(ctx)

	haveIngress, haveEgress, err := findSecurityGroupRuleIDsBySecurityGroupID(ctx, conn, securityGroupID)
	if err != nil {
		return err
	}

	createIngress, removeIngress, _ := intflex.DiffSlices(haveIngress, wantIngress, func(s1, s2 string) bool { return s1 == s2 })
	createEgress, removeEgress, _ := intflex.DiffSlices(haveEgress, wantEgress, func(s1, s2 string) bool { return s1 == s2 })

	// Emit warnings for rules that need to be created
	for _, ruleID := range createIngress {
		diags.AddWarning(
			"Ingress Rule Not Found",
			fmt.Sprintf("Security group rule %q is configured but not currently associated with security group %q. "+
				"Use the aws_vpc_security_group_ingress_rule resource to create this rule.", ruleID, securityGroupID),
		)
	}

	for _, ruleID := range createEgress {
		diags.AddWarning(
			"Egress Rule Not Found",
			fmt.Sprintf("Security group rule %q is configured but not currently associated with security group %q. "+
				"Use the aws_vpc_security_group_egress_rule resource to create this rule.", ruleID, securityGroupID),
		)
	}

	for _, ruleID := range removeIngress {
		input := ec2.RevokeSecurityGroupIngressInput{
			GroupId:              aws.String(securityGroupID),
			SecurityGroupRuleIds: []string{ruleID},
		}
		if _, err := conn.RevokeSecurityGroupIngress(ctx, &input); err != nil {
			return err
		}
	}

	for _, ruleID := range removeEgress {
		input := ec2.RevokeSecurityGroupEgressInput{
			GroupId:              aws.String(securityGroupID),
			SecurityGroupRuleIds: []string{ruleID},
		}
		if _, err := conn.RevokeSecurityGroupEgress(ctx, &input); err != nil {
			return err
		}
	}

	return nil
}

func (r *resourceSecurityGroupRulesExclusive) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("security_group_id"), req, resp)
}

func findSecurityGroupRuleIDsBySecurityGroupID(ctx context.Context, conn *ec2.Client, id string) ([]string, []string, error) {
	rules, err := findSecurityGroupRulesBySecurityGroupID(ctx, conn, id)
	if err != nil {
		return nil, nil, err
	}

	var ingressRuleIDs, egressRuleIDs []string
	for _, rule := range rules {
		if rule.SecurityGroupRuleId != nil {
			if aws.ToBool(rule.IsEgress) {
				egressRuleIDs = append(egressRuleIDs, aws.ToString(rule.SecurityGroupRuleId))
			} else {
				ingressRuleIDs = append(ingressRuleIDs, aws.ToString(rule.SecurityGroupRuleId))
			}
		}
	}

	return ingressRuleIDs, egressRuleIDs, nil
}

type resourceSecurityGroupRulesExclusiveData struct {
	framework.WithRegionModel
	EgressRuleIDs   fwtypes.SetOfString `tfsdk:"egress_rule_ids"`
	IngressRuleIDs  fwtypes.SetOfString `tfsdk:"ingress_rule_ids"`
	SecurityGroupID types.String        `tfsdk:"security_group_id"`
}
