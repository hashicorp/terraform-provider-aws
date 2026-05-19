// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_dynamodb_table")
func newTableResourceAsListResource() inttypes.ListResourceForSDK {
	l := tableListResource{}
	l.SetResourceSchema(resourceTable())
	return &l
}

type tableListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type tableListResourceModel struct {
	framework.WithRegionModel
}

func (l *tableListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.DynamoDBClient(ctx)

	var query tableListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing DynamoDB tables")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input dynamodb.ListTablesInput
		for tableName, err := range listTables(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), tableName)
			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(tableName)
			rd.Set(names.AttrName, tableName)

			if request.IncludeResource {
				table, err := findTableByName(ctx, conn, tableName)
				if err != nil {
					if retry.NotFound(err) {
						continue
					}
					result = fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading DynamoDB Table (%s): %w", tableName, err))
					yield(result)
					return
				}

				diags := resourceTableFlatten(ctx, awsClient, rd, table)
				if diags.HasError() || rd.Id() == "" {
					tflog.Error(ctx, "Flattening DynamoDB table", map[string]any{
						names.AttrName: tableName,
					})
					continue
				}
			}

			result.DisplayName = tableName

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

func listTables(ctx context.Context, conn *dynamodb.Client, input *dynamodb.ListTablesInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := dynamodb.NewListTablesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield("", fmt.Errorf("listing DynamoDB Tables: %w", err))
				return
			}

			for _, tableName := range page.TableNames {
				if !yield(tableName, nil) {
					return
				}
			}
		}
	}
}
