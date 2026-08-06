// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package mailmanager

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mailmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_mailmanager_ingress_point")
func newIngressPointResourceAsListResource() list.ListResourceWithConfigure {
	return &ingressPointListResource{}
}

var _ list.ListResource = &ingressPointListResource{}

type ingressPointListResource struct {
	ingressPointResource
	framework.WithList
}

type listIngressPointModel struct {
	framework.WithRegionModel
}

func (l *ingressPointListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().MailManagerClient(ctx)

	var query listIngressPointModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input mailmanager.ListIngressPointsInput
		for item, err := range listIngressPoints(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(item.IngressPointId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			var out *mailmanager.GetIngressPointOutput
			if request.IncludeResource {
				out, err = findIngressPointByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data ingressPointResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = types.StringValue(id)
				data.Name = types.StringPointerValue(item.IngressPointName)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, out, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}
				result.DisplayName = aws.ToString(item.IngressPointName)
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

func listIngressPoints(ctx context.Context, conn *mailmanager.Client, input *mailmanager.ListIngressPointsInput) iter.Seq2[awstypes.IngressPoint, error] {
	return func(yield func(awstypes.IngressPoint, error) bool) {
		pages := mailmanager.NewListIngressPointsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.IngressPoint{}, smarterr.NewError(err))
				return
			}
			for _, item := range page.IngressPoints {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
