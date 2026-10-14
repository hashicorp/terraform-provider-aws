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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_devopsagent_asset")
func newAssetResourceAsListResource() list.ListResourceWithConfigure {
	return &assetListResource{}
}

type assetListResource struct {
	assetResource
	framework.WithList
}

func (l *assetListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"agent_space_id": listschema.StringAttribute{
				Required:    true,
				Description: "Identifier of the Agent Space to list assets from.",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 2048)},
			},
			"asset_type": listschema.StringAttribute{
				Optional:    true,
				Description: "Limits results to assets of this type.",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 64)},
			},
			"updated_after": listschema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Optional:    true,
				Description: "Limits results to assets updated after this timestamp.",
			},
			"updated_before": listschema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Optional:    true,
				Description: "Limits results to assets updated before this timestamp.",
			},
		},
	}
}

func (l *assetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().DevOpsAgentClient(ctx)

	var query listAssetModel
	var diags diag.Diagnostics
	smerr.AddEnrich(ctx, &diags, request.Config.Get(ctx, &query))
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	var input devopsagent.ListAssetsInput
	smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, query, &input))
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("agent_space_id"), query.AgentSpaceID.ValueString())
	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listAssets(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}
			id := aws.ToString(item.AssetId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("asset_id"), id)
			result := request.NewListResult(ctx)
			var data assetResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.AgentSpaceID = query.AgentSpaceID
				data.AssetID = types.StringValue(id)
				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, &item, &data), smerr.ID, id)
					if result.Diagnostics.HasError() {
						return
					}
				}
				result.DisplayName = id
			})

			if !yield(result) || result.Diagnostics.HasError() {
				return
			}
		}
	}
}

type listAssetModel struct {
	framework.WithRegionModel
	AgentSpaceID  types.String      `tfsdk:"agent_space_id"`
	AssetType     types.String      `tfsdk:"asset_type"`
	UpdatedAfter  timetypes.RFC3339 `tfsdk:"updated_after"`
	UpdatedBefore timetypes.RFC3339 `tfsdk:"updated_before"`
}

func listAssets(ctx context.Context, conn *devopsagent.Client, input *devopsagent.ListAssetsInput) iter.Seq2[awstypes.Asset, error] {
	return func(yield func(awstypes.Asset, error) bool) {
		pages := devopsagent.NewListAssetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Asset{}, smarterr.NewError(err))
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
