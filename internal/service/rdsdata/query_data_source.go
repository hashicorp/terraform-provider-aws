// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rdsdata

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	rdsdatatypes "github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_rdsdata_query", name="Query")
func newDataSourceQuery(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceQuery{}, nil
}

type dataSourceQuery struct {
	framework.DataSourceWithModel[dataSourceQueryModel]
}

func (d *dataSourceQuery) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrDatabase: schema.StringAttribute{
				Optional: true,
			},
			names.AttrResourceARN: schema.StringAttribute{
				Required: true,
			},
			"secret_arn": schema.StringAttribute{
				Required: true,
			},
			"sql": schema.StringAttribute{
				Required: true,
			},
			"records": schema.StringAttribute{
				Computed: true,
			},
			"number_of_records_updated": schema.Int64Attribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrParameters: schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrValue: schema.StringAttribute{
							Required: true,
						},
						"type_hint": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

type dataSourceQueryModel struct {
	framework.WithRegionModel
	ID                     types.String                    `tfsdk:"id"`
	Database               types.String                    `tfsdk:"database"`
	ResourceARN            types.String                    `tfsdk:"resource_arn"`
	SecretARN              types.String                    `tfsdk:"secret_arn"`
	SQL                    types.String                    `tfsdk:"sql"`
	Parameters             []dataSourceQueryParameterModel `tfsdk:"parameters"`
	Records                types.String                    `tfsdk:"records"`
	NumberOfRecordsUpdated types.Int64                     `tfsdk:"number_of_records_updated"`
}

type dataSourceQueryParameterModel struct {
	Name     types.String `tfsdk:"name"`
	Value    types.String `tfsdk:"value"`
	TypeHint types.String `tfsdk:"type_hint"`
}

func (d *dataSourceQuery) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceQueryModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().RDSDataClient(ctx)

	input := rdsdata.ExecuteStatementInput{
		ResourceArn:     data.ResourceARN.ValueStringPointer(),
		SecretArn:       data.SecretARN.ValueStringPointer(),
		Sql:             data.SQL.ValueStringPointer(),
		FormatRecordsAs: rdsdatatypes.RecordsFormatTypeJson,
	}

	if !data.Database.IsNull() {
		input.Database = data.Database.ValueStringPointer()
	}

	if len(data.Parameters) > 0 {
		input.Parameters = expandSQLParameters(data.Parameters)
	}

	output, err := conn.ExecuteStatement(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("executing RDS Data API statement", err.Error())
		return
	}

	data.ID = types.StringValue(data.ResourceARN.ValueString() + ":" + data.SQL.ValueString())
	data.Records = types.StringPointerValue(output.FormattedRecords)
	data.NumberOfRecordsUpdated = types.Int64Value(output.NumberOfRecordsUpdated)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func expandSQLParameters(tfList []dataSourceQueryParameterModel) []rdsdatatypes.SqlParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []rdsdatatypes.SqlParameter

	for _, tfObj := range tfList {
		apiObject := rdsdatatypes.SqlParameter{
			Name: tfObj.Name.ValueStringPointer(),
		}

		if !tfObj.TypeHint.IsNull() {
			apiObject.TypeHint = rdsdatatypes.TypeHint(tfObj.TypeHint.ValueString())
		}

		// Convert value to Field type
		valueStr := tfObj.Value.ValueString()
		var field rdsdatatypes.Field

		// Try to parse as JSON first, otherwise treat as string
		var jsonValue any
		if err := json.Unmarshal([]byte(valueStr), &jsonValue); err == nil {
			switch v := jsonValue.(type) {
			case string:
				field = &rdsdatatypes.FieldMemberStringValue{Value: v}
			case float64:
				field = &rdsdatatypes.FieldMemberDoubleValue{Value: v}
			case bool:
				field = &rdsdatatypes.FieldMemberBooleanValue{Value: v}
			default:
				field = &rdsdatatypes.FieldMemberStringValue{Value: valueStr}
			}
		} else {
			field = &rdsdatatypes.FieldMemberStringValue{Value: valueStr}
		}

		apiObject.Value = field
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
