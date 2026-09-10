// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package mailmanager

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mailmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_mailmanager_rule_set")
func newRuleSetResourceAsListResource() list.ListResourceWithConfigure {
	return &ruleSetListResource{}
}

var _ list.ListResource = &ruleSetListResource{}

type ruleSetListResource struct {
	ruleSetResource
	framework.WithList
}

func (l *ruleSetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().MailManagerClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input mailmanager.ListRuleSetsInput
		for item, err := range listRuleSets(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(item.RuleSetId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			var out *mailmanager.GetRuleSetOutput
			if request.IncludeResource {
				out, err = findRuleSetByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data ruleSetResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = types.StringValue(id)
				data.Name = types.StringPointerValue(item.RuleSetName)

				if request.IncludeResource {
					result.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("RuleSet"))...)
					if result.Diagnostics.HasError() {
						return
					}
				}
				result.DisplayName = aws.ToString(item.RuleSetName)
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

func listRuleSets(ctx context.Context, conn *mailmanager.Client, input *mailmanager.ListRuleSetsInput) iter.Seq2[awstypes.RuleSet, error] {
	return func(yield func(awstypes.RuleSet, error) bool) {
		pages := mailmanager.NewListRuleSetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.RuleSet{}, fmt.Errorf("listing SES Mail Manager Rule Set resources: %w", err))
				return
			}
			for _, item := range page.RuleSets {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
