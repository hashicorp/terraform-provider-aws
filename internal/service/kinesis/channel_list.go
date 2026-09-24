// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_kinesis_channel")
func newChannelResourceAsListResource() list.ListResourceWithConfigure {
	return &channelListResource{}
}

var _ list.ListResource = &channelListResource{}

type channelListResource struct {
	channelResource
	framework.WithList
}

func (l *channelListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().KinesisClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input kinesis.ListChannelsInput
		for item, err := range listChannels(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			if item.ChannelStatus == awstypes.ChannelStatusCreating || item.ChannelStatus == awstypes.ChannelStatusDeleting {
				continue
			}

			arn := aws.ToString(item.ChannelARN)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var out *awstypes.ChannelDescription
			if request.IncludeResource {
				out, err = findChannelByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
				if out.ChannelStatus == awstypes.ChannelStatusCreating || out.ChannelStatus == awstypes.ChannelStatusDeleting {
					continue
				}
			}

			result := request.NewListResult(ctx)
			var data channelResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = types.StringValue(arn)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, out, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.ChannelName)
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

func listChannels(ctx context.Context, conn *kinesis.Client, input *kinesis.ListChannelsInput) iter.Seq2[awstypes.ChannelSummary, error] {
	return func(yield func(awstypes.ChannelSummary, error) bool) {
		pages := kinesis.NewListChannelsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ChannelSummary{}, fmt.Errorf("listing Kinesis Channel resources: %w", err))
				return
			}

			for _, item := range page.ChannelSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
