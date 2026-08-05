// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
)

// @FrameworkListResource("aws_pinpointsmsvoicev2_sender_id")
func newSenderIDResourceAsListResource() list.ListResourceWithConfigure {
	return &senderIDListResource{}
}

var _ list.ListResource = &senderIDListResource{}

type senderIDListResource struct {
	senderIDResource
	framework.WithList
}

func (l *senderIDListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	var query listSenderIDModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing End User Messaging SMS Sender IDs")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := pinpointsmsvoicev2.DescribeSenderIdsInput{}
		senderIDs, err := findSenderIDs(ctx, conn, &input)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		for i := range senderIDs {
			item := &senderIDs[i]
			senderID := aws.ToString(item.SenderId)
			isoCountryCode := aws.ToString(item.IsoCountryCode)

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("sender_id"), senderID)
			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("iso_country_code"), isoCountryCode)

			result := request.NewListResult(ctx)

			var data senderIDResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(data.flatten(ctx, item)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = senderID
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listSenderIDModel struct {
	framework.WithRegionModel
}
