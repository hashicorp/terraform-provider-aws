// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// @FrameworkListResource("aws_directory_service_directory_settings")
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
			// Not all directory types (e.g. SimpleAD) support directory settings.
			tflog.Debug(ctx, "Reading Directory Service Directory Settings", map[string]any{
				"error": err.Error(),
			})
			return
		}

		// Only directories with at least one requested (non-default) setting have
		// a corresponding aws_directory_service_directory_settings resource.
		requested := tfslices.Filter(entries, func(e awstypes.SettingEntry) bool {
			return e.RequestedValue != nil
		})
		if len(requested) == 0 {
			tflog.Debug(ctx, "Directory has no requested Settings, skipping")
			return
		}

		result := request.NewListResult(ctx)

		var data directorySettingsResourceModel

		l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
			data.DirectoryID = fwflex.StringValueToFramework(ctx, directoryID)

			if request.IncludeResource {
				result.Diagnostics.Append(flattenDirectorySettingsForList(ctx, requested, &data)...)
				if result.Diagnostics.HasError() {
					return
				}
			}

			result.DisplayName = directoryID
		})

		yield(result)
	}
}

// listDirectorySettingsModel is the List Resource's configuration schema data.
type listDirectorySettingsModel struct {
	framework.WithRegionModel
	DirectoryID types.String `tfsdk:"directory_id"`
}

// flattenDirectorySettingsForList builds the `setting` list for the List Resource from
// the settings that have been explicitly requested for the directory.
//
// This is intentionally separate from mergeSettingEntries, which is used by
// Create/Read/Update: those merge against a known list of requested settings from
// plan or state, while listing has no such prior knowledge and instead relies on the
// API's RequestedValue field to identify managed settings.
func flattenDirectorySettingsForList(ctx context.Context, entries []awstypes.SettingEntry, data *directorySettingsResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	models := make([]*directorySettingModel, len(entries))
	for i, e := range entries {
		models[i] = &directorySettingModel{
			Name:          fwflex.StringToFramework(ctx, e.Name),
			Value:         fwflex.StringToFramework(ctx, e.RequestedValue),
			AppliedValue:  fwflex.StringToFramework(ctx, e.AppliedValue),
			RequestStatus: fwflex.StringValueToFramework(ctx, e.RequestStatus),
			Type:          fwflex.StringToFramework(ctx, e.Type),
		}
	}

	settings, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, models, nil)
	diags.Append(d...)
	data.Settings = settings

	return diags
}
