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

// @FrameworkListResource("aws_pinpointsmsvoicev2_protect_configuration")
func newProtectConfigurationResourceAsListResource() list.ListResourceWithConfigure {
	return &protectConfigurationListResource{}
}

var _ list.ListResource = &protectConfigurationListResource{}

type protectConfigurationListResource struct {
	protectConfigurationResource
	framework.WithList
}

func (l *protectConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input pinpointsmsvoicev2.DescribeProtectConfigurationsInput
		for item, err := range listProtectConfigurations(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			id := aws.ToString(item.ProtectConfigurationId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data protectConfigurationResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = fwflex.StringValueToFramework(ctx, id)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, data.flatten(ctx, conn, &item), smerr.ID, id)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = id
			})

			if result.Diagnostics.HasError() {
				// The protect configuration can be deleted after it is listed.
				if request.IncludeResource {
					if _, err := findProtectConfigurationByID(ctx, conn, id); retry.NotFound(err) {
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

func listProtectConfigurations(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeProtectConfigurationsInput) iter.Seq2[awstypes.ProtectConfigurationInformation, error] {
	return tfiter.ConcatValuesWithError(listProtectConfigurationPages(ctx, conn, input))
}

func listProtectConfigurationPages(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeProtectConfigurationsInput) iter.Seq2[[]awstypes.ProtectConfigurationInformation, error] {
	return func(yield func([]awstypes.ProtectConfigurationInformation, error) bool) {
		pages := pinpointsmsvoicev2.NewDescribeProtectConfigurationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.ProtectConfigurations, nil) {
				return
			}
		}
	}
}
