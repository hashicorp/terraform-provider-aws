// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func FindRegexMatchSetByID(ctx context.Context, conn *wafregional.WAFRegional, id string) (*waf.RegexMatchSet, error) {
	result, err := conn.GetRegexMatchSetWithContext(ctx, &waf.GetRegexMatchSetInput{
		RegexMatchSetId: aws.String(id),
	})

	return result.RegexMatchSet, err
}

func FindSubscribedRuleGroupByNameOrMetricName(ctx context.Context, conn *wafregional.WAFRegional, name string, metricName string) (*waf.SubscribedRuleGroupSummary, error) {
	hasName := name != ""
	hasMetricName := metricName != ""
	hasMatch := false

	if !hasName && !hasMetricName {
		return nil, errors.New("must specify either name or metricName")
	}

	input := &waf.ListSubscribedRuleGroupsInput{}

	matchingRuleGroup := &waf.SubscribedRuleGroupSummary{}

	for {
		output, err := conn.ListSubscribedRuleGroupsWithContext(ctx, input)

		if tfawserr.ErrCodeContains(err, waf.ErrCodeNonexistentItemException) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, ruleGroup := range output.RuleGroups {
			respName := aws.StringValue(ruleGroup.Name)
			respMetricName := aws.StringValue(ruleGroup.MetricName)

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

			matchingRuleGroup = ruleGroup
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
