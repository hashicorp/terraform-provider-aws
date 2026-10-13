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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_pinpointsmsvoicev2_opt_out_list")
func newOptOutListResourceAsListResource() list.ListResourceWithConfigure {
	return &optOutListListResource{}
}

var _ list.ListResource = &optOutListListResource{}

type optOutListListResource struct {
	optOutListResource
	framework.WithList
}

func (l *optOutListListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	tflog.Info(ctx, "Listing End User Messaging SMS Opt-out Lists")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := pinpointsmsvoicev2.DescribeOptOutListsInput{}
		optOutLists, err := findOptOutLists(ctx, conn, &input)
		if err != nil {
			yield(fwdiag.NewListResultErrorDiagnostic(err))
			return
		}

		for i := range optOutLists {
			item := &optOutLists[i]
			name := aws.ToString(item.OptOutListName)

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			var data optOutListResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, fwflex.Flatten(ctx, item, &data))
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = name
			})

			if !yield(result) {
				return
			}
		}
	}
}
