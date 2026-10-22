// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_securityhub_insight")
func newInsightResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceInsight(), framework.SDKListSpec[awstypes.Insight, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.Insight, error], diag.Diagnostics) {
			var input securityhub.GetInsightsInput

			return listInsights(ctx, c.SecurityHubClient(ctx), &input), nil
		},
		ID: func(i awstypes.Insight, _ framework.RegionQuery) string {
			return aws.ToString(i.InsightArn)
		},
		Flatten:     resourceInsightFlatten,
		DisplayName: func(i awstypes.Insight) string { return aws.ToString(i.Name) },
	})
}

func listInsights(ctx context.Context, conn *securityhub.Client, input *securityhub.GetInsightsInput) iter.Seq2[awstypes.Insight, error] {
	return tfiter.PageItems[securityhub.Options](ctx, "listing Security Hub Insights", securityhub.NewGetInsightsPaginator(conn, input),
		func(o *securityhub.GetInsightsOutput) []awstypes.Insight { return o.Insights })
}
