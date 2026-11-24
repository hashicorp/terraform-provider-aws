// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsagent/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_devopsagent_asset_file")
func newAssetFileResourceAsListResource() list.ListResourceWithConfigure {
	return &assetFileListResource{}
}

type assetFileListResource struct {
	assetFileResource
	framework.WithList
}

func (l *assetFileListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"agent_space_id": listschema.StringAttribute{
				Required:    true,
				Description: "Identifier of the Agent Space containing the asset.",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 2048)},
			},
			"asset_id": listschema.StringAttribute{
				Required:    true,
				Description: "Identifier of the asset to list files from.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
		},
	}
}

func (l *assetFileListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().DevOpsAgentClient(ctx)

	var query listAssetFileModel
	var diags diag.Diagnostics
	smerr.AddEnrich(ctx, &diags, request.Config.Get(ctx, &query))
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	var input devopsagent.ListAssetFilesInput
	smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, query, &input))
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	agentSpaceID, assetID := query.AgentSpaceID.ValueString(), query.AssetID.ValueString()
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("agent_space_id"), agentSpaceID)
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("asset_id"), assetID)
	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listAssetFiles(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}
			filePath := aws.ToString(item.Path)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrPath), filePath)

			// ListAssetFiles returns summaries without content, so fetch the full file.
			var file *awstypes.AssetFile
			if request.IncludeResource {
				file, err = findAssetFileByThreePartKey(ctx, conn, agentSpaceID, assetID, filePath)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err))
					return
				}
			}

			result := request.NewListResult(ctx)
			var data assetFileResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.AgentSpaceID = query.AgentSpaceID
				data.AssetID = query.AssetID
				data.Path = types.StringValue(filePath)
				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, file, &data), smerr.ID, filePath)
					if result.Diagnostics.HasError() {
						return
					}
				}
				result.DisplayName = filePath
			})

			if !yield(result) || result.Diagnostics.HasError() {
				return
			}
		}
	}
}

type listAssetFileModel struct {
	framework.WithRegionModel
	AgentSpaceID types.String `tfsdk:"agent_space_id"`
	AssetID      types.String `tfsdk:"asset_id"`
}

func listAssetFiles(ctx context.Context, conn *devopsagent.Client, input *devopsagent.ListAssetFilesInput) iter.Seq2[awstypes.AssetFileSummary, error] {
	return func(yield func(awstypes.AssetFileSummary, error) bool) {
		pages := devopsagent.NewListAssetFilesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AssetFileSummary{}, smarterr.NewError(err))
				return
			}
			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
