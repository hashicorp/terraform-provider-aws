// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_s3files_file_system")
func newFileSystemResourceAsListResource() list.ListResourceWithConfigure {
	return &fileSystemListResource{}
}

var _ list.ListResource = &fileSystemListResource{}

type fileSystemListResource struct {
	fileSystemResource
	framework.WithList
}

func (r *fileSystemListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().S3FilesClient(ctx)

	var query listFileSystemModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		input := s3files.ListFileSystemsInput{}
		output, err := conn.ListFileSystems(ctx, &input)
		if err != nil {
			result = fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		for _, fs := range output.FileSystems {
			fsID := aws.ToString(fs.FileSystemId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), fsID)

			fsOutput, err := findFileSystemByID(ctx, conn, fsID)
			if retry.NotFound(err) {
				tflog.Warn(ctx, "S3 File System not found during listing")
				continue
			}
			if err != nil {
				tflog.Error(ctx, "Reading S3 File System", map[string]any{"error": err.Error()})
				continue
			}

			var data fileSystemResourceModel
			r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
				flattenFileSystemResource(ctx, fsOutput, &data, &result.Diagnostics)
				data.ID = fwflex.StringValueToFramework(ctx, fsID)

				result.DisplayName = fsID
			})

			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listFileSystemModel struct {
	framework.WithRegionModel
}
