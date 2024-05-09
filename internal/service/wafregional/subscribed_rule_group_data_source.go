// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	DSNameSubscribedRuleGroup = "Subscribed Rule Group Data Source"
)

// @SDKDataSource("aws_wafregional_subscribed_rule_group", name="Subscribed Rule Group")
func dataSourceSubscribedRuleGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSubscribedRuleGroupRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metric_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceSubscribedRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	name, nameOk := d.Get(names.AttrName).(string)
	metricName, metricNameOk := d.Get("metric_name").(string)

	// Error out if string-assertion fails for either name or metricName
	if !nameOk || !metricNameOk {
		if !nameOk {
			name = DSNameSubscribedRuleGroup
		}

		err := errors.New("unable to read attributes")
		return create.DiagError(names.WAFRegional, create.ErrActionReading, DSNameSubscribedRuleGroup, name, err)
	}

	output, err := findSubscribedRuleGroupByNameOrMetricName(ctx, conn, name, metricName)

	if err != nil {
		return create.DiagError(names.WAFRegional, create.ErrActionReading, DSNameSubscribedRuleGroup, name, err)
	}

	d.SetId(aws.ToString(output.RuleGroupId))
	d.Set("metric_name", output.MetricName)
	d.Set(names.AttrName, output.Name)

	return nil
}

func findSubscribedRuleGroupByNameOrMetricName(ctx context.Context, conn *wafregional.Client, name string, metricName string) (*awstypes.SubscribedRuleGroupSummary, error) {
	hasName := name != ""
	hasMetricName := metricName != ""
	hasMatch := false

	if !hasName && !hasMetricName {
		return nil, errors.New("must specify either name or metricName")
	}

	input := &wafregional.ListSubscribedRuleGroupsInput{}

	matchingRuleGroup := &awstypes.SubscribedRuleGroupSummary{}

	for {
		output, err := conn.ListSubscribedRuleGroups(ctx, input)

		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, ruleGroup := range output.RuleGroups {
			respName := aws.ToString(ruleGroup.Name)
			respMetricName := aws.ToString(ruleGroup.MetricName)

			if hasName && respName != name {
				continue
			}
			if hasMetricName && respMetricName != metricName {
				continue
			}
			if hasName && hasMetricName && (name != respName || metricName != respMetricName) {
				continue
			}
			// Previous conditionals catch all non-matches
			if hasMatch {
				return nil, fmt.Errorf("multiple matches found for name %s and metricName %s", name, metricName)
			}

			copyObject := ruleGroup
			matchingRuleGroup = &copyObject
			hasMatch = true
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if !hasMatch {
		return nil, fmt.Errorf("no matches found for name %s and metricName %s", name, metricName)
	}

	return matchingRuleGroup, nil
}
