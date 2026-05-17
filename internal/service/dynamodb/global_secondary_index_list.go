// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_dynamodb_global_secondary_index")
func newGlobalSecondaryIndexResourceAsListResource() list.ListResourceWithConfigure {
	return &globalSecondaryIndexListResource{}
}

var _ list.ListResource = &globalSecondaryIndexListResource{}

type globalSecondaryIndexListResource struct {
	resourceGlobalSecondaryIndex
	framework.WithList
}

type globalSecondaryIndexListModel struct {
	framework.WithRegionModel
	TableName types.String `tfsdk:"table_name"`
}

func (l *globalSecondaryIndexListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrTableName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the DynamoDB table.",
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *globalSecondaryIndexListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query globalSecondaryIndexListModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	conn := l.Meta().DynamoDBClient(ctx)

	tableName := query.TableName.ValueString()

	tflog.Info(ctx, "Listing DynamoDB Global Secondary Indexes", map[string]any{
		logging.ResourceAttributeKey(names.AttrTableName): tableName,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		table, err := findTableByName(ctx, conn, tableName)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading DynamoDB Table (%s): %w", tableName, err))
			yield(result)
			return
		}

		for _, index := range table.GlobalSecondaryIndexes {
			indexName := aws.ToString(index.IndexName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("index_name"), indexName)

			result := request.NewListResult(ctx)

			var data resourceGlobalSecondaryIndexModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(flattenGlobalSecondaryIndex(ctx, &data, &index, table)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = indexName
			})

			if !yield(result) {
				return
			}
		}
	}
}
