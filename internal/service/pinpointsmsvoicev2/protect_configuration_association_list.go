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
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_pinpointsmsvoicev2_protect_configuration_association")
func newProtectConfigurationAssociationResourceAsListResource() list.ListResourceWithConfigure {
	return &protectConfigurationAssociationListResource{}
}

var _ list.ListResource = &protectConfigurationAssociationListResource{}

type protectConfigurationAssociationListResource struct {
	protectConfigurationAssociationResource
	framework.WithList
}

func (l *protectConfigurationAssociationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input pinpointsmsvoicev2.DescribeConfigurationSetsInput
		for item, err := range listConfigurationSets(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// Only configuration sets with a protect configuration have an association.
			if aws.ToString(item.ProtectConfigurationId) == "" {
				continue
			}

			name := aws.ToString(item.ConfigurationSetName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("configuration_set_name"), name)

			result := request.NewListResult(ctx)

			var data protectConfigurationAssociationResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, fwflex.Flatten(ctx, &item, &data), smerr.ID, name)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = name
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

func listConfigurationSets(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeConfigurationSetsInput) iter.Seq2[awstypes.ConfigurationSetInformation, error] {
	return tfiter.ConcatValuesWithError(listConfigurationSetPages(ctx, conn, input))
}

func listConfigurationSetPages(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeConfigurationSetsInput) iter.Seq2[[]awstypes.ConfigurationSetInformation, error] {
	return func(yield func([]awstypes.ConfigurationSetInformation, error) bool) {
		pages := pinpointsmsvoicev2.NewDescribeConfigurationSetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.ConfigurationSets, nil) {
				return
			}
		}
	}
}
