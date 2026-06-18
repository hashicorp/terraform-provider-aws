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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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

type stateValue string

func (v stateValue) Values() []stateValue {
	return []stateValue{
		stateValueEnabled,
		stateValueDisabled,
	}
}

const (
	stateValueEnabled  stateValue = "ENABLED"
	stateValueDisabled stateValue = "DISABLED"
)

// @FrameworkResource("aws_cloudwatch_contributor_managed_insight_rule", name="Contributor Managed Insight Rule")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newContributorManagedInsightRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &contributorManagedInsightRuleResource{}

	return r, nil
}

const (
	ResNameContributorManagedInsightRule = "Contributor Managed Insight Rule"
)

type contributorManagedInsightRuleResource struct {
	framework.ResourceWithModel[contributorManagedInsightRuleResourceModel]
}

func (r *contributorManagedInsightRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrResourceARN: schema.StringAttribute{
				Required: true,
			},
			"template_name": schema.StringAttribute{
				Required: true,
			},
			names.AttrState: schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				CustomType: fwtypes.StringEnumType[stateValue](),
				Default:    stringdefault.StaticString(string(stateValueEnabled)),
			},
			"rule_name": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *contributorManagedInsightRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var plan contributorManagedInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudwatch.PutManagedInsightRulesInput{
		ManagedRules: []awstypes.ManagedRule{
			{
				ResourceARN:  plan.ResourceArn.ValueStringPointer(),
				TemplateName: plan.TemplateName.ValueStringPointer(),
			},
		},
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ManagedRules[0].Tags = getTagsIn(ctx)

	if plan.State.ValueEnum() == stateValueEnabled || plan.State.IsNull() {
		out, err := conn.PutManagedInsightRules(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceArn.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ResourceArn.String())
			return
		}
		rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, plan.ResourceArn.ValueString(), plan.TemplateName.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceArn.String())
			return
		}

		plan.RuleName = fwflex.StringToFramework(ctx, rule.RuleState.RuleName)

		cmirARN := r.Meta().RegionalARN(ctx, "cloudwatch", fmt.Sprintf("insight-rule/%s", plan.RuleName.ValueString()))
		plan.ARN = fwflex.StringValueToFramework(ctx, cmirARN)
	} else if plan.State.ValueEnum() == stateValueDisabled {
		rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, plan.ResourceArn.ValueString(), plan.TemplateName.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceArn.String())
			return
		}
		input := cloudwatch.DisableInsightRulesInput{
			RuleNames: []string{*rule.RuleState.RuleName},
		}
		_, err = conn.DisableInsightRules(ctx, &input)

		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceArn.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *contributorManagedInsightRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state contributorManagedInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, state.ResourceArn.ValueString(), state.TemplateName.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ResourceArn.String())
		return
	}

	if out.RuleState != nil && out.RuleState.RuleName != nil {
		state.RuleName = fwflex.StringToFramework(ctx, out.RuleState.RuleName)
		cmirARN := r.Meta().RegionalARN(ctx, "cloudwatch", fmt.Sprintf("insight-rule/%s", state.RuleName.ValueString()))
		state.ARN = fwflex.StringValueToFramework(ctx, cmirARN)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state), smerr.ID, state.ResourceArn.String())
}

func (r *contributorManagedInsightRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new contributorManagedInsightRuleResourceModel

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &old))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &new))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	if !new.State.IsNull() && !old.State.Equal(new.State) {
		if new.State.ValueEnum() == stateValueEnabled || new.State.IsNull() {
			input := cloudwatch.PutManagedInsightRulesInput{
				ManagedRules: []awstypes.ManagedRule{
					{
						ResourceARN:  new.ResourceArn.ValueStringPointer(),
						TemplateName: new.TemplateName.ValueStringPointer(),
					},
				},
			}
			_, err := conn.PutManagedInsightRules(ctx, &input)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, new.ResourceArn.String())
			}
		} else if new.State.ValueEnum() == stateValueDisabled {
			rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, new.ResourceArn.ValueString(), new.TemplateName.ValueString())
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, new.ResourceArn.String())
			}
			input := cloudwatch.DisableInsightRulesInput{
				RuleNames: []string{aws.ToString(rule.RuleState.RuleName)},
			}
			_, err = conn.DisableInsightRules(ctx, &input)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, new.ResourceArn.String())
			}
		}
	}

	new.RuleName = old.RuleName

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &new))
}

func (r *contributorManagedInsightRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state contributorManagedInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, state.ResourceArn.ValueString(), state.TemplateName.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ResourceArn.String())
		return
	}

	if rule.RuleState == nil || rule.RuleState.RuleName == nil {
		return
	}

	input := cloudwatch.DeleteInsightRulesInput{
		RuleNames: []string{aws.ToString(rule.RuleState.RuleName)},
	}
	_, err = conn.DeleteInsightRules(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ResourceArn.String())
		return
	}
}

func (r *contributorManagedInsightRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	const resourceIDParts = 2
	idParts, err := flex.ExpandResourceId(req.ID, resourceIDParts, false)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, req.ID)
		return
	}
	resourceARN := idParts[0]
	templateName := idParts[1]

	cmir, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, r.Meta().CloudWatchClient(ctx), resourceARN, templateName)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrResourceARN), aws.ToString(cmir.ResourceARN)), smerr.ID, resourceARN)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root("template_name"), templateName), smerr.ID, resourceARN)
}

type contributorManagedInsightRuleResourceModel struct {
	framework.WithRegionModel
	ARN          types.String                   `tfsdk:"arn"`
	ResourceArn  types.String                   `tfsdk:"resource_arn"`
	TemplateName types.String                   `tfsdk:"template_name"`
	State        fwtypes.StringEnum[stateValue] `tfsdk:"state"`
	RuleName     types.String                   `tfsdk:"rule_name"`
	Tags         tftags.Map                     `tfsdk:"tags"`
	TagsAll      tftags.Map                     `tfsdk:"tags_all"`
}

func findContributorManagedInsightRules(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.ListManagedInsightRulesInput, filter tfslices.Predicate[*awstypes.ManagedRuleDescription]) ([]awstypes.ManagedRuleDescription, error) {
	var output []awstypes.ManagedRuleDescription

	pages := cloudwatch.NewListManagedInsightRulesPaginator(conn, input)
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

		for _, rule := range page.ManagedRules {
			if filter(&rule) {
				output = append(output, rule)
			}
		}
	}

	return output, nil
}

func findContributorManagedInsightRule(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.ListManagedInsightRulesInput, filter tfslices.Predicate[*awstypes.ManagedRuleDescription]) (*awstypes.ManagedRuleDescription, error) {
	output, err := findContributorManagedInsightRules(ctx, conn, input, filter)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findContributorManagedInsightRuleDescriptionByTemplateName(ctx context.Context, conn *cloudwatch.Client, resourceARN string, templateName string) (*awstypes.ManagedRuleDescription, error) {
	input := &cloudwatch.ListManagedInsightRulesInput{
		ResourceARN: aws.String(resourceARN),
	}

	rule, err := findContributorManagedInsightRule(ctx, conn, input, func(v *awstypes.ManagedRuleDescription) bool {
		return aws.ToString(v.TemplateName) == templateName
	})

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return rule, nil
}
