// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_wafv2_rule_group", name="Rule Group")
func dataSourceRuleGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRuleGroupRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrScope: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.Scope](),
				},
			}
		},
	}
}

func dataSourceRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)
	name := d.Get(names.AttrName).(string)

	var foundRuleGroup awstypes.RuleGroupSummary
	input := &wafv2.ListRuleGroupsInput{
		Scope: awstypes.Scope(d.Get(names.AttrScope).(string)),
		Limit: aws.Int32(100),
	}

	for {
		resp, err := conn.ListRuleGroups(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 RuleGroups: %s", err)
		}

		if resp == nil || resp.RuleGroups == nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 RuleGroups")
		}

		for _, ruleGroup := range resp.RuleGroups {
			if aws.ToString(ruleGroup.Name) == name {
				foundRuleGroup = ruleGroup
				break
			}
		}

		if resp.NextMarker == nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundRuleGroup.Id == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 RuleGroup not found for name: %s", name)
	}

	d.SetId(aws.ToString(foundRuleGroup.Id))
	d.Set(names.AttrARN, foundRuleGroup.ARN)
	d.Set(names.AttrDescription, foundRuleGroup.Description)

	return diags
}
