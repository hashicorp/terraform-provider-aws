// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"

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

// @FrameworkListResource("aws_s3files_access_point")
func newAccessPointResourceAsListResource() list.ListResourceWithConfigure {
	return &accessPointListResource{}
}

var _ list.ListResource = &accessPointListResource{}

type accessPointListResource struct {
	accessPointResource
	framework.WithList
}

func (r *accessPointListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema.Attributes = map[string]listschema.Attribute{
		names.AttrFileSystemID: listschema.StringAttribute{
			Required:    true,
			Description: "File system ID",
		},
	}
}

func (r *accessPointListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().S3FilesClient(ctx)

	var query listAccessPointModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)

		paginator := s3files.NewListAccessPointsPaginator(conn, &s3files.ListAccessPointsInput{
			FileSystemId: query.FileSystemID.ValueStringPointer(),
		})
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, item := range page.AccessPoints {
				itemID := aws.ToString(item.AccessPointId)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), itemID)

				itemOutput, err := findAccessPointByID(ctx, conn, itemID)
				if retry.NotFound(err) {
					tflog.Warn(ctx, "Resource not found during listing")
					continue
				}
				if err != nil {
					tflog.Error(ctx, "Reading resource", map[string]any{"error": err.Error()})
					continue
				}

				var data accessPointResourceModel
				r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
					flattenAccessPointResource(ctx, itemOutput, &data, false, &result.Diagnostics)
					data.ID = fwflex.StringValueToFramework(ctx, itemID)
					result.DisplayName = itemID
				})

				if !yield(result) {
					return
				}
			}
		}
	}
}

type listAccessPointModel struct {
	framework.WithRegionModel
	FileSystemID types.String `tfsdk:"file_system_id"`
}
