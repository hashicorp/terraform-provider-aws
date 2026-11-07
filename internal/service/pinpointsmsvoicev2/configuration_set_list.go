// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_pinpointsmsvoicev2_configuration_set")
func newConfigurationSetResourceAsListResource() list.ListResourceWithConfigure {
	return &configurationSetListResource{}
}

var _ list.ListResource = &configurationSetListResource{}

type configurationSetListResource struct {
	configurationSetResource
	framework.WithList
}

func (l *configurationSetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	tflog.Info(ctx, "Listing End User Messaging SMS Configuration Sets")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := pinpointsmsvoicev2.DescribeConfigurationSetsInput{}
		configurationSets, err := findConfigurationSets(ctx, conn, &input)
		if err != nil {
			yield(smerr.NewListResultError(ctx, err))
			return
		}

		for i := range configurationSets {
			item := &configurationSets[i]
			name := aws.ToString(item.ConfigurationSetName)

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			var data configurationSetResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, item, &data))
				if result.Diagnostics.HasError() {
					return
				}

				data.setID()

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
