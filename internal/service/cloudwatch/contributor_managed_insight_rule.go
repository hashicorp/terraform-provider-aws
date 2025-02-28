// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
func newResourceContributorManagedInsightRule(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceContributorManagedInsightRule{}

	return r, nil
}

const (
	ResNameContributorManagedInsightRule = "Contributor Managed Insight Rule"
)

type resourceContributorManagedInsightRule struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[resourceContributorManagedInsightRuleData]
}

func (r *resourceContributorManagedInsightRule) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_cloudwatch_contributor_managed_insight_rule"
}

func (r *resourceContributorManagedInsightRule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *resourceContributorManagedInsightRule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var plan resourceContributorManagedInsightRuleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.ManagedRules[0].Tags = getTagsIn(ctx)

	if plan.State.ValueEnum() == stateValueEnabled || plan.State.IsNull() {
		out, err := conn.PutManagedInsightRules(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudWatch, create.ErrActionCreating, ResNameContributorManagedInsightRule, plan.ResourceArn.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudWatch, create.ErrActionCreating, ResNameContributorManagedInsightRule, plan.ResourceArn.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, plan.ResourceArn.ValueString(), plan.TemplateName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudWatch, create.ErrActionReading, ResNameContributorManagedInsightRule, plan.ResourceArn.String(), err),
				err.Error(),
			)
			return
		}

		plan.RuleName = fwflex.StringToFramework(ctx, rule.RuleState.RuleName)

		cmirARN := r.Meta().RegionalARN(ctx, "cloudwatch", fmt.Sprintf("insight-rule/%s", plan.RuleName.ValueString()))
		plan.ARN = fwflex.StringValueToFramework(ctx, cmirARN)
	} else if plan.State.ValueEnum() == stateValueDisabled {
		rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, plan.ResourceArn.ValueString(), plan.TemplateName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudWatch, create.ErrActionCreating, ResNameContributorManagedInsightRule, plan.ResourceArn.String(), err),
				err.Error(),
			)
			return
		}
		input := cloudwatch.DisableInsightRulesInput{
			RuleNames: []string{*rule.RuleState.RuleName},
		}
		_, err = conn.DisableInsightRules(ctx, &input)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudWatch, create.ErrActionCreating, ResNameContributorManagedInsightRule, plan.ResourceArn.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceContributorManagedInsightRule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state resourceContributorManagedInsightRuleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, state.ResourceArn.ValueString(), state.TemplateName.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionSetting, ResNameContributorManagedInsightRule, state.ResourceArn.String(), err),
			err.Error(),
		)
		return
	}

	if out.RuleState != nil && out.RuleState.RuleName != nil {
		state.RuleName = fwflex.StringToFramework(ctx, out.RuleState.RuleName)
		cmirARN := r.Meta().RegionalARN(ctx, "cloudwatch", fmt.Sprintf("insight-rule/%s", state.RuleName.ValueString()))
		state.ARN = fwflex.StringValueToFramework(ctx, cmirARN)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceContributorManagedInsightRule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceContributorManagedInsightRuleData

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
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
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.CloudWatch, create.ErrActionUpdating, ResNameContributorManagedInsightRule, new.ResourceArn.String(), err),
					err.Error(),
				)
			}
		} else if new.State.ValueEnum() == stateValueDisabled {
			rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, new.ResourceArn.ValueString(), new.TemplateName.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.CloudWatch, create.ErrActionUpdating, ResNameContributorManagedInsightRule, new.ResourceArn.String(), err),
					err.Error(),
				)
			}
			input := cloudwatch.DisableInsightRulesInput{
				RuleNames: []string{aws.ToString(rule.RuleState.RuleName)},
			}
			_, err = conn.DisableInsightRules(ctx, &input)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.CloudWatch, create.ErrActionUpdating, ResNameContributorManagedInsightRule, new.ResourceArn.String(), err),
					err.Error(),
				)
			}
		}
	}

	new.RuleName = old.RuleName

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceContributorManagedInsightRule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state resourceContributorManagedInsightRuleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, state.ResourceArn.ValueString(), state.TemplateName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionDeleting, ResNameContributorManagedInsightRule, state.ResourceArn.String(), err),
			err.Error(),
		)
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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionDeleting, ResNameContributorManagedInsightRule, state.ResourceArn.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceContributorManagedInsightRule) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	const resourceIDParts = 2
	idParts, err := flex.ExpandResourceId(req.ID, resourceIDParts, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Resource Import Invalid ID",
			fmt.Sprintf("Wrong format for import ID (%s), use: 'resource-arn,template-name'", req.ID),
		)
		return
	}
	resourceARN := idParts[0]
	templateName := idParts[1]

	cmir, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, r.Meta().CloudWatchClient(ctx), resourceARN, templateName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing Resource",
			err.Error(),
		)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrResourceARN), aws.ToString(cmir.ResourceARN))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("template_name"), templateName)...)
}

type resourceContributorManagedInsightRuleData struct {
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
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
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
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findContributorManagedInsightRuleDescriptionByTemplateName(ctx context.Context, conn *cloudwatch.Client, resourceARN string, templateName string) (*awstypes.ManagedRuleDescription, error) {
	input := &cloudwatch.ListManagedInsightRulesInput{
		ResourceARN: aws.String(resourceARN),
	}

	rule, err := findContributorManagedInsightRule(ctx, conn, input, func(v *awstypes.ManagedRuleDescription) bool {
		return aws.ToString(v.TemplateName) == templateName
	})

	if err != nil {
		return nil, err
	}

	return rule, nil
}
