// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
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

// @FrameworkListResource("aws_pinpointsmsvoicev2_notify_configuration")
func newNotifyConfigurationResourceAsListResource() list.ListResourceWithConfigure {
	return &notifyConfigurationListResource{}
}

var _ list.ListResource = &notifyConfigurationListResource{}

type notifyConfigurationListResource struct {
	notifyConfigurationResource
	framework.WithList
}

func (l *notifyConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input pinpointsmsvoicev2.DescribeNotifyConfigurationsInput
		for item, err := range listNotifyConfigurations(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			id := aws.ToString(item.NotifyConfigurationId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data notifyConfigurationResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = fwflex.StringValueToFramework(ctx, id)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, data.flatten(ctx, &item), smerr.ID, id)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.DisplayName)
			})

			if result.Diagnostics.HasError() {
				// The notify configuration can be deleted after it is listed.
				if request.IncludeResource {
					if _, err := findNotifyConfigurationByID(ctx, conn, id); retry.NotFound(err) {
						continue
					}
				}

				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}
			if !yield(result) {
				return
			}
		}
	}
}

func listNotifyConfigurations(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeNotifyConfigurationsInput) iter.Seq2[awstypes.NotifyConfigurationInformation, error] {
	return tfiter.ConcatValuesWithError(listNotifyConfigurationPages(ctx, conn, input))
}

func listNotifyConfigurationPages(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeNotifyConfigurationsInput) iter.Seq2[[]awstypes.NotifyConfigurationInformation, error] {
	return func(yield func([]awstypes.NotifyConfigurationInformation, error) bool) {
		pages := pinpointsmsvoicev2.NewDescribeNotifyConfigurationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.NotifyConfigurations, nil) {
				return
			}
		}
	}
}
