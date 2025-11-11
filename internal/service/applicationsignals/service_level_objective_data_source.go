// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationsignals

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationsignals"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationsignals/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
			names.AttrCreatedTime: schema.StringAttribute{
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
							"dependency_config": schema.StringAttribute{Computed: true},
							"key_attributes":    schema.MapAttribute{CustomType: fwtypes.MapOfStringType, ElementType: types.StringType, Computed: true},
							"metric_type":       schema.StringAttribute{Computed: true},
							"operation_name":    schema.StringAttribute{Computed: true},
						},
						Blocks: map[string]schema.Block{
							"total_request_count_metric": metricDataQueriesBlock(ctx),
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
				Blocks: map[string]schema.Block{
					"sli_metric": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[sliMetricModel](ctx),
						Attributes: map[string]schema.Attribute{
							"dependency_config": schema.StringAttribute{Computed: true},
							"key_attributes":    schema.MapAttribute{CustomType: fwtypes.MapOfStringType, ElementType: types.StringType, Computed: true},
							"metric_type":       schema.StringAttribute{Computed: true},
							"operation_name":    schema.StringAttribute{Computed: true},
						},
						Blocks: map[string]schema.Block{
							"metric_data_queries": metricDataQueriesBlock(ctx),
						},
					},
				},
			},
		},
	}
}

func metricDataQueriesBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[metricDataQueryModel](ctx),
		NestedObject: schema.NestedBlockObject{
			CustomType: fwtypes.NewObjectTypeOf[metricDataQueryModel](ctx),
			Attributes: map[string]schema.Attribute{
				"id":          schema.StringAttribute{Computed: true},
				"account_id":  schema.StringAttribute{Computed: true},
				"expression":  schema.StringAttribute{Computed: true},
				"label":       schema.StringAttribute{Computed: true},
				"period":      schema.Int32Attribute{Computed: true},
				"return_data": schema.BoolAttribute{Computed: true},
			},
			Blocks: map[string]schema.Block{
				"metric_stat": schema.SingleNestedBlock{
					CustomType: fwtypes.NewObjectTypeOf[metricStatModel](ctx),
					Attributes: map[string]schema.Attribute{
						"period": schema.Int32Attribute{Computed: true},
						"stat":   schema.StringAttribute{Computed: true},
						"unit":   schema.StringAttribute{Computed: true},
					},
					Blocks: map[string]schema.Block{
						"metric": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[metricModel](ctx),
							Attributes: map[string]schema.Attribute{
								"metric_name": schema.StringAttribute{Computed: true},
								"namespace":   schema.StringAttribute{Computed: true},
							},
							Blocks: map[string]schema.Block{
								"dimensions": schema.ListNestedBlock{
									CustomType: fwtypes.NewListNestedObjectTypeOf[dimensionModel](ctx),
									NestedObject: schema.NestedBlockObject{
										CustomType: fwtypes.NewObjectTypeOf[dimensionModel](ctx),
										Attributes: map[string]schema.Attribute{
											"name":  schema.StringAttribute{Computed: true},
											"value": schema.StringAttribute{Computed: true},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceServiceLevelObjective) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ApplicationSignalsClient(ctx)

	var data dataSourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findServiceLevelObjectiveByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	data.CreatedTime = types.StringValue(aws.ToTime(out.CreatedTime).Format(time.RFC3339))
	data.LastUpdatedTime = types.StringValue(aws.ToTime(out.LastUpdatedTime).Format(time.RFC3339))

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data), smerr.ID, data.ID.String())
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.ID.String())
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

var _ flex.Flattener = &intervalModel{}

func (m *intervalModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.CalendarInterval = fwtypes.NewObjectValueOfNull[calendarIntervalModel](ctx)
	m.RollingInterval = fwtypes.NewObjectValueOfNull[rollingIntervalModel](ctx)

	switch t := v.(type) {

	case awstypes.IntervalMemberCalendarInterval:
		var model calendarIntervalModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		if !diags.HasError() {
			m.CalendarInterval = fwtypes.NewObjectValueOfMust(ctx, &model)
		}

	case awstypes.IntervalMemberRollingInterval:
		var model rollingIntervalModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		if !diags.HasError() {
			m.RollingInterval = fwtypes.NewObjectValueOfMust(ctx, &model)
		}
	}

	return diags
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
	ComparisonOperator types.String                          `tfsdk:"comparison_operator"`
	MetricThreshold    types.Float64                         `tfsdk:"metric_threshold"`
	SliMetric          fwtypes.ObjectValueOf[sliMetricModel] `tfsdk:"sli_metric"`
}

type requestBasedSliModel struct {
	RequestBasedSliMetric fwtypes.ObjectValueOf[requestBasedSliMetricModel] `tfsdk:"request_based_sli_metric"`
	ComparisonOperator    types.String                                      `tfsdk:"comparison_operator"`
	MetricThreshold       types.Float64                                     `tfsdk:"metric_threshold"`
}

type burnRateConfigurationModel struct {
	LookBackWindowMinutes types.Int32 `tfsdk:"look_back_window_minutes"`
}

type requestBasedSliMetricModel struct {
	TotalRequestCountMetric fwtypes.ListNestedObjectValueOf[metricDataQueryModel] `tfsdk:"total_request_count_metric"`
	DependencyConfig        fwtypes.ObjectValueOf[dependencyConfigModel]          `tfsdk:"dependency_config"`
	KeyAttributes           fwtypes.MapOfString                                   `tfsdk:"key_attributes"`
	MetricType              types.String                                          `tfsdk:"metric_type"`
	OperationName           types.String                                          `tfsdk:"operation_name"`
}

type sliMetricModel struct {
	MetricDataQueries fwtypes.ListNestedObjectValueOf[metricDataQueryModel] `tfsdk:"metric_data_queries"`
	DependencyConfig  fwtypes.ObjectValueOf[dependencyConfigModel]          `tfsdk:"dependency_config"`
	KeyAttributes     fwtypes.MapOfString                                   `tfsdk:"key_attributes"`
	MetricType        types.String                                          `tfsdk:"metric_type"`
	OperationName     types.String                                          `tfsdk:"operation_name"`
}

type metricDataQueryModel struct {
	Id         types.String                           `tfsdk:"id"`
	AccountId  types.String                           `tfsdk:"account_id"`
	Expression types.String                           `tfsdk:"expression"`
	Label      types.String                           `tfsdk:"label"`
	MetricStat fwtypes.ObjectValueOf[metricStatModel] `tfsdk:"metric_stat"`
	Period     types.Int32                            `tfsdk:"period"`
	ReturnData types.Bool                             `tfsdk:"return_data"`
}

type metricStatModel struct {
	Metric fwtypes.ObjectValueOf[metricModel] `tfsdk:"metric"`
	Period types.Int32                        `tfsdk:"period"`
	Stat   types.String                       `tfsdk:"stat"`
	Unit   types.String                       `tfsdk:"unit"`
}

type metricModel struct {
	Dimensions fwtypes.ListNestedObjectValueOf[dimensionModel] `tfsdk:"dimensions"`
	MetricName types.String                                    `tfsdk:"metric_name"`
	Namespace  types.String                                    `tfsdk:"namespace"`
}

type dimensionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type dependencyConfigModel struct {
	DependencyKeyAttributes types.String `tfsdk:"dependency_key_attributes"`
	DependencyOperationName types.String `tfsdk:"dependency_operation_name"`
}
