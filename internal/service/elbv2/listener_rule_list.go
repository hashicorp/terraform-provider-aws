// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_lb_listener_rule")
func newListenerRuleResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceListenerRule(), framework.SDKListSpec[listenerRuleResult, listListenerRuleModel]{
		ConfigSchema: listListenerRuleConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listListenerRuleModel) (iter.Seq2[listenerRuleResult, error], diag.Diagnostics) {
			input := elasticloadbalancingv2.DescribeRulesInput{
				ListenerArn: query.ListenerARN.ValueStringPointer(),
			}

			return listListenerRules(ctx, c.ELBV2Client(ctx), &input), nil
		},
		ID: func(r listenerRuleResult, _ listListenerRuleModel) string {
			return aws.ToString(r.rule.RuleArn)
		},
		Prepare: func(ctx context.Context, r listenerRuleResult, query listListenerRuleModel, rd *schema.ResourceData) {
			// Both are part of this resource's identity. The tags came back
			// with the page, so staging them costs nothing.
			rd.Set(names.AttrARN, aws.ToString(r.rule.RuleArn))
			rd.Set("listener_arn", query.ListenerARN.ValueString())
			setTagsOut(ctx, r.tags)
		},
		Flatten: func(ctx context.Context, c *conns.AWSClient, r *listenerRuleResult, rd *schema.ResourceData) error {
			return resourceListenerRuleFlatten(ctx, c, &r.rule, rd)
		},
		DisplayName: func(r listenerRuleResult) string {
			// Rule ARNs are long and nearly identical, so prefer the Name tag.
			for _, tag := range r.tags {
				if aws.ToString(tag.Key) == "Name" {
					return aws.ToString(tag.Value)
				}
			}

			return aws.ToString(r.rule.RuleArn)
		},
	})
}

type listListenerRuleModel struct {
	framework.WithRegionModel
	ListenerARN fwtypes.ARN `tfsdk:"listener_arn"`
}

type listenerRuleResult struct {
	rule awstypes.Rule
	tags []awstypes.Tag
}

func listListenerRules(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeRulesInput) iter.Seq2[listenerRuleResult, error] {
	return func(yield func(listenerRuleResult, error) bool) {
		pages := elasticloadbalancingv2.NewDescribeRulesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(listenerRuleResult{}, fmt.Errorf("listing ELB Listener Rule resources: %w", err))
				return
			}

			ruleARNs := make([]string, len(page.Rules))
			for i, rule := range page.Rules {
				ruleARNs[i] = aws.ToString(rule.RuleArn)
			}

			tags, err := batchListTags(ctx, conn, ruleARNs)
			if err != nil {
				yield(listenerRuleResult{}, fmt.Errorf("listing ELB Listener Rule resource tags: %w", err))
				return
			}

			for _, item := range page.Rules {
				if aws.ToBool(item.IsDefault) {
					continue
				}
				v := listenerRuleResult{
					rule: item,
					tags: tags[aws.ToString(item.RuleArn)],
				}
				if !yield(v, nil) {
					return
				}
			}
		}
	}
}

func listListenerRuleConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"listener_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the Listener to list ListenerRules from.",
			},
		},
	}
}
