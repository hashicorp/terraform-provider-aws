// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_cloudwatch_log_group")
func newLogGroupResourceAsListResource() inttypes.ListResourceForSDK {
	l := logGroupListResource{}
	l.SetResourceSchema(resourceGroup())

	return &l
}

type logGroupListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type logGroupListResourceModel struct {
	framework.WithRegionModel
}

func (l *logGroupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.LogsClient(ctx)

	var query logGroupListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input cloudwatchlogs.DescribeLogGroupsInput
		for output, err := range listLogGroups(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.LogGroup]()) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			rd := l.ResourceData()
			rd.SetId(aws.ToString(output.LogGroupName))
			resourceGroupFlatten(ctx, rd, output)

			result.DisplayName = aws.ToString(output.LogGroupName)

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
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

func listLogGroups(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogGroupsInput, filter tfslices.Predicate[*awstypes.LogGroup]) iter.Seq2[awstypes.LogGroup, error] {
	return func(yield func(awstypes.LogGroup, error) bool) {
		pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.LogGroup{}, fmt.Errorf("listing CloudWatch Logs Log Groups: %w", err))
				return
			}

			for _, v := range page.LogGroups {
				if filter(&v) {
					if !yield(v, nil) {
						return
					}
				}
			}
		}
	}
}
