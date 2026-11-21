// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_apigatewayv2_portal")
func newPortalResourceAsListResource() list.ListResourceWithConfigure {
	return &portalListResource{}
}

var _ list.ListResource = &portalListResource{}

type portalListResource struct {
	portalResource
	framework.WithList
}

func (l *portalListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().APIGatewayV2Client(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input apigatewayv2.ListPortalsInput
		for item, err := range listPortals(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(item.PortalId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("portal_id"), id)

			var output *apigatewayv2.GetPortalOutput
			if request.IncludeResource {
				var err error
				output, err = findPortalByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data portalResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
				} else {
					data.PortalID = fwflex.StringToFramework(ctx, item.PortalId)
				}
				if result.Diagnostics.HasError() {
					return
				}

				if v := item.PortalContent; v != nil {
					result.DisplayName = aws.ToString(v.DisplayName)
				}
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listPortals(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.ListPortalsInput) iter.Seq2[awstypes.PortalSummary, error] {
	return func(yield func(awstypes.PortalSummary, error) bool) {
		var stopped bool
		err := listPortalsPages(ctx, conn, input, func(page *apigatewayv2.ListPortalsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}
			for _, item := range page.Items {
				if !yield(item, nil) {
					stopped = true
					return false
				}
			}
			return !lastPage
		})
		if !stopped && err != nil {
			yield(inttypes.Zero[awstypes.PortalSummary](), fmt.Errorf("listing API Gateway V2 Portals: %w", err))
		}
	}
}
