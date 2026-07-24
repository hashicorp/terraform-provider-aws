// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_cloudfront_distribution")
func newDistributionResourceAsListResource() inttypes.ListResourceForSDK {
	l := distributionListResource{}
	l.SetResourceSchema(resourceDistribution())
	return &l
}

var _ list.ListResource = &distributionListResource{}

type distributionListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *distributionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.CloudFrontClient(ctx)

	var query listDistributionModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing CloudFront Distributions")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input cloudfront.ListDistributionsInput

		for item, err := range listDistributions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.Id)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)

			output, err := findDistributionByID(ctx, conn, id)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				tflog.Error(ctx, "Reading CloudFront Distribution", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			if connectionMode := output.Distribution.DistributionConfig.ConnectionMode; connectionMode == awstypes.ConnectionModeTenantOnly {
				continue
			}

			if request.IncludeResource {
				if err := resourceDistributionFlatten(ctx, awsClient, output, rd); err != nil {
					tflog.Error(ctx, "Reading CloudFront Distribution", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = distributionListDisplayName(output)

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

type listDistributionModel struct{}

func listDistributions(ctx context.Context, conn *cloudfront.Client, input *cloudfront.ListDistributionsInput) iter.Seq2[awstypes.DistributionSummary, error] {
	return func(yield func(awstypes.DistributionSummary, error) bool) {
		pages := cloudfront.NewListDistributionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.DistributionSummary{}, fmt.Errorf("listing CloudFront Distribution resources: %w", err))
				return
			}

			if page.DistributionList != nil {
				for _, item := range page.DistributionList.Items {
					if !yield(item, nil) {
						return
					}
				}
			}
		}
	}
}

func distributionListDisplayName(output *cloudfront.GetDistributionOutput) string {
	if comment := aws.ToString(output.Distribution.DistributionConfig.Comment); comment != "" {
		return comment
	}

	if domainName := aws.ToString(output.Distribution.DomainName); domainName != "" {
		return domainName
	}

	return aws.ToString(output.Distribution.Id)
}
