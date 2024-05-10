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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_waf_rule", name="Rule")
func dataSourceRule() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRuleRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)
	name := d.Get(names.AttrName).(string)

	rules := make([]awstypes.RuleSummary, 0)
	// ListRulesInput does not have a name parameter for filtering
	input := &waf.ListRulesInput{}
	for {
		output, err := conn.ListRules(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAF Rules: %s", err)
		}
		for _, rule := range output.Rules {
			if aws.ToString(rule.Name) == name {
				rules = append(rules, rule)
			}
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if len(rules) == 0 {
		return sdkdiag.AppendErrorf(diags, "WAF Rules not found for name: %s", name)
	}

	if len(rules) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple WAF Rules found for name: %s", name)
	}

	rule := rules[0]

	d.SetId(aws.ToString(rule.RuleId))

	return diags
}
