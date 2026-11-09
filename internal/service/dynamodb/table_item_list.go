// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_dynamodb_table_item")
func newTableItemResourceAsListResource() inttypes.ListResourceForSDK {
	l := tableItemListResource{}
	l.SetResourceSchema(resourceTableItem())
	return &l
}

var (
	_ list.ListResource                 = &tableItemListResource{}
	_ list.ListResourceWithRawV5Schemas = &tableItemListResource{}
)

type tableItemListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type tableItemListResourceModel struct {
	framework.WithRegionModel
	TableName types.String `tfsdk:"table_name"`
}

func (l *tableItemListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrTableName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the DynamoDB table to list items from.",
			},
		},
	}
}

func (l *tableItemListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.DynamoDBClient(ctx)

	var query tableItemListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tableName := query.TableName.ValueString()

	table, err := findTableByName(ctx, conn, tableName)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading DynamoDB Table (%s): %w", tableName, err)).Diagnostics)
		return
	}

	var hashKey, rangeKey string
	for _, attribute := range table.KeySchema {
		switch attribute.KeyType {
		case awstypes.KeyTypeHash:
			hashKey = aws.ToString(attribute.AttributeName)
		case awstypes.KeyTypeRange:
			rangeKey = aws.ToString(attribute.AttributeName)
		}
	}

	tflog.Info(ctx, "Listing DynamoDB Table Items", map[string]any{
		names.AttrTableName: tableName,
	})
	stream.Results = func(yield func(list.ListResult) bool) {
		input := dynamodb.ScanInput{
			ConsistentRead: aws.Bool(true),
			TableName:      aws.String(tableName),
		}
		for item, err := range scanTableItems(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			hkv, rkv := tableItemKeyValues(item, hashKey, rangeKey)
			id := tableItemCreateResourceID(tableName, hkv, rkv)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()

			if flattenDiags := resourceTableItemFlatten(rd, tableName, hashKey, rangeKey, item); flattenDiags.HasError() {
				tflog.Error(ctx, "Flattening DynamoDB Table Item", map[string]any{
					names.AttrID: id,
				})
				continue
			}

			result.DisplayName = id

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

func scanTableItems(ctx context.Context, conn *dynamodb.Client, input *dynamodb.ScanInput) iter.Seq2[map[string]awstypes.AttributeValue, error] {
	return func(yield func(map[string]awstypes.AttributeValue, error) bool) {
		pages := dynamodb.NewScanPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(nil, fmt.Errorf("scanning DynamoDB Table Items: %w", err))
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
