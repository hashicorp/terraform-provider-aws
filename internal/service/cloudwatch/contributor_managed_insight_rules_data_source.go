// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cloudwatch_contributor_managed_insight_rules", name="Contributor Managed Insight Rules")
func newContributorManagedInsightRulesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &contributorManagedInsightRulesDataSource{}, nil
}

const (
	DSNameContributorManagedInsightRules = "Contributor Managed Insight Rules Data Source"
)

type contributorManagedInsightRulesDataSource struct {
	framework.DataSourceWithModel[contributorManagedInsightRulesDataSourceModel]
}

func (d *contributorManagedInsightRulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: schema.StringAttribute{
				Required: true,
			},
			"managed_rules": framework.DataSourceComputedListOfObjectAttribute[managedRuleDescription](ctx),
		},
	}
}

func (d *contributorManagedInsightRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().CloudWatchClient(ctx)

	var data contributorManagedInsightRulesDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}
	resourceARN := data.ResourceARN.ValueString()
	input := &cloudwatch.ListManagedInsightRulesInput{
		ResourceARN: aws.String(resourceARN),
	}

	filter := tfslices.PredicateTrue[*awstypes.ManagedRuleDescription]()

	output, err := findContributorManagedInsightRules(ctx, conn, input, filter)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ResourceARN.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(
		ctx,
		struct {
			ManagedRules []awstypes.ManagedRuleDescription
		}{
			ManagedRules: output,
		},
		&data), smerr.ID, data.ResourceARN.String())
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.ResourceARN.String())
}

type contributorManagedInsightRulesDataSourceModel struct {
	framework.WithRegionModel
	ResourceARN  types.String                                            `tfsdk:"resource_arn"`
	ManagedRules fwtypes.ListNestedObjectValueOf[managedRuleDescription] `tfsdk:"managed_rules"`
}

type managedRuleDescription struct {
	ResourceARN  types.String                               `tfsdk:"resource_arn"`
	RuleState    fwtypes.ListNestedObjectValueOf[ruleState] `tfsdk:"rule_state"`
	TemplateName types.String                               `tfsdk:"template_name"`
}

type ruleState struct {
	RuleName types.String `tfsdk:"rule_name"`
	State    types.String `tfsdk:"state"`
}
