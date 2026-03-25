// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudwatch

import (
	"context"
	"errors"
	"fmt"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
// @Testing(importStateIdFunc="testAccContributorInsightRuleImportStateIDFunc")
// @Testing(importStateIdAttribute="rule_name")
// @Testing(importIgnore="rule_definition;rule_state")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatch/types;types.InsightRule")
func newContributorInsightRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &contributorInsightRuleResource{}

	return r, nil
}

const (
	ResNameContributorInsightRule = "Contributor Insight Rule"
)

type contributorInsightRuleResource struct {
	framework.ResourceWithModel[contributorInsightRuleResourceModel]
}

func (r *contributorInsightRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: framework.ARNAttributeComputedOnly(),
			"rule_definition": schema.StringAttribute{
				Required: true,
			},
			"rule_name": schema.StringAttribute{
				Required: true,
			},
			"rule_state": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[stateValue](),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *contributorInsightRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var plan contributorInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudwatch.PutInsightRuleInput{
		RuleDefinition: plan.RuleDefinition.ValueStringPointer(),
		RuleName:       plan.RuleName.ValueStringPointer(),
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.PutInsightRule(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RuleName.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.RuleName.String())
		return
	}

	cirARN := r.Meta().RegionalARN(ctx, "cloudwatch", fmt.Sprintf("insight-rule/%s", plan.RuleName.ValueString()))
	plan.ResourceARN = fwflex.StringValueToFramework(ctx, cirARN)

	if !plan.RuleState.IsNull() {
		if plan.RuleState.ValueEnum() == stateValueEnabled {
			input := cloudwatch.EnableInsightRulesInput{
				RuleNames: []string{plan.RuleName.ValueString()},
			}
			_, err = conn.EnableInsightRules(ctx, &input)
		}

		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RuleName.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *contributorInsightRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state contributorInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findContributorInsightRuleByName(ctx, conn, state.RuleName.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RuleName.String())
		return
	}

	cirARN := r.Meta().RegionalARN(ctx, "cloudwatch", fmt.Sprintf("insight-rule/%s", state.RuleName.ValueString()))
	state.ResourceARN = fwflex.StringValueToFramework(ctx, cirARN)

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state), smerr.ID, state.RuleName.String())
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state), smerr.ID, state.RuleName.String())
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

	if !new.RuleState.IsNull() && !old.RuleState.Equal(new.RuleState) {
		if new.RuleState.ValueEnum() == stateValueEnabled {
			input := cloudwatch.EnableInsightRulesInput{
				RuleNames: []string{new.RuleName.ValueString()},
			}
			_, err := conn.EnableInsightRules(ctx, &input)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, new.RuleName.String())
			}
		} else if new.RuleState.ValueEnum() == stateValueDisabled {
			input := cloudwatch.DisableInsightRulesInput{
				RuleNames: []string{new.RuleName.ValueString()},
			}
			_, err := conn.DisableInsightRules(ctx, &input)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, new.RuleName.String())
			}
		}
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &new), smerr.ID, new.RuleName.String())
}

func (r *contributorInsightRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state contributorInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	in := &cloudwatch.DeleteInsightRulesInput{
		RuleNames: []string{state.RuleName.ValueString()},
	}

	_, err := conn.DeleteInsightRules(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RuleName.String())
		return
	}
}

func (r *contributorInsightRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule_name"), req, resp)
}

func findContributorInsightRuleByName(ctx context.Context, conn *cloudwatch.Client, name string) (*awstypes.InsightRule, error) {
	input := &cloudwatch.DescribeInsightRulesInput{}

	return findContributorInsightRule(ctx, conn, input, func(v *awstypes.InsightRule) bool {
		return aws.ToString(v.Name) == name
	})
}

func findContributorInsightRule(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.DescribeInsightRulesInput, filter tfslices.Predicate[*awstypes.InsightRule]) (*awstypes.InsightRule, error) {
	output, err := findContributorInsightRules(ctx, conn, input, filter)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findContributorInsightRules(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.DescribeInsightRulesInput, filter tfslices.Predicate[*awstypes.InsightRule]) ([]awstypes.InsightRule, error) {
	var output []awstypes.InsightRule

	paginator := cloudwatch.NewDescribeInsightRulesPaginator(conn, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.InsightRules {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

type contributorInsightRuleResourceModel struct {
	framework.WithRegionModel
	ResourceARN    types.String                   `tfsdk:"resource_arn"`
	RuleDefinition types.String                   `tfsdk:"rule_definition"`
	RuleName       types.String                   `tfsdk:"rule_name"`
	RuleState      fwtypes.StringEnum[stateValue] `tfsdk:"rule_state"`
	Tags           tftags.Map                     `tfsdk:"tags"`
	TagsAll        tftags.Map                     `tfsdk:"tags_all"`
}
