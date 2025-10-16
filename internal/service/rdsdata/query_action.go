// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rdsdata

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	fwtypes2 "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_rdsdata_query, name="RDS Data Query")
func newQueryAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &queryAction{}, nil
}

// NewQueryAction creates a new query action for testing
func NewQueryAction(ctx context.Context) (action.ActionWithConfigure, error) {
	return newQueryAction(ctx)
}

var (
	_ action.Action = (*queryAction)(nil)
)

type queryAction struct {
	framework.ActionWithModel[queryActionModel]
	client *rdsdata.Client
}

type queryActionModel struct {
	framework.WithRegionModel
	ResourceArn           fwtypes2.String                                    `tfsdk:"resource_arn"`
	SecretArn             fwtypes2.String                                    `tfsdk:"secret_arn"`
	SQL                   fwtypes2.String                                    `tfsdk:"sql"`
	Database              fwtypes2.String                                    `tfsdk:"database"`
	Parameters            fwtypes.ListNestedObjectValueOf[sqlParameterModel] `tfsdk:"parameters"`
	IncludeResultMetadata fwtypes2.Bool                                      `tfsdk:"include_result_metadata"`
}

type sqlParameterModel struct {
	Name  fwtypes2.String `tfsdk:"name"`
	Value fwtypes2.String `tfsdk:"value"`
}

func (a *queryAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Executes SQL queries against Aurora Serverless clusters using RDS Data API",
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: schema.StringAttribute{
				Description: "ARN of the Aurora Serverless cluster",
				Required:    true,
			},
			"secret_arn": schema.StringAttribute{
				Description: "ARN of the Secrets Manager secret containing database credentials",
				Required:    true,
			},
			"sql": schema.StringAttribute{
				Description: "SQL statement to execute",
				Required:    true,
			},
			names.AttrDatabase: schema.StringAttribute{
				Description: "Name of the database",
				Optional:    true,
			},
			"include_result_metadata": schema.BoolAttribute{
				Description: "Include column metadata in query results",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrParameters: schema.ListNestedBlock{
				Description: "SQL parameters for prepared statements",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[sqlParameterModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Description: "Parameter name",
							Required:    true,
						},
						names.AttrValue: schema.StringAttribute{
							Description: "Parameter value",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (a *queryAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	meta, ok := req.ProviderData.(*conns.AWSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *conns.AWSClient, got: %T", req.ProviderData),
		)
		return
	}

	a.client = meta.RDSDataClient(ctx)
}

func (a *queryAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var model queryActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Starting RDS Data query execution", map[string]any{
		names.AttrResourceARN: model.ResourceArn.ValueString(),
		"sql_length":          len(model.SQL.ValueString()),
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Executing SQL query...",
	})

	input := rdsdata.ExecuteStatementInput{
		ResourceArn: model.ResourceArn.ValueStringPointer(),
		SecretArn:   model.SecretArn.ValueStringPointer(),
		Sql:         model.SQL.ValueStringPointer(),
	}

	if !model.Database.IsNull() {
		input.Database = model.Database.ValueStringPointer()
	}

	if !model.IncludeResultMetadata.IsNull() {
		input.IncludeResultMetadata = model.IncludeResultMetadata.ValueBool()
	}

	if !model.Parameters.IsNull() {
		var params []sqlParameterModel
		resp.Diagnostics.Append(model.Parameters.ElementsAs(ctx, &params, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var sqlParams []types.SqlParameter
		for _, param := range params {
			sqlParams = append(sqlParams, types.SqlParameter{
				Name:  param.Name.ValueStringPointer(),
				Value: &types.FieldMemberStringValue{Value: param.Value.ValueString()},
			})
		}
		input.Parameters = sqlParams
	}

	output, err := a.client.ExecuteStatement(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("SQL Execution Failed", err.Error())
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Query executed successfully. Records affected: %d", output.NumberOfRecordsUpdated),
	})

	tflog.Info(ctx, "RDS Data query completed successfully", map[string]any{
		"records_updated": output.NumberOfRecordsUpdated,
		"has_records":     len(output.Records) > 0,
	})
}
