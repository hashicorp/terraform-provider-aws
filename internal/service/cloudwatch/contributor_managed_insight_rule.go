// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudwatch

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
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
// @IdentityAttribute("resource_arn")
// @IdentityAttribute("template_name")
// @ImportIDHandler("contributorManagedInsightRuleImportID")
// @Testing(importStateIdFunc=testAccContributorManagedInsightRuleImportStateIDFunc)
// @Testing(preIdentityVersion="v6.52.0")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatch/types;awstypes;awstypes.ManagedRuleDescription")
// @Testing(generator="randomLoadBalancerName(t)")
// @Testing(importStateIdAttribute="resource_arn")
func newContributorManagedInsightRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &contributorManagedInsightRuleResource{}

	return r, nil
}

type contributorManagedInsightRuleResource struct {
	framework.ResourceWithModel[contributorManagedInsightRuleResourceModel]
	framework.WithImportByIdentity
}

func (r *contributorManagedInsightRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	stateType := fwtypes.StringEnumType[stateValue]()
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				CustomType: stateType,
				Optional:   true,
				Computed:   true,
				Default:    stateType.AttributeDefault(stateValueEnabled),
			},
			names.AttrTags:    tftags.TagsAttributeForceNew(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"template_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`[0-9A-Za-z][\-\.\_0-9A-Za-z]{0,126}[0-9A-Za-z]`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *contributorManagedInsightRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	c := r.Meta()
	conn := c.CloudWatchClient(ctx)

	var plan contributorManagedInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	resourceARN, templateName := fwflex.StringValueFromFramework(ctx, plan.ResourceARN), fwflex.StringValueFromFramework(ctx, plan.TemplateName)
	input := cloudwatch.PutManagedInsightRulesInput{
		ManagedRules: []awstypes.ManagedRule{
			{
				ResourceARN:  aws.String(resourceARN),
				Tags:         getTagsIn(ctx),
				TemplateName: aws.String(templateName),
			},
		},
	}
	output, err := conn.PutManagedInsightRules(ctx, &input)
	if err == nil {
		err = partialFailuresError(output.Failures)
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
		return
	}

	// Taint the resource.
	resp.State.SetAttribute(ctx, path.Root(names.AttrResourceARN), resourceARN)
	resp.State.SetAttribute(ctx, path.Root("template_name"), templateName)

	rule, err := findManagedRuleByTwoPartKey(ctx, conn, resourceARN, templateName)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
		return
	}
	ruleName := aws.ToString(rule.RuleState.RuleName)

	switch state := plan.State.ValueEnum(); state {
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

	plan.ARN = fwflex.StringValueToFramework(ctx, insightRuleARN(ctx, c, ruleName))
	plan.RuleName = fwflex.StringValueToFramework(ctx, ruleName)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *contributorManagedInsightRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	c := r.Meta()
	conn := c.CloudWatchClient(ctx)

	var state contributorManagedInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resourceARN, templateName := fwflex.StringValueFromFramework(ctx, state.ResourceARN), fwflex.StringValueFromFramework(ctx, state.TemplateName)
	rule, err := findManagedRuleByTwoPartKey(ctx, conn, resourceARN, templateName)
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
		return
	}

	// Set attributes for import.
	ruleName := aws.ToString(rule.RuleState.RuleName)
	state.ARN = fwflex.StringValueToFramework(ctx, insightRuleARN(ctx, c, ruleName))
	state.RuleName = fwflex.StringValueToFramework(ctx, ruleName)
	state.State = fwtypes.StringEnumValue(stateValue(aws.ToString(rule.RuleState.State)))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
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

	ruleName := fwflex.StringValueFromFramework(ctx, new.RuleName)
	switch state := new.State.ValueEnum(); state {
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

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &new))
}

func (r *contributorManagedInsightRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state contributorManagedInsightRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	ruleName := fwflex.StringValueFromFramework(ctx, state.RuleName)
	input := cloudwatch.DeleteInsightRulesInput{
		RuleNames: []string{ruleName},
	}
	output, err := conn.DeleteInsightRules(ctx, &input)
	if err == nil {
		err = partialFailuresError(output.Failures)
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ruleName)
		return
	}
}

const contributorManagedInsightRuleImportIDSeparator = intflex.ResourceIdSeparator

func contributorManagedInsightRuleParseImportID(id string) (string, string, error) {
	parts := strings.Split(id, contributorManagedInsightRuleImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected resource-arn%[2]stemplate-name", id, contributorManagedInsightRuleImportIDSeparator)
}

var (
	_ inttypes.ImportIDParser = contributorManagedInsightRuleImportID{}
)

type contributorManagedInsightRuleImportID struct{}

func (contributorManagedInsightRuleImportID) Parse(identifier string) (string, map[string]any, error) {
	resourceARN, templateName, err := contributorManagedInsightRuleParseImportID(identifier)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		names.AttrResourceARN: resourceARN,
		"template_name":       templateName,
	}

	return identifier, result, nil
}

type contributorManagedInsightRuleResourceModel struct {
	framework.WithRegionModel
	ARN          types.String                   `tfsdk:"arn"`
	ResourceARN  fwtypes.ARN                    `tfsdk:"resource_arn"`
	RuleName     types.String                   `tfsdk:"rule_name"`
	State        fwtypes.StringEnum[stateValue] `tfsdk:"state"`
	Tags         tftags.Map                     `tfsdk:"tags"`
	TagsAll      tftags.Map                     `tfsdk:"tags_all"`
	TemplateName types.String                   `tfsdk:"template_name"`
}

func findManagedRules(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.ListManagedInsightRulesInput, filter tfslices.Predicate[awstypes.ManagedRuleDescription]) ([]awstypes.ManagedRuleDescription, error) {
	var output []awstypes.ManagedRuleDescription

	pages := cloudwatch.NewListManagedInsightRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ManagedRules {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findManagedRule(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.ListManagedInsightRulesInput, filter tfslices.Predicate[awstypes.ManagedRuleDescription]) (*awstypes.ManagedRuleDescription, error) {
	output, err := findManagedRules(ctx, conn, input, filter)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findManagedRuleByTwoPartKey(ctx context.Context, conn *cloudwatch.Client, resourceARN, templateName string) (*awstypes.ManagedRuleDescription, error) {
	input := cloudwatch.ListManagedInsightRulesInput{
		ResourceARN: aws.String(resourceARN),
	}

	output, err := findManagedRule(ctx, conn, &input, func(v awstypes.ManagedRuleDescription) bool {
		return aws.ToString(v.TemplateName) == templateName
	})
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output.RuleState == nil || aws.ToString(output.RuleState.RuleName) == "" {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, smarterr.NewError(err)
}
