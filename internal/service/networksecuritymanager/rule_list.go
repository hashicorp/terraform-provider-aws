// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_networksecuritymanager_rule")
func newRuleResourceAsListResource() list.ListResourceWithConfigure {
	return &ruleListResource{}
}

var _ list.ListResource = &ruleListResource{}

type ruleListResource struct {
	ruleResource
	framework.WithList
}

func (l *ruleListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().NetworkSecurityManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input networksecuritymanager.ListRulesInput
		for item, err := range listRules(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// ListRules reports the draft ARN; the resource is identified by the
			// base ARN, the same way Read and the sweeper treat it.
			arn := ruleBaseARN(aws.ToString(item.RuleArn))
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var output *networksecuritymanager.GetRuleOutput
			if request.IncludeResource {
				var err error
				output, err = findRuleByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, arn))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data ruleResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, flattenRule(ctx, output, &data), smerr.ID, arn)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.RuleName)
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}
			if !yield(result) {
				return
			}
		}
	}
}

func listRules(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListRulesInput, optFns ...func(*networksecuritymanager.Options)) iter.Seq2[awstypes.RuleSummary, error] {
	return tfiter.ConcatValuesWithError(listRulePages(ctx, conn, input, optFns...))
}

func listRulePages(ctx context.Context, conn *networksecuritymanager.Client, input *networksecuritymanager.ListRulesInput, optFns ...func(*networksecuritymanager.Options)) iter.Seq2[[]awstypes.RuleSummary, error] {
	return func(yield func([]awstypes.RuleSummary, error) bool) {
		pages := networksecuritymanager.NewListRulesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.Rules, nil) {
				return
			}
		}
	}
}
