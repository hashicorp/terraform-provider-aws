// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rdsdata

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	rdsdatatypes "github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rdsdata_query", name="Query")
func newResourceQuery(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceQuery{}, nil
}

type resourceQuery struct {
	framework.ResourceWithModel[resourceQueryModel]
}

func (r *resourceQuery) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrDatabase: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrResourceARN: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secret_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sql": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

type resourceQueryModel struct {
	framework.WithRegionModel
	ID                     types.String                  `tfsdk:"id"`
	Database               types.String                  `tfsdk:"database"`
	ResourceARN            types.String                  `tfsdk:"resource_arn"`
	SecretARN              types.String                  `tfsdk:"secret_arn"`
	SQL                    types.String                  `tfsdk:"sql"`
	Parameters             []resourceQueryParameterModel `tfsdk:"parameters"`
	Records                types.String                  `tfsdk:"records"`
	NumberOfRecordsUpdated types.Int64                   `tfsdk:"number_of_records_updated"`
}

type resourceQueryParameterModel struct {
	Name     types.String `tfsdk:"name"`
	Value    types.String `tfsdk:"value"`
	TypeHint types.String `tfsdk:"type_hint"`
}

func (r *resourceQuery) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceQueryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSDataClient(ctx)

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
		// Convert resource parameter model to data source parameter model for compatibility
		var params []dataSourceQueryParameterModel
		for _, p := range data.Parameters {
			params = append(params, dataSourceQueryParameterModel(p))
		}
		input.Parameters = expandSQLParameters(params)
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

func (r *resourceQuery) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// No-op: query results are stored in state and don't need to be refreshed
}

func (r *resourceQuery) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No-op: all changes require replacement
}

func (r *resourceQuery) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: no API call needed, just remove from state
}
