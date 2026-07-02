// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_s3files_mount_target")
func newMountTargetResourceAsListResource() list.ListResourceWithConfigure {
	return &mountTargetListResource{}
}

var _ list.ListResource = &mountTargetListResource{}

type mountTargetListResource struct {
	mountTargetResource
	framework.WithList
}

func (r *mountTargetListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema.Attributes = map[string]listschema.Attribute{
		names.AttrFileSystemID: listschema.StringAttribute{
			Required:    true,
			Description: "File system ID",
		},
	}
}

func (r *mountTargetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().S3FilesClient(ctx)

	var query listMountTargetModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)

		if query.FileSystemID.IsNull() {
			result = fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("%s is required", names.AttrFileSystemID))
			yield(result)
			return
		}

		input := s3files.ListMountTargetsInput{
			FileSystemId: query.FileSystemID.ValueStringPointer(),
		}
		paginator := s3files.NewListMountTargetsPaginator(conn, &input)

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, mt := range page.MountTargets {
				mtID := aws.ToString(mt.MountTargetId)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), mtID)

				mtOutput, err := findMountTargetByID(ctx, conn, mtID)
				if retry.NotFound(err) {
					tflog.Warn(ctx, "S3 Files Mount Target not found during listing", map[string]any{
						names.AttrID: mtID,
					})
					continue
				}
				if err != nil {
					tflog.Error(ctx, "Reading S3 Files Mount Target", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				var data mountTargetResourceModel
				r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
					flattenMountTargetResource(ctx, mtOutput, &data, &result.Diagnostics)
					data.ID = fwflex.StringValueToFramework(ctx, mtID)

					result.DisplayName = mtID
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
}

type listMountTargetModel struct {
	framework.WithRegionModel
	FileSystemID types.String `tfsdk:"file_system_id"`
}
