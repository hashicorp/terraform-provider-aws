// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_wafregional_subscribed_rule_group", name="Subscribed Rule Group")
func dataSourceSubscribedRuleGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSubscribedRuleGroupRead,

		Schema: map[string]*schema.Schema{
			names.AttrMetricName: {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{names.AttrName, names.AttrMetricName},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{names.AttrName, names.AttrMetricName},
			},
		},
	}
}

func dataSourceSubscribedRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	var filter tfslices.Predicate[*awstypes.SubscribedRuleGroupSummary]

	if v, ok := d.GetOk(names.AttrMetricName); ok {
		name := v.(string)
		filter = func(v *awstypes.SubscribedRuleGroupSummary) bool {
			return aws.ToString(v.MetricName) == name
		}
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		f := func(v *awstypes.SubscribedRuleGroupSummary) bool {
			return aws.ToString(v.Name) == name
		}

		if filter != nil {
			filter = tfslices.PredicateAnd(filter, f)
		} else {
			filter = f
		}
	}

	input := &wafregional.ListSubscribedRuleGroupsInput{}
	output, err := findSubscribedRuleGroup(ctx, conn, input, filter)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("WAF Regional Subscribed Rule Group", err))
	}

	d.SetId(aws.ToString(output.RuleGroupId))
	d.Set(names.AttrMetricName, output.MetricName)
	d.Set(names.AttrName, output.Name)

	return diags
}

func findSubscribedRuleGroup(ctx context.Context, conn *wafregional.Client, input *wafregional.ListSubscribedRuleGroupsInput, filter tfslices.Predicate[*awstypes.SubscribedRuleGroupSummary]) (*awstypes.SubscribedRuleGroupSummary, error) {
	output, err := findSubscribedRuleGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSubscribedRuleGroups(ctx context.Context, conn *wafregional.Client, input *wafregional.ListSubscribedRuleGroupsInput, filter tfslices.Predicate[*awstypes.SubscribedRuleGroupSummary]) ([]awstypes.SubscribedRuleGroupSummary, error) {
	var output []awstypes.SubscribedRuleGroupSummary

	err := listSubscribedRuleGroupsPages(ctx, conn, input, func(page *wafregional.ListSubscribedRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RuleGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
