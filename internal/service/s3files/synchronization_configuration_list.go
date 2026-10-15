// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"

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

// @FrameworkListResource("aws_s3files_synchronization_configuration")
func newSynchronizationConfigurationResourceAsListResource() list.ListResourceWithConfigure {
	return &synchronizationConfigurationListResource{}
}

var _ list.ListResource = &synchronizationConfigurationListResource{}

type synchronizationConfigurationListResource struct {
	synchronizationConfigurationResource
	framework.WithList
}

func (r *synchronizationConfigurationListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema.Attributes = map[string]listschema.Attribute{
		names.AttrFileSystemID: listschema.StringAttribute{
			Required:    true,
			Description: "File system ID",
		},
	}
}

func (r *synchronizationConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().S3FilesClient(ctx)

	var query listSynchronizationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	fileSystemID := query.FileSystemID.ValueString()

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), fileSystemID)

		output, err := findSynchronizationConfigurationByFileSystemID(ctx, conn, fileSystemID)
		if retry.NotFound(err) {
			tflog.Warn(ctx, "Synchronization configuration not found")
			return
		}
		if err != nil {
			result = fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		var data synchronizationConfigurationResourceModel
		r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
			flattenSynchronizationConfigurationResource(ctx, output, &data, &result.Diagnostics)
			data.FileSystemID = fwflex.StringValueToFramework(ctx, fileSystemID)
			result.DisplayName = fileSystemID
		})

		yield(result)
	}
}

type listSynchronizationModel struct {
	framework.WithRegionModel
	FileSystemID types.String `tfsdk:"file_system_id"`
}
