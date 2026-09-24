// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_directory_service_ip_routes")
func newIPRoutesResourceAsListResource() list.ListResourceWithConfigure {
	return &ipRoutesListResource{}
}

var _ list.ListResource = &ipRoutesListResource{}

type ipRoutesListResource struct {
	ipRoutesResource
	framework.WithList
}

func (l *ipRoutesListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().DSClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		for directory, err := range listDirectories(ctx, conn) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			directoryID := aws.ToString(directory.DirectoryId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("directory_id"), directoryID)

			// The IP routes resource for a directory exists only if the directory
			// has at least one route. Query them to determine existence and, when
			// requested, to populate the resource.
			routes, err := findIPRoutesByDirectoryID(ctx, conn, directoryID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				yield(smerr.NewListResultError(ctx, err, smerr.ID, directoryID))
				return
			}

			result := request.NewListResult(ctx)

			var data ipRoutesResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.DirectoryID = types.StringValue(directoryID)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, routes, &data), smerr.ID, directoryID)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(directory.Name)
			})

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listDirectories(ctx context.Context, conn *directoryservice.Client) iter.Seq2[awstypes.DirectoryDescription, error] {
	return func(yield func(awstypes.DirectoryDescription, error) bool) {
		input := directoryservice.DescribeDirectoriesInput{}
		pages := directoryservice.NewDescribeDirectoriesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.DirectoryDescription{}, smarterr.NewError(err))
				return
			}

			for _, item := range page.DirectoryDescriptions {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
