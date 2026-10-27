// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_directory_service_directory_setting")
func newDirectorySettingsResourceAsListResource() list.ListResourceWithConfigure {
	return &directorySettingsListResource{}
}

var _ list.ListResource = &directorySettingsListResource{}

type directorySettingsListResource struct {
	directorySettingsResource
	framework.WithList
}

func (l *directorySettingsListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"directory_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the directory to list Settings for.",
				Validators: []validator.String{
					directoryIDValidator,
				},
			},
		},
	}
}

func (l *directorySettingsListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().DSClient(ctx)

	var query listDirectorySettingsModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	directoryID := fwflex.StringValueFromFramework(ctx, query.DirectoryID)
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("directory_id"), directoryID)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("directory_id"): directoryID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		entries, err := findDirectorySettingsByDirectoryID(ctx, conn, directoryID)
		if err != nil {
			yield(smerr.NewListResultError(ctx, err, smerr.ID, directoryID))
			return
		}

		// Each requested (non-default) setting maps to its own
		// aws_directory_service_directory_setting resource instance.
		for _, entry := range entries {
			if entry.RequestStatus == awstypes.DirectoryConfigurationStatusDefault {
				continue
			}

			result := request.NewListResult(ctx)

			var data directorySettingsResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.DirectoryID = fwflex.StringValueToFramework(ctx, directoryID)
				data.Name = fwflex.StringToFramework(ctx, entry.Name)

				if request.IncludeResource {
					l.flatten(ctx, &entry, &data)
				}

				result.DisplayName = fwflex.StringValueFromFramework(ctx, data.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

// listDirectorySettingsModel is the List Resource's configuration schema data.
type listDirectorySettingsModel struct {
	framework.WithRegionModel
	DirectoryID types.String `tfsdk:"directory_id"`
}
