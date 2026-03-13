// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	ifwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_rds_cluster_parameter_group",name=Cluster Parameter Group)
func newClusterParameterGroupDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &clusterParameterGroupDataSource{}, nil
}

const (
	DSNameClusterParameterGroup = "Cluster Parameter Group Data Source"

	dbClusterParameterGroupPrefix = "DBClusterParameterGroup"
)

type clusterParameterGroupDataSource struct {
	framework.DataSourceWithModel[clusterParameterGroupDataSourceModel]
}

func (d *clusterParameterGroupDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrFamily: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrParameters: schema.SetNestedBlock{
				CustomType: ifwtypes.NewSetNestedObjectTypeOf[parameterModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"apply_method": schema.StringAttribute{
							Computed: true,
						},
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
						names.AttrValue: schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *clusterParameterGroupDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().RDSClient(ctx)
	var data clusterParameterGroupDataSourceModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findDBClusterParameterGroupByName(ctx, conn, data.Name.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionReading, DSNameClusterParameterGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix(dbClusterParameterGroupPrefix))...)

	if response.Diagnostics.HasError() {
		return
	}

	data.Family = fwflex.StringToFramework(ctx, output.DBParameterGroupFamily)

	// Read parameters
	input := &rds.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: data.Name.ValueStringPointer(),
	}

	parameters, err := findDBClusterParameters(ctx, conn, input, tfslices.PredicateTrue[*types.Parameter]())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionReading, DSNameClusterParameterGroup+" parameters", data.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Convert parameters to parameterModel structs for Plugin Framework
	parameterModels := make([]parameterModel, 0, len(parameters))
	for _, param := range parameters {
		if param.ParameterName == nil {
			continue
		}

		model := parameterModel{
			ApplyMethod: fwtypes.StringValue(string(param.ApplyMethod)),
			Name:        fwtypes.StringValue(aws.ToString(param.ParameterName)),
			Value:       fwtypes.StringValue(""),
		}

		if param.ParameterValue != nil {
			model.Value = fwtypes.StringValue(aws.ToString(param.ParameterValue))
		}

		parameterModels = append(parameterModels, model)
	}

	// Convert to SetNestedObjectValueOf for Plugin Framework
	parametersValue, diags := ifwtypes.NewSetNestedObjectValueOfValueSlice(ctx, parameterModels)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Parameters = parametersValue

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type clusterParameterGroupDataSourceModel struct {
	framework.WithRegionModel
	ARN         fwtypes.String                                  `tfsdk:"arn"`
	Description fwtypes.String                                  `tfsdk:"description"`
	Family      fwtypes.String                                  `tfsdk:"family"`
	Name        fwtypes.String                                  `tfsdk:"name"`
	Parameters  ifwtypes.SetNestedObjectValueOf[parameterModel] `tfsdk:"parameters"`
}

type parameterModel struct {
	ApplyMethod fwtypes.String `tfsdk:"apply_method"`
	Name        fwtypes.String `tfsdk:"name"`
	Value       fwtypes.String `tfsdk:"value"`
}
