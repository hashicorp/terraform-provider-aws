// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_waf_subscribed_rule_group", name="Subscribed Rule Group")
func dataSourceSubscribedRuleGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSubscribedRuleGroupRead,

		Schema: map[string]*schema.Schema{
			"metric_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{names.AttrName, "metric_name"},
			},
		},
	}
}

func dataSourceSubscribedRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	output, err := findSubscribedRuleGroupByNameOrMetricName(ctx, conn, d.Get(names.AttrName).(string), d.Get("metric_name").(string))

	if err != nil {
		return diag.Errorf("reading WAF Subscribed Rule Group: %s", err)
	}

	d.SetId(aws.ToString(output.RuleGroupId))
	d.Set("metric_name", output.MetricName)
	d.Set(names.AttrName, output.Name)

	return nil
}

func findSubscribedRuleGroupByNameOrMetricName(ctx context.Context, conn *waf.Client, name, metricName string) (*awstypes.SubscribedRuleGroupSummary, error) {
	hasName := name != ""
	hasMetricName := metricName != ""
	hasMatch := false

	input := &waf.ListSubscribedRuleGroupsInput{}
	var matchingRuleGroup *awstypes.SubscribedRuleGroupSummary
	err := listSubscribedRuleGroupsPages(ctx, conn, input, func(page *waf.ListSubscribedRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RuleGroups {
			respName := aws.ToString(v.Name)
			respMetricName := aws.ToString(v.MetricName)

			if hasName && respName != name {
				continue
			}
			if hasMetricName && respMetricName != metricName {
				continue
			}
			if hasName && hasMetricName && (name != respName || metricName != respMetricName) {
				continue
			}

			matchingRuleGroup = &v
			hasMatch = true
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if !hasMatch {
		return nil, fmt.Errorf("no matches found for name %s and metricName %s", name, metricName)
	}

	return matchingRuleGroup, nil
}
