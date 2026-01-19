// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_wafv2_managed_rule_group", name="Managed Rule Group")
func newManagedRuleGroupDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &managedRuleGroupDataSource{}, nil
}

type managedRuleGroupDataSource struct {
	framework.DataSourceWithModel[managedRuleGroupDataSourceModel]
}

func (d *managedRuleGroupDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"available_labels": framework.DataSourceComputedListOfObjectAttribute[labelSummaryModel](ctx),
			"capacity": schema.Int64Attribute{
				Computed: true,
			},
			"consumed_labels": framework.DataSourceComputedListOfObjectAttribute[labelSummaryModel](ctx),
			"label_namespace": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"rules": framework.DataSourceComputedListOfObjectAttribute[ruleSummaryModel](ctx),
			names.AttrScope: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Scope](),
				Required:   true,
			},
			names.AttrSNSTopicARN: schema.StringAttribute{
				Computed: true,
			},
			"vendor_name": schema.StringAttribute{
				Required: true,
			},
			"version_name": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *managedRuleGroupDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data managedRuleGroupDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().WAFV2Client(ctx)

	var input wafv2.DescribeManagedRuleGroupInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findManagedRuleGroup(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("reading WAFv2 Managed Rule Group", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findManagedRuleGroup(ctx context.Context, conn *wafv2.Client, input *wafv2.DescribeManagedRuleGroupInput) (*wafv2.DescribeManagedRuleGroupOutput, error) {
	output, err := conn.DescribeManagedRuleGroup(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type managedRuleGroupDataSourceModel struct {
	framework.WithRegionModel
	AvailableLabels fwtypes.ListNestedObjectValueOf[labelSummaryModel] `tfsdk:"available_labels"`
	Capacity        types.Int64                                        `tfsdk:"capacity"`
	ConsumedLabels  fwtypes.ListNestedObjectValueOf[labelSummaryModel] `tfsdk:"consumed_labels"`
	LabelNamespace  types.String                                       `tfsdk:"label_namespace"`
	Name            types.String                                       `tfsdk:"name"`
	Rules           fwtypes.ListNestedObjectValueOf[ruleSummaryModel]  `tfsdk:"rules"`
	Scope           fwtypes.StringEnum[awstypes.Scope]                 `tfsdk:"scope"`
	SNSTopicARN     types.String                                       `tfsdk:"sns_topic_arn"`
	VendorName      types.String                                       `tfsdk:"vendor_name"`
	VersionName     types.String                                       `tfsdk:"version_name"`
}

type labelSummaryModel struct {
	Name types.String `tfsdk:"name"`
}

type ruleSummaryModel struct {
	Action fwtypes.ListNestedObjectValueOf[ruleActionModel] `tfsdk:"action"`
	Name   types.String                                     `tfsdk:"name"`
}
