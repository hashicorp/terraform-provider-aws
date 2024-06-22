// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_waf_rate_based_rule", name="Rate Based Rule")
func dataSourceRateBasedRule() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRateBasedRuleRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceRateBasedRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &waf.ListRateBasedRulesInput{}
	output, err := findRateBasedRule(ctx, conn, input, func(v *awstypes.RuleSummary) bool {
		return aws.ToString(v.Name) == name
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("WAF Rate Based Rule", err))
	}

	d.SetId(aws.ToString(output.RuleId))

	return diags
}

func findRateBasedRule(ctx context.Context, conn *waf.Client, input *waf.ListRateBasedRulesInput, filter tfslices.Predicate[*awstypes.RuleSummary]) (*awstypes.RuleSummary, error) {
	output, err := findRateBasedRules(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRateBasedRules(ctx context.Context, conn *waf.Client, input *waf.ListRateBasedRulesInput, filter tfslices.Predicate[*awstypes.RuleSummary]) ([]awstypes.RuleSummary, error) {
	var output []awstypes.RuleSummary

	err := listRateBasedRulesPages(ctx, conn, input, func(page *waf.ListRateBasedRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Rules {
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
