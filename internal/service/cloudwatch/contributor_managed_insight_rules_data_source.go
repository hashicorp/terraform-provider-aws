// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cloudwatch_contributor_managed_insight_rules", name="Contributor Managed Insight Rules")
func newDataSourceContributorManagedInsightRules(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceContributorManagedInsightRules{}, nil
}

const (
	DSNameContributorManagedInsightRules = "Contributor Managed Insight Rules Data Source"
)

type dataSourceContributorManagedInsightRules struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceContributorManagedInsightRules) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_cloudwatch_contributor_managed_insight_rules"
}

func (d *dataSourceContributorManagedInsightRules) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: schema.StringAttribute{
				Required: true,
			},
			"managed_rules": framework.DataSourceComputedListOfObjectAttribute[managedRuleDescription](ctx),
		},
	}
}

func (d *dataSourceContributorManagedInsightRules) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().CloudWatchClient(ctx)

	var data dataSourceContributorManagedInsightRulesData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	paginator := cloudwatch.NewListManagedInsightRulesPaginator(conn, &cloudwatch.ListManagedInsightRulesInput{
		ResourceARN: data.ResourceARN.ValueStringPointer(),
	})

	var out cloudwatch.ListManagedInsightRulesOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionReading, ResNameContributorManagedInsightRule, data.ResourceARN.String(), err)
			return
		}

		if page != nil && len(page.ManagedRules) > 0 {
			out.ManagedRules = append(out.ManagedRules, page.ManagedRules...)
		}
		resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceContributorManagedInsightRulesData struct {
	ResourceARN  types.String                                            `tfsdk:"resource_arn"`
	ManagedRules fwtypes.ListNestedObjectValueOf[managedRuleDescription] `tfsdk:"managed_rules"`
}

type managedRuleDescription struct {
	ResourceARN  types.String                     `tfsdk:"resource_arn"`
	RuleState    fwtypes.ObjectValueOf[ruleState] `tfsdk:"rule_state"`
	TemplateName types.String                     `tfsdk:"template_name"`
}

type ruleState struct {
	RuleName types.String `tfsdk:"rule_name"`
	State    types.String `tfsdk:"state"`
}
