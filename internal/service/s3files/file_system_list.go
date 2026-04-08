// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @FrameworkListResource("aws_s3files_file_system")
func newFileSystemResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceFileSystem{}
}

var _ list.ListResource = &listResourceFileSystem{}

type listResourceFileSystem struct {
	fileSystemResource
	framework.WithList
}

func (r *listResourceFileSystem) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := r.Meta()
	conn := awsClient.S3FilesClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		for fs, err := range listFileSystems(ctx, conn) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data fileSystemResourceModel
			r.SetResult(ctx, awsClient, request.IncludeResource, &data, &result, func() {
				data.ID = types.StringPointerValue(fs.FileSystemId)
				data.FileSystemId = types.StringPointerValue(fs.FileSystemId)
				data.FileSystemArn = types.StringPointerValue(fs.FileSystemArn)
				result.DisplayName = data.FileSystemId.ValueString()
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

func listFileSystems(ctx context.Context, conn *s3files.Client) iter.Seq2[awstypes.ListFileSystemsDescription, error] {
	return func(yield func(awstypes.ListFileSystemsDescription, error) bool) {
		pages := s3files.NewListFileSystemsPaginator(conn, &s3files.ListFileSystemsInput{})
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ListFileSystemsDescription{}, fmt.Errorf("listing S3 Files File Systems: %w", err))
				return
			}
			for _, fs := range page.FileSystems {
				if !yield(fs, nil) {
					return
				}
			}
		}
	}
}
