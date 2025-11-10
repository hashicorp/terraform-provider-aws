// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationsignals

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationsignals"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationsignals/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_applicationsignals_service_level_objective", name="Service Level Objective")
func newDataSourceServiceLevelObjective(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServiceLevelObjective{}, nil
}

const (
	DSNameServiceLevelObjective = "Service Level Objective Data Source"
)

type dataSourceServiceLevelObjective struct {
	framework.DataSourceWithModel[dataSourceServiceLevelObjectiveModel]
}

func (d *dataSourceServiceLevelObjective) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_time": schema.StringAttribute{
				Computed: true,
			},
			"last_updated_time": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"metric_source_type": schema.StringAttribute{
				Computed: true,
			},
			"evaluation_type": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"goal": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[goalModel](ctx),
				Attributes: map[string]schema.Attribute{
					"attainment_goal":   schema.Float64Attribute{Computed: true},
					"warning_threshold": schema.Float64Attribute{Computed: true},
				},
				Blocks: map[string]schema.Block{
					"interval": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[intervalModel](ctx),
						Blocks: map[string]schema.Block{
							"calendar_interval": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[calendarIntervalModel](ctx),
								Attributes: map[string]schema.Attribute{
									"duration":      schema.Int32Attribute{Computed: true},
									"duration_unit": schema.StringAttribute{Computed: true},
									"start_time":    schema.StringAttribute{Computed: true},
								},
							},
							"rolling_interval": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[rollingIntervalModel](ctx),
								Attributes: map[string]schema.Attribute{
									"duration":      schema.Int32Attribute{Computed: true},
									"duration_unit": schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
			"burn_rate_configurations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[burnRateConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"look_back_window_minutes": schema.Int32Attribute{Computed: true},
					},
				},
			},
			"request_based_sli": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[requestBasedSliModel](ctx),
				Attributes: map[string]schema.Attribute{
					"metric_threshold":    schema.Float64Attribute{Computed: true},
					"comparison_operator": schema.StringAttribute{Computed: true},
				},
				Blocks: map[string]schema.Block{
					"request_based_sli_metric": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[requestBasedSliMetricModel](ctx),
						Attributes: map[string]schema.Attribute{
							"metric_type":    schema.StringAttribute{Computed: true},
							"operation_name": schema.StringAttribute{Computed: true},
						},
					},
				},
			},
			"sli": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[sliModel](ctx),
				Attributes: map[string]schema.Attribute{
					"metric_threshold":    schema.Float64Attribute{Computed: true},
					"comparison_operator": schema.StringAttribute{Computed: true},
				},
			},
		},
	}
}

func (d *dataSourceServiceLevelObjective) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ApplicationSignalsClient(ctx)

	var data dataSourceServiceLevelObjectiveModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findServiceLevelObjectiveByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data), smerr.ID, data.ID.String())
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.ID.String())
}

func findServiceLevelObjectiveByID(ctx context.Context, conn *applicationsignals.Client, id string) (*awstypes.ServiceLevelObjective, error) {
	input := &applicationsignals.GetServiceLevelObjectiveInput{
		Id: aws.String(id),
	}

	output, err := conn.GetServiceLevelObjective(ctx, input)

	if err != nil {
		return nil, err
	}

	return output.Slo, nil
}

type dataSourceServiceLevelObjectiveModel struct {
	framework.WithRegionModel
	ID                     types.String                                                `tfsdk:"id"`
	ARN                    types.String                                                `tfsdk:"arn"`
	CreatedTime            types.String                                                `tfsdk:"created_time"`
	BurnRateConfigurations fwtypes.ListNestedObjectValueOf[burnRateConfigurationModel] `tfsdk:"burn_rate_configurations"`
	LastUpdatedTime        types.String                                                `tfsdk:"last_updated_time"`
	Name                   types.String                                                `tfsdk:"name"`
	Description            types.String                                                `tfsdk:"description"`
	MetricSourceType       types.String                                                `tfsdk:"metric_source_type"`
	EvaluationType         types.String                                                `tfsdk:"evaluation_type"`
	Goal                   fwtypes.ObjectValueOf[goalModel]                            `tfsdk:"goal"`
	Sli                    fwtypes.ObjectValueOf[sliModel]                             `tfsdk:"sli"`
	RequestBasedSli        fwtypes.ObjectValueOf[requestBasedSliModel]                 `tfsdk:"request_based_sli"`
}

type goalModel struct {
	AttainmentGoal   types.Float64                        `tfsdk:"attainment_goal"`
	WarningThreshold types.Float64                        `tfsdk:"warning_threshold"`
	Interval         fwtypes.ObjectValueOf[intervalModel] `tfsdk:"interval"`
}

type intervalModel struct {
	CalendarInterval fwtypes.ObjectValueOf[calendarIntervalModel] `tfsdk:"calendar_interval"`
	RollingInterval  fwtypes.ObjectValueOf[rollingIntervalModel]  `tfsdk:"rolling_interval"`
}

type calendarIntervalModel struct {
	Duration     types.Int32  `tfsdk:"duration"`
	DurationUnit types.String `tfsdk:"duration_unit"`
	StartTime    types.String `tfsdk:"start_time"`
}

type rollingIntervalModel struct {
	Duration     types.Int32  `tfsdk:"duration"`
	DurationUnit types.String `tfsdk:"duration_unit"`
}

type sliModel struct {
	MetricThreshold    types.Float64 `tfsdk:"metric_threshold"`
	ComparisonOperator types.String  `tfsdk:"comparison_operator"`
}

type requestBasedSliModel struct {
	RequestBasedSliMetric fwtypes.ObjectValueOf[requestBasedSliMetricModel] `tfsdk:"request_based_sli_metric"`
	MetricThreshold       types.Float64                                     `tfsdk:"metric_threshold"`
	ComparisonOperator    types.String                                      `tfsdk:"comparison_operator"`
}

type burnRateConfigurationModel struct {
	LookBackWindowMinutes types.Int32 `tfsdk:"look_back_window_minutes"`
}

type requestBasedSliMetricModel struct {
	MetricType    types.String `tfsdk:"metric_type"`
	OperationName types.String `tfsdk:"operation_name"`
}
