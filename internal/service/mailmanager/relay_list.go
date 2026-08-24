// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mailmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_mailmanager_relay")
func newRelayResourceAsListResource() list.ListResourceWithConfigure {
	return &relayListResource{}
}

var _ list.ListResource = &relayListResource{}

type relayListResource struct {
	relayResource
	framework.WithList
}

type listRelayModel struct {
	framework.WithRegionModel
}

func (l *relayListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().MailManagerClient(ctx)

	var query listRelayModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input mailmanager.ListRelaysInput
		for item, err := range listRelays(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(item.RelayId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			var out *mailmanager.GetRelayOutput
			if request.IncludeResource {
				out, err = findRelayByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data relayResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = types.StringValue(id)
				data.Name = types.StringPointerValue(item.RelayName)

				if request.IncludeResource {
					result.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("Relay"))...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.RelayName)
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

func listRelays(ctx context.Context, conn *mailmanager.Client, input *mailmanager.ListRelaysInput) iter.Seq2[awstypes.Relay, error] {
	return func(yield func(awstypes.Relay, error) bool) {
		pages := mailmanager.NewListRelaysPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Relay{}, fmt.Errorf("listing SES Mail Manager Relay resources: %w", err))
				return
			}

			for _, item := range page.Relays {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
