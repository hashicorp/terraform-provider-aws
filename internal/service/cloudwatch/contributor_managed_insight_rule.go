// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			"state": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("ENABLED"),
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

	in := &cloudwatch.PutManagedInsightRulesInput{
		ManagedRules: []awstypes.ManagedRule{
			{
				ResourceARN:  plan.ResourceArn.ValueStringPointer(),
				TemplateName: plan.TemplateName.ValueStringPointer(),
			},
		},
	}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.ManagedRules[0].Tags = getTagsIn(ctx)

	if plan.State.ValueString() == "ENABLED" || plan.State.IsNull() {
		out, err := conn.PutManagedInsightRules(ctx, in)
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

		// Get the rule name from the list API
		rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, plan.ResourceArn.ValueString(), plan.TemplateName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudWatch, create.ErrActionReading, ResNameContributorManagedInsightRule, plan.ResourceArn.String(), err),
				err.Error(),
			)
			return
		}

		plan.RuleName = types.StringValue(*rule.RuleState.RuleName)

		cmirARN := r.Meta().RegionalARN(ctx, "cloudwatch", fmt.Sprintf("insight-rule/%s", plan.RuleName.ValueString()))
		plan.ARN = fwflex.StringValueToFramework(ctx, cmirARN)

	} else if plan.State.ValueString() == "DISABLED" {
		rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, plan.ResourceArn.ValueString(), plan.TemplateName.ValueString())
		_, err = conn.DisableInsightRules(ctx, &cloudwatch.DisableInsightRulesInput{
			RuleNames: []string{*rule.RuleState.RuleName},
		})

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

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.RuleName = fwflex.StringValueToFramework(ctx, *out.RuleState.RuleName)

	cmirARN := r.Meta().RegionalARN(ctx, "cloudwatch", fmt.Sprintf("insight-rule/%s", state.RuleName.ValueString()))
	// cmirARN := r.Meta().RegionalARN(ctx, "cloudwatch", fmt.Sprintf("insight-rule/%s", *out.RuleState.RuleName))
	state.ARN = fwflex.StringValueToFramework(ctx, cmirARN)

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
		if new.State.ValueString() == "ENABLED" || new.State.IsNull() {
			_, err := conn.PutManagedInsightRules(ctx, &cloudwatch.PutManagedInsightRulesInput{
				ManagedRules: []awstypes.ManagedRule{
					{
						ResourceARN:  aws.String(new.ResourceArn.ValueString()),
						TemplateName: aws.String(new.TemplateName.ValueString()),
					},
				},
			})
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.CloudWatch, create.ErrActionUpdating, ResNameContributorManagedInsightRule, new.ResourceArn.String(), err),
					err.Error(),
				)
			}
		} else if new.State.ValueString() == "DISABLED" {
			rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, new.ResourceArn.ValueString(), new.TemplateName.ValueString())
			_, err = conn.DisableInsightRules(ctx, &cloudwatch.DisableInsightRulesInput{
				RuleNames: []string{*rule.RuleState.RuleName},
			})
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

	fmt.Printf("Delete: Getting rule for ARN: %s, Template: %s\n", state.ResourceArn.ValueString(), state.TemplateName.ValueString())

	in := &cloudwatch.DeleteInsightRulesInput{
		RuleNames: []string{state.RuleName.ValueString()},
	}

	// rule, err := findContributorManagedInsightRuleDescriptionByTemplateName(ctx, conn, state.ResourceArn.ValueString(), state.TemplateName.ValueString())
	// if err != nil {
	// 	fmt.Printf("Delete: Error finding rule: %v\n", err)
	// 	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
	// 		return
	// 	}
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.CloudWatch, create.ErrActionDeleting, ResNameContributorManagedInsightRule, state.ResourceArn.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// fmt.Printf("Delete: Found rule with name: %s\n", *rule.RuleState.RuleName)

	_, err := conn.DeleteInsightRules(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionDeleting, ResNameContributorManagedInsightRule, state.ResourceArn.String(), err),
			err.Error(),
		)
		return
	}
	fmt.Printf("Delete: Successfully deleted rule\n")
}

func (r *resourceContributorManagedInsightRule) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ";")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Resource Import Invalid ID",
			fmt.Sprintf("Wrong format for import ID (%s), use: 'resource-arn;template-name'", req.ID),
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

func (r *resourceContributorManagedInsightRule) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceContributorManagedInsightRuleData struct {
	ARN          types.String `tfsdk:"arn"`
	ResourceArn  types.String `tfsdk:"resource_arn"`
	TemplateName types.String `tfsdk:"template_name"`
	State        types.String `tfsdk:"state"`
	RuleName     types.String `tfsdk:"rule_name"`
	Tags         tftags.Map   `tfsdk:"tags"`
	TagsAll      tftags.Map   `tfsdk:"tags_all"`
}

func findContributorManagedInsightRules(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.ListManagedInsightRulesInput, filter tfslices.Predicate[*awstypes.ManagedRule]) ([]awstypes.ManagedRuleDescription, error) {
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

		for _, v := range page.ManagedRules {
			managedRule := awstypes.ManagedRule{
				ResourceARN:  v.ResourceARN,
				TemplateName: v.TemplateName,
			}
			if filter(&managedRule) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findContributorManagedInsightRule(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.ListManagedInsightRulesInput, filter tfslices.Predicate[*awstypes.ManagedRule]) (*awstypes.ManagedRuleDescription, error) {
	output, err := findContributorManagedInsightRules(ctx, conn, input, filter)
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findContributorManagedInsightRuleByTwoPartKey(ctx context.Context, conn *cloudwatch.Client, resourceARN string, templateName string) (*awstypes.ManagedRule, error) {
	input := &cloudwatch.ListManagedInsightRulesInput{
		ResourceARN: aws.String(resourceARN),
	}

	output, err := findContributorManagedInsightRule(ctx, conn, input, func(v *awstypes.ManagedRule) bool {
		return aws.ToString(v.TemplateName) == templateName
	})
	if err != nil {
		return nil, err
	}
	return &awstypes.ManagedRule{
		ResourceARN:  output.ResourceARN,
		TemplateName: output.TemplateName,
	}, nil
}

func findContributorManagedInsightRuleDescriptionByTemplateName(ctx context.Context, conn *cloudwatch.Client, resourceARN string, templateName string) (*awstypes.ManagedRuleDescription, error) {
	input := &cloudwatch.ListManagedInsightRulesInput{
		ResourceARN: aws.String(resourceARN),
	}

	rules, err := findContributorManagedInsightRules(ctx, conn, input, func(v *awstypes.ManagedRule) bool {
		return aws.ToString(v.TemplateName) == templateName
	})
	if err != nil {
		return nil, err
	}

	if len(rules) == 0 {
		return nil, &retry.NotFoundError{
			LastError:   fmt.Errorf("no matching rule found"),
			LastRequest: input,
		}
	}

	return &rules[0], nil
}
