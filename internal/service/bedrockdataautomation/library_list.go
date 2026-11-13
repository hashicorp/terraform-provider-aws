// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockdataautomation

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types"
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

// @FrameworkListResource("aws_bedrockdataautomation_library")
func newLibraryResourceAsListResource() list.ListResourceWithConfigure {
	return &libraryListResource{}
}

var _ list.ListResource = &libraryListResource{}

type libraryListResource struct {
	libraryResource
	framework.WithList
}

func (l *libraryListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockDataAutomationClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		var input bedrockdataautomation.ListDataAutomationLibrariesInput
		for item, err := range listLibraries(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			arn := aws.ToString(item.LibraryArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			var library *awstypes.DataAutomationLibrary
			if request.IncludeResource {
				var err error
				library, err = findLibraryByARN(ctx, conn, arn)
				if retry.NotFound(err) {
					continue
				}
				// Tags can't be read from a library that is being deleted.
				if err == nil && library.Status == awstypes.DataAutomationLibraryStatusDeleting {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, arn))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data libraryResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ARN = fwflex.StringValueToFramework(ctx, arn)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, library, &data), smerr.ID, arn)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.LibraryName)
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

func listLibraries(ctx context.Context, conn *bedrockdataautomation.Client, input *bedrockdataautomation.ListDataAutomationLibrariesInput) iter.Seq2[awstypes.DataAutomationLibrarySummary, error] {
	return tfiter.ConcatValuesWithError(listLibraryPages(ctx, conn, input))
}

func listLibraryPages(ctx context.Context, conn *bedrockdataautomation.Client, input *bedrockdataautomation.ListDataAutomationLibrariesInput) iter.Seq2[[]awstypes.DataAutomationLibrarySummary, error] {
	return func(yield func([]awstypes.DataAutomationLibrarySummary, error) bool) {
		pages := bedrockdataautomation.NewListDataAutomationLibrariesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, smarterr.NewError(err))
				return
			}

			if !yield(page.Libraries, nil) {
				return
			}
		}
	}
}
