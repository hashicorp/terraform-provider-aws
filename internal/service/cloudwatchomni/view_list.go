// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_cloudwatchomni_view")
func newViewResourceAsListResource() list.ListResourceWithConfigure {
	return &viewListResource{}
}

var _ list.ListResource = &viewListResource{}

type viewListResource struct {
	viewResource
	framework.WithList
}

func (l *viewListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchOmniClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		// Managed views are provisioned by AWS and cannot be managed with this resource type.
		input := cloudwatchomni.ListViewsInput{
			Type: awstypes.ViewTypeUser,
		}
		for item, err := range listViews(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			name := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			var output *cloudwatchomni.GetViewOutput
			if request.IncludeResource {
				var err error
				output, err = findViewByName(ctx, conn, name)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data viewResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.Name = fwflex.StringValueToFramework(ctx, name)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = name
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listViews(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListViewsInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[awstypes.ViewSummary, error] {
	return tfiter.ConcatValuesWithError(listViewPages(ctx, conn, input, optFns...))
}

func listViewPages(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListViewsInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[[]awstypes.ViewSummary, error] {
	return func(yield func([]awstypes.ViewSummary, error) bool) {
		pages := cloudwatchomni.NewListViewsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Omni Views: %w", err))
				return
			}

			if !yield(page.Items, nil) {
				return
			}
		}
	}
}
