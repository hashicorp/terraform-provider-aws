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
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_mailmanager_archive")
func newArchiveResourceAsListResource() list.ListResourceWithConfigure {
	return &archiveListResource{}
}

var _ list.ListResource = &archiveListResource{}

type archiveListResource struct {
	archiveResource
	framework.WithList
}

type listArchiveModel struct {
	framework.WithRegionModel
}

func (l *archiveListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().MailManagerClient(ctx)

	var query listArchiveModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input mailmanager.ListArchivesInput
		for item, err := range listArchives(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			if item.ArchiveState == awstypes.ArchiveStatePendingDeletion {
				continue
			}

			id := aws.ToString(item.ArchiveId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			var out *mailmanager.GetArchiveOutput
			if request.IncludeResource {
				out, err = findArchiveByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data archiveResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = types.StringValue(id)
				data.Name = types.StringPointerValue(item.ArchiveName)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, out, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.ArchiveName)
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

func listArchives(ctx context.Context, conn *mailmanager.Client, input *mailmanager.ListArchivesInput) iter.Seq2[awstypes.Archive, error] {
	return func(yield func(awstypes.Archive, error) bool) {
		pages := mailmanager.NewListArchivesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Archive{}, fmt.Errorf("listing SES Mail Manager Archive resources: %w", err))
				return
			}
			for _, item := range page.Archives {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
