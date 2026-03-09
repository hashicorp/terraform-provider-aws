// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_cloudwatch_event_rule")
func newRuleResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceRule{}
	l.SetResourceSchema(resourceRule())
	return &l
}

var _ list.ListResource = &listResourceRule{}

type listResourceRule struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceRule) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EventsClient(ctx)

	var query listRuleModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing EventBridge Rule")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input eventbridge.ListRulesInput
		for item, err := range listRules(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.Name)
			eventBusName := aws.ToString(item.EventBusName)
			id := ruleCreateResourceID(eventBusName, name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)

			tflog.Info(ctx, "Reading EventBridge Rule")
			output, err := findRuleByTwoPartKey(ctx, conn, eventBusName, name)
			if err != nil {
				tflog.Error(ctx, "Reading EventBridge Rule", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			diags := resourceRuleFlatten(ctx, eventBusName, rd, output)
			if diags.HasError() {
				result = fwdiag.NewListResultSDKDiagnostics(diags)
				yield(result)
				return
			}

			result.DisplayName = name

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

type listRuleModel struct {
	framework.WithRegionModel
}

func listRules(ctx context.Context, conn *eventbridge.Client, input *eventbridge.ListRulesInput) iter.Seq2[awstypes.Rule, error] {
	return func(yield func(awstypes.Rule, error) bool) {
		err := listRulesPages(ctx, conn, input, func(page *eventbridge.ListRulesOutput, lastPage bool) bool {
			for _, item := range page.Rules {
				if !yield(item, nil) {
					return !lastPage
				}
			}
			return !lastPage
		})
		if err != nil {
			yield(awstypes.Rule{}, fmt.Errorf("listing EventBridge Rule resources: %w", err))
			return
		}
	}
}
