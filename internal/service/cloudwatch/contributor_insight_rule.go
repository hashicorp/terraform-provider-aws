// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudwatch

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_contributor_insight_rule", name="Contributor Insight Rule")
// @Tags(identifierAttribute="resource_arn")
// @IdentityAttribute("rule_name")
// @Testing(preIdentityVersion="v6.52.0")
// @Testing(importStateIdAttribute="rule_name")
// @Testing(importIgnore="rule_definition")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatch/types;types.InsightRule")
func newContributorInsightRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &contributorInsightRuleResource{}

	return r, nil
}

type contributorInsightRuleResource struct {
	framework.ResourceWithModel[contributorInsightRuleResourceModel]
	framework.WithImportByIdentity
}

func (r *contributorInsightRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: framework.ARNAttributeComputedOnly(),
			"rule_definition": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
			},
			"rule_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_state": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[stateValue](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *contributorInsightRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	c := r.Meta()
	conn := c.CloudWatchClient(ctx)

	var plan contributorInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	ruleName := fwflex.StringValueFromFramework(ctx, plan.RuleName)
	var input cloudwatch.PutInsightRuleInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	_, err := conn.PutInsightRule(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	// Taint the resource.
	resp.State.SetAttribute(ctx, path.Root("rule_name"), ruleName)

	switch ruleState := plan.RuleState.ValueEnum(); ruleState {
	case stateValueEnabled:
		if err := enableInsightRule(ctx, conn, ruleName); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ruleName)
			return
		}
	case stateValueDisabled:
		if err := disableInsightRule(ctx, conn, ruleName); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ruleName)
			return
		}
	}

	// Set values for unknowns.
	plan.ResourceARN = fwflex.StringValueToFramework(ctx, insightRuleARN(ctx, c, ruleName))
	if plan.RuleState.IsUnknown() {
		plan.RuleState = fwtypes.StringEnumValue(stateValueEnabled)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *contributorInsightRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	c := r.Meta()
	conn := c.CloudWatchClient(ctx)

	var state contributorInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	ruleName := fwflex.StringValueFromFramework(ctx, state.RuleName)
	out, err := findInsightRuleByName(ctx, conn, ruleName)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	// Set attributes for import.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	state.ResourceARN = fwflex.StringValueToFramework(ctx, insightRuleARN(ctx, c, ruleName))
	state.RuleState = fwtypes.StringEnumValue(stateValue(aws.ToString(out.State)))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *contributorInsightRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new contributorInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &old))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &new))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, new.RuleName)

	if !new.RuleDefinition.Equal(old.RuleDefinition) {
		input := cloudwatch.PutInsightRuleInput{
			RuleDefinition: fwflex.StringFromFramework(ctx, new.RuleDefinition),
			RuleName:       aws.String(ruleName),
		}
		_, err := conn.PutInsightRule(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ruleName)
			return
		}
	}

	if !new.RuleState.Equal(old.RuleState) {
		switch ruleState := new.RuleState.ValueEnum(); ruleState {
		case stateValueEnabled:
			if err := enableInsightRule(ctx, conn, ruleName); err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ruleName)
				return
			}
		case stateValueDisabled:
			if err := disableInsightRule(ctx, conn, ruleName); err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ruleName)
				return
			}
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &new))
}

func (r *contributorInsightRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state contributorInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	ruleName := fwflex.StringValueFromFramework(ctx, state.RuleName)
	input := cloudwatch.DeleteInsightRulesInput{
		RuleNames: []string{ruleName},
	}
	_, err := conn.DeleteInsightRules(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ruleName)
		return
	}
}

func insightRuleARN(ctx context.Context, c *conns.AWSClient, ruleName string) string {
	return c.RegionalARN(ctx, "cloudwatch", "insight-rule/"+ruleName)
}

func enableInsightRule(ctx context.Context, conn *cloudwatch.Client, ruleName string) error {
	input := cloudwatch.EnableInsightRulesInput{
		RuleNames: []string{ruleName},
	}
	output, err := conn.EnableInsightRules(ctx, &input)
	if err == nil {
		err = partialFailuresError(output.Failures)
	}
	if err != nil {
		return smarterr.NewError(err)
	}
	return nil
}

func disableInsightRule(ctx context.Context, conn *cloudwatch.Client, ruleName string) error {
	input := cloudwatch.DisableInsightRulesInput{
		RuleNames: []string{ruleName},
	}
	output, err := conn.DisableInsightRules(ctx, &input)
	if err == nil {
		err = partialFailuresError(output.Failures)
	}
	if err != nil {
		return smarterr.NewError(err)
	}
	return nil
}

func findInsightRuleByName(ctx context.Context, conn *cloudwatch.Client, name string) (*awstypes.InsightRule, error) {
	var input cloudwatch.DescribeInsightRulesInput

	return findInsightRule(ctx, conn, &input, func(v awstypes.InsightRule) bool {
		return aws.ToString(v.Name) == name
	})
}

func findInsightRule(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.DescribeInsightRulesInput, filter tfslices.Predicate[awstypes.InsightRule]) (*awstypes.InsightRule, error) {
	output, err := findInsightRules(ctx, conn, input, filter)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findInsightRules(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.DescribeInsightRulesInput, filter tfslices.Predicate[awstypes.InsightRule]) ([]awstypes.InsightRule, error) {
	var output []awstypes.InsightRule

	pages := cloudwatch.NewDescribeInsightRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.InsightRules {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

type contributorInsightRuleResourceModel struct {
	framework.WithRegionModel
	ResourceARN    types.String                   `tfsdk:"resource_arn"`
	RuleDefinition jsontypes.Normalized           `tfsdk:"rule_definition"`
	RuleName       types.String                   `tfsdk:"rule_name"`
	RuleState      fwtypes.StringEnum[stateValue] `tfsdk:"rule_state"`
	Tags           tftags.Map                     `tfsdk:"tags"`
	TagsAll        tftags.Map                     `tfsdk:"tags_all"`
}
