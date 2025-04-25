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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_dynamodb_tables", name="Tables")
func newDataSourceTables(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceTables{}, nil
}

const (
	DSNameTables = "Tables Data Source"
)

type dataSourceTables struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceTables) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrIDs: schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceTables) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	conn := d.Meta().DynamoDBClient(ctx)

	var data dataSourceTablesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := dynamodb.ListTablesInput{}
	out, err := findTables(ctx, conn, &input)

	if err != nil {
		resp.Diagnostics.AddError("reading DynamoDB Tables", err.Error())
		return
	}

	data.ID = types.StringValue(d.Meta().Region(ctx))
	data.TableIDs = fwflex.FlattenFrameworkStringValueList(ctx, out)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findTables(ctx context.Context, conn *dynamodb.Client, input *dynamodb.ListTablesInput) ([]string, error) {
	var output []string

	pages := dynamodb.NewListTablesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.TableNames {
			output = append(output, v)
		}
	}

	return output, nil
}

type dataSourceTablesModel struct {
	ID       types.String `tfsdk:"id"`
	TableIDs types.List   `tfsdk:"ids"`
}
