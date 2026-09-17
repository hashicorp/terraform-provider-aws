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

// @FrameworkListResource("aws_s3files_file_system_policy")
func newFileSystemPolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &fileSystemPolicyListResource{}
}

var _ list.ListResource = &fileSystemPolicyListResource{}

type fileSystemPolicyListResource struct {
	fileSystemPolicyResource
	framework.WithList
}

func (r *fileSystemPolicyListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrFileSystemID: listschema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *fileSystemPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().S3FilesClient(ctx)

	var query listFileSystemPolicyModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		fsID := query.FileSystemID.ValueString()
		ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), fsID)

		policyOutput, err := findFileSystemPolicyByID(ctx, conn, fsID)
		if retry.NotFound(err) {
			tflog.Warn(ctx, "File System Policy not found during listing")
			return
		}
		if err != nil {
			result = fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		var data fileSystemPolicyResourceModel
		r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
			flattenFileSystemPolicyResource(ctx, policyOutput, &data, &result.Diagnostics)
			data.FileSystemID = fwflex.StringValueToFramework(ctx, fsID)
			result.DisplayName = fsID
		})

		if result.Diagnostics.HasError() {
			yield(result)
			return
		}

		yield(result)
	}
}

type listFileSystemPolicyModel struct {
	framework.WithRegionModel
	FileSystemID types.String `tfsdk:"file_system_id"`
}
