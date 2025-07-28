// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_dynamodb_tables", name="Tables")
func newTablesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &tablesDataSource{}, nil
}

type tablesDataSource struct {
	framework.DataSourceWithModel[tablesDataSourceModel]
}

func (d *tablesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrNames: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *tablesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data tablesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().DynamoDBClient(ctx)

	var input dynamodb.ListTablesInput
	out, err := findTables(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("reading DynamoDB Tables", err.Error())

		return
	}

	data.Names = fwflex.FlattenFrameworkStringValueListOfString(ctx, out)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findTables(ctx context.Context, conn *dynamodb.Client, input *dynamodb.ListTablesInput) ([]string, error) {
	var output []string

	pages := dynamodb.NewListTablesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.TableNames...)
	}

	return output, nil
}

type tablesDataSourceModel struct {
	framework.WithRegionModel
	Names fwtypes.ListOfString `tfsdk:"names"`
}
