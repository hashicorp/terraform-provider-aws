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
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_lb_listener_rule")
func newListenerRuleResourceAsListResource() inttypes.ListResourceForSDK {
	l := listenerRuleListResource{}
	l.SetResourceSchema(resourceListenerRule())
	return &l
}

var _ list.ListResource = &listenerRuleListResource{}

type listenerRuleListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listenerRuleListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"listener_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the Listener to list ListenerRules from.",
			},
		},
	}
}

func (l *listenerRuleListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ELBV2Client(ctx)

	var query listListenerRuleModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	listenerARN := query.ListenerARN.ValueString()

	logParams := map[string]any{
		"tf_list.request.include_resource":    request.IncludeResource,
		"tf_list.request.limit":               request.Limit,
		"tf_list.request.config.listener_arn": listenerARN,
	}
	if !query.Region.IsNull() {
		logParams["tf_list.request.config.region"] = query.Region.ValueString()
	}

	tflog.Info(ctx, "Listing Resources", logParams)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := elasticloadbalancingv2.DescribeRulesInput{
			ListenerArn: aws.String(listenerARN),
		}
		for item, err := range listListenerRules(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}
			arn := aws.ToString(item.rule.RuleArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				setTagsOut(ctx, item.tags)
				rd.Set("listener_arn", listenerARN)
				if err := resourceListenerRuleFlatten(ctx, &item.rule, rd); err != nil {
					tflog.Error(ctx, "Reading ELB Listener Rule", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			tags := keyValueTags(ctx, item.tags)
			if v, ok := tags["Name"]; ok {
				result.DisplayName = v.ValueString()
			} else {
				result.DisplayName = aws.ToString(item.rule.RuleArn)
			}

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
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
