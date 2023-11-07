// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_wafv2_rule_group")
func DataSourceRuleGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRuleGroupRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"description": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"scope": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(wafv2.Scope_Values(), false),
				},
			}
		},
	}
}

func dataSourceRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)
	name := d.Get("name").(string)

	var foundRuleGroup *wafv2.RuleGroupSummary
	input := &wafv2.ListRuleGroupsInput{
		Scope: aws.String(d.Get("scope").(string)),
		Limit: aws.Int64(100),
	}

	for {
		resp, err := conn.ListRuleGroupsWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 RuleGroups: %s", err)
		}

		if resp == nil || resp.RuleGroups == nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 RuleGroups")
		}

		for _, ruleGroup := range resp.RuleGroups {
			if aws.StringValue(ruleGroup.Name) == name {
				foundRuleGroup = ruleGroup
				break
			}
		}

		if resp.NextMarker == nil || foundRuleGroup != nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundRuleGroup == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 RuleGroup not found for name: %s", name)
	}

	d.SetId(aws.StringValue(foundRuleGroup.Id))
	d.Set("arn", foundRuleGroup.ARN)
	d.Set("description", foundRuleGroup.Description)

	return diags
}
