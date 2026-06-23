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

// @SDKListResource("aws_securityhub_insight")
func newInsightResourceAsListResource() inttypes.ListResourceForSDK {
	l := insightListResource{}
	l.SetResourceSchema(resourceInsight())
	return &l
}

var _ list.ListResource = &insightListResource{}

type insightListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *insightListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SecurityHubClient(ctx)

	var query listInsightModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input securityhub.GetInsightsInput
		for item, err := range listInsights(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.InsightArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				if err := resourceInsightFlatten(ctx, &item, rd); err != nil {
					tflog.Error(ctx, "Reading Security Hub Insight", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.Name)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

type listInsightModel struct {
	framework.WithRegionModel
}

func listInsights(ctx context.Context, conn *securityhub.Client, input *securityhub.GetInsightsInput) iter.Seq2[awstypes.Insight, error] {
	return func(yield func(awstypes.Insight, error) bool) {
		pages := securityhub.NewGetInsightsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.Insight](), fmt.Errorf("listing Security Hub Insights: %w", err))
				return
			}

			for _, item := range page.Insights {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
