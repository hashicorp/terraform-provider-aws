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
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_pinpointsmsvoicev2_phone_number")
func newPhoneNumberResourceAsListResource() list.ListResourceWithConfigure {
	return &phoneNumberListResource{}
}

var _ list.ListResource = &phoneNumberListResource{}

type phoneNumberListResource struct {
	phoneNumberResource
	framework.WithList
}

func (l *phoneNumberListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	tflog.Info(ctx, "Listing End User Messaging SMS Phone Numbers")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := pinpointsmsvoicev2.DescribePhoneNumbersInput{}
		phoneNumbers, err := findPhoneNumbers(ctx, conn, &input)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		for i := range phoneNumbers {
			item := &phoneNumbers[i]
			id := aws.ToString(item.PhoneNumberId)

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			var data phoneNumberResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, data.flatten(ctx, item), smerr.ID, id)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.PhoneNumber)
			})

			if !yield(result) {
				return
			}
		}
	}
}
