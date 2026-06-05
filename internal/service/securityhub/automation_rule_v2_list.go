// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_securityhub_automation_rule_v2")
func newAutomationRuleV2ResourceAsListResource() list.ListResourceWithConfigure {
	return &automationRuleV2ListResource{}
}

var _ list.ListResource = &automationRuleV2ListResource{}

type automationRuleV2ListResource struct {
	automationRuleV2Resource
	framework.WithList
}

func (l *automationRuleV2ListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SecurityHubClient(ctx)

	var query listAutomationRuleV2Model
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input securityhub.ListAutomationRulesV2Input
		for item, err := range listAutomationRuleV2s(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.RuleArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			output, err := findAutomationRuleV2ByARN(ctx, conn, arn)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data automationRuleV2ResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.RuleName)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listAutomationRuleV2Model struct {
	framework.WithRegionModel
}

func listAutomationRuleV2s(ctx context.Context, conn *securityhub.Client, input *securityhub.ListAutomationRulesV2Input) iter.Seq2[awstypes.AutomationRulesMetadataV2, error] {
	return func(yield func(awstypes.AutomationRulesMetadataV2, error) bool) {
		var stopped bool
		err := listAutomationRulesV2Pages(ctx, conn, input, func(page *securityhub.ListAutomationRulesV2Output, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, item := range page.Rules {
				if !yield(item, nil) {
					stopped = true
					return false
				}
			}

			return !lastPage
		})

		if !stopped && err != nil {
			yield(inttypes.Zero[awstypes.AutomationRulesMetadataV2](), fmt.Errorf("listing Security Hub V2 Automation Rules: %w", err))
			return
		}
	}
}
