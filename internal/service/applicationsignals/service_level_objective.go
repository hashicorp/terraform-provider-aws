// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationsignals

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationsignals"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationsignals/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_applicationsignals_service_level_objective", name="Service Level Objective")
func newResourceServiceLevelObjective(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceServiceLevelObjective{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameServiceLevelObjective = "Service Level Objective"
)

type resourceServiceLevelObjective struct {
	framework.ResourceWithModel[resourceServiceLevelObjectiveModel]
	framework.WithTimeouts
}

func (r *resourceServiceLevelObjective) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"evaluation_type": schema.StringAttribute{
				Computed: true,
			},
			"last_updated_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"metric_source_type": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"burn_rate_configurations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[burnRateConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"look_back_window_minutes": schema.Int32Attribute{Optional: true},
					},
				},
			},
			"goal": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[goalModel](ctx),
				Attributes: map[string]schema.Attribute{
					"attainment_goal":   schema.Float64Attribute{Required: true},
					"warning_threshold": schema.Float64Attribute{Required: true},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Blocks: map[string]schema.Block{
					"interval": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[intervalModel](ctx),
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						Blocks: map[string]schema.Block{
							"calendar_interval": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[calendarIntervalModel](ctx),
								Validators: []validator.Object{
									objectvalidator.ExactlyOneOf(
										path.MatchRelative().AtParent().AtName("rolling_interval"),
									),
								},
								Attributes: map[string]schema.Attribute{
									"duration":      schema.Int32Attribute{Optional: true},
									"duration_unit": schema.StringAttribute{Optional: true},
									"start_time": schema.StringAttribute{
										CustomType: timetypes.RFC3339Type{},
										Optional:   true},
								},
							},
							"rolling_interval": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[rollingIntervalModel](ctx),
								Validators: []validator.Object{
									objectvalidator.ExactlyOneOf(
										path.MatchRelative().AtParent().AtName("calendar_interval"),
									),
								},
								Attributes: map[string]schema.Attribute{
									"duration":      schema.Int32Attribute{Optional: true},
									"duration_unit": schema.StringAttribute{Optional: true},
								},
							},
						},
					},
				},
			},
			"request_based_sli": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[requestBasedSliModel](ctx),
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("sli"),
					),
				},
				Attributes: map[string]schema.Attribute{
					"comparison_operator": schema.StringAttribute{Optional: true},
					"metric_threshold":    schema.Float64Attribute{Optional: true},
				},
				Blocks: map[string]schema.Block{
					"request_based_sli_metric": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[requestBasedSliMetricModel](ctx),
						Attributes: map[string]schema.Attribute{
							"key_attributes": schema.MapAttribute{
								CustomType:  fwtypes.MapOfStringType,
								ElementType: types.StringType,
								Optional:    true,
							},
							"metric_type":    schema.StringAttribute{Optional: true},
							"operation_name": schema.StringAttribute{Optional: true},
						},
						Blocks: map[string]schema.Block{
							"dependency_config": dependencyConfigBlock(ctx),
							"monitored_request_count_metric": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[monitoredRequestCountMetricModel](ctx),
								Blocks: map[string]schema.Block{
									"good_count_metric": metricDataQueriesBlock(ctx),
									"bad_count_metric":  metricDataQueriesBlock(ctx),
								},
							},
							"total_request_count_metric": metricDataQueriesBlock(ctx),
						},
					},
				},
			},
			"sli": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[sliModel](ctx),
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("request_based_sli"),
					),
				},
				Attributes: map[string]schema.Attribute{
					"metric_threshold":    schema.Float64Attribute{Optional: true},
					"comparison_operator": schema.StringAttribute{Optional: true},
				},
				Blocks: map[string]schema.Block{
					"sli_metric": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[sliMetricModel](ctx),
						Attributes: map[string]schema.Attribute{
							"key_attributes": schema.MapAttribute{
								CustomType:  fwtypes.MapOfStringType,
								ElementType: types.StringType,
								Optional:    true,
							},
							"metric_type":    schema.StringAttribute{Optional: true},
							"metric_name":    schema.StringAttribute{Optional: true},
							"operation_name": schema.StringAttribute{Optional: true},
							"period_seconds": schema.Int32Attribute{Optional: true},
							"statistic":      schema.StringAttribute{Optional: true},
						},
						Blocks: map[string]schema.Block{
							"dependency_config":   dependencyConfigBlock(ctx),
							"metric_data_queries": metricDataQueriesBlock(ctx),
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func dependencyConfigBlock(ctx context.Context) schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		CustomType: fwtypes.NewObjectTypeOf[dependencyConfigModel](ctx),
		Attributes: map[string]schema.Attribute{
			"dependency_key_attributes": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			"dependency_operation_name": schema.StringAttribute{Optional: true},
		},
	}
}

func metricDataQueriesBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[metricDataQueryModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"account_id":  schema.StringAttribute{Optional: true},
				"expression":  schema.StringAttribute{Optional: true},
				"id":          schema.StringAttribute{Optional: true},
				"label":       schema.StringAttribute{Optional: true},
				"period":      schema.Int32Attribute{Optional: true},
				"return_data": schema.BoolAttribute{Optional: true},
			},
			Blocks: map[string]schema.Block{
				"metric_stat": schema.SingleNestedBlock{
					CustomType: fwtypes.NewObjectTypeOf[metricStatModel](ctx),
					Attributes: map[string]schema.Attribute{
						"period": schema.Int32Attribute{Optional: true},
						"stat":   schema.StringAttribute{Optional: true},
						"unit":   schema.StringAttribute{Optional: true},
					},
					Blocks: map[string]schema.Block{
						"metric": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[metricModel](ctx),
							Attributes: map[string]schema.Attribute{
								"metric_name": schema.StringAttribute{Optional: true},
								"namespace":   schema.StringAttribute{Optional: true},
							},
							Blocks: map[string]schema.Block{
								"dimensions": schema.ListNestedBlock{
									CustomType: fwtypes.NewListNestedObjectTypeOf[dimensionModel](ctx),
									NestedObject: schema.NestedBlockObject{
										CustomType: fwtypes.NewObjectTypeOf[dimensionModel](ctx),
										Attributes: map[string]schema.Attribute{
											"name":  schema.StringAttribute{Optional: true},
											"value": schema.StringAttribute{Optional: true},
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

func (r *resourceServiceLevelObjective) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ApplicationSignalsClient(ctx)

	var plan resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input applicationsignals.CreateServiceLevelObjectiveInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateServiceLevelObjective(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.Slo == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.Slo, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceServiceLevelObjective) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ApplicationSignalsClient(ctx)

	var state resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findServiceLevelObjectiveByID(ctx, conn, state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceServiceLevelObjective) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ApplicationSignalsClient(ctx)

	var plan, state resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input applicationsignals.UpdateServiceLevelObjectiveInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateServiceLevelObjective(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil || out.Slo == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.Slo, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceServiceLevelObjective) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ApplicationSignalsClient(ctx)

	var state resourceServiceLevelObjectiveModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := applicationsignals.DeleteServiceLevelObjectiveInput{
		Id: state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteServiceLevelObjective(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *resourceServiceLevelObjective) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), req, resp)
}

func findServiceLevelObjectiveByID(ctx context.Context, conn *applicationsignals.Client, name string) (*awstypes.ServiceLevelObjective, error) {
	input := applicationsignals.GetServiceLevelObjectiveInput{
		Id: aws.String(name),
	}

	out, err := conn.GetServiceLevelObjective(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Slo == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out.Slo, nil
}

func stringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	val := v.ValueString()
	return &val
}

func flattenStringPtr(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func flattenTimePtr(t *time.Time) timetypes.RFC3339 {
	if t == nil {
		return timetypes.NewRFC3339Null()
	}
	return timetypes.NewRFC3339ValueMust(t.Format(time.RFC3339))
}

func expandBurnRateConfigurations(ctx context.Context, v fwtypes.ListNestedObjectValueOf[burnRateConfigurationModel], diags *diag.Diagnostics) []awstypes.BurnRateConfiguration {
	if v.IsNull() {
		return nil
	}
	var models []burnRateConfigurationModel
	diags.Append(v.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil
	}

	burns := make([]awstypes.BurnRateConfiguration, len(models))
	for i, c := range models {
		burns[i] = awstypes.BurnRateConfiguration{
			LookBackWindowMinutes: c.LookBackWindowMinutes.ValueInt32Pointer(),
		}
	}

	return burns
}

func expandGoal(ctx context.Context, v fwtypes.ObjectValueOf[goalModel], diags *diag.Diagnostics) *awstypes.Goal {
	if v.IsNull() {
		return nil
	}
	goalData, d := v.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	var goal awstypes.Goal
	diags.Append(flex.Expand(ctx, goalData, &goal)...)
	if diags.HasError() {
		return nil
	}
	return &goal
}

func expandSli(ctx context.Context, v fwtypes.ObjectValueOf[sliModel], diags *diag.Diagnostics) *awstypes.ServiceLevelIndicatorConfig {
	if v.IsNull() {
		return nil
	}
	sliData, d := v.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	var sli awstypes.ServiceLevelIndicatorConfig
	diags.Append(flex.Expand(ctx, sliData, &sli)...)
	if diags.HasError() {
		return nil
	}
	return &sli
}

func expandRequestBasedSli(ctx context.Context, v fwtypes.ObjectValueOf[requestBasedSliModel], diags *diag.Diagnostics) *awstypes.RequestBasedServiceLevelIndicatorConfig {
	if v.IsNull() {
		return nil
	}
	reqSliData, d := v.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}
	var reqSli awstypes.RequestBasedServiceLevelIndicatorConfig
	diags.Append(flex.Expand(ctx, reqSliData, &reqSli)...)
	if diags.HasError() {
		return nil
	}
	return &reqSli
}

var (
	_ flex.Expander  = intervalModel{}
	_ flex.Flattener = &intervalModel{}

	_ flex.Expander  = monitoredRequestCountMetricModel{}
	_ flex.Flattener = &monitoredRequestCountMetricModel{}

	_ flex.Expander  = requestBasedSliModel{}
	_ flex.Flattener = &requestBasedSliModel{}

	_ flex.TypedExpander = resourceServiceLevelObjectiveModel{}
	_ flex.Flattener     = &resourceServiceLevelObjectiveModel{}

	_ flex.Expander = sliModel{}
)

func (m intervalModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.RollingInterval.IsNull():
		rollingData, d := m.RollingInterval.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.IntervalMemberRollingInterval
		diags.Append(flex.Expand(ctx, rollingData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.CalendarInterval.IsNull():
		calendarData, d := m.CalendarInterval.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.IntervalMemberCalendarInterval
		diags.Append(flex.Expand(ctx, calendarData, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

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

func (m monitoredRequestCountMetricModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.GoodCountMetric.IsNull() && !m.GoodCountMetric.IsUnknown():
		var r awstypes.MonitoredRequestCountMetricDataQueriesMemberGoodCountMetric
		diags.Append(flex.Expand(ctx, m.GoodCountMetric, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.BadCountMetric.IsNull() && !m.BadCountMetric.IsUnknown():
		var r awstypes.MonitoredRequestCountMetricDataQueriesMemberBadCountMetric
		diags.Append(flex.Expand(ctx, m.BadCountMetric, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}
	return nil, diags
}

func (m *monitoredRequestCountMetricModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.GoodCountMetric = fwtypes.NewListNestedObjectValueOfNull[metricDataQueryModel](ctx)
	m.BadCountMetric = fwtypes.NewListNestedObjectValueOfNull[metricDataQueryModel](ctx)

	switch t := v.(type) {
	case awstypes.MonitoredRequestCountMetricDataQueriesMemberGoodCountMetric:

		models := make([]metricDataQueryModel, 0, len(t.Value))
		for _, apiValue := range t.Value {
			var model metricDataQueryModel
			diags.Append(flex.Flatten(ctx, apiValue, &model)...)
			if diags.HasError() {
				return diags
			}
			models = append(models, model)
		}

		listValue, listDiags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, models)
		diags.Append(listDiags...)

		m.GoodCountMetric = listValue

	case awstypes.MonitoredRequestCountMetricDataQueriesMemberBadCountMetric:

		models := make([]metricDataQueryModel, 0, len(t.Value))
		for _, apiValue := range t.Value {
			var model metricDataQueryModel
			diags.Append(flex.Flatten(ctx, apiValue, &model)...)
			if diags.HasError() {
				return diags
			}
			models = append(models, model)
		}

		listValue, listDiags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, models)
		diags.Append(listDiags...)

		m.BadCountMetric = listValue
	}

	return diags
}

func (m requestBasedSliModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var config awstypes.RequestBasedServiceLevelIndicatorConfig

	if !m.ComparisonOperator.IsNull() {
		config.ComparisonOperator = awstypes.ServiceLevelIndicatorComparisonOperator(m.ComparisonOperator.ValueString())
	}

	if !m.MetricThreshold.IsNull() {
		val := m.MetricThreshold.ValueFloat64()
		config.MetricThreshold = &val
	}

	if !m.RequestBasedSliMetric.IsNull() {
		sliMetricData, d := m.RequestBasedSliMetric.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var metric awstypes.RequestBasedServiceLevelIndicatorMetricConfig
		diags.Append(flex.Expand(ctx, sliMetricData, &metric)...)
		if diags.HasError() {
			return nil, diags
		}

		config.RequestBasedSliMetricConfig = &metric
	}

	return &config, diags
}

func (m *requestBasedSliModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	apiModel, ok := v.(awstypes.RequestBasedServiceLevelIndicator)
	if !ok {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("Flatten Error", "Invalid type passed to Flatten for requestBasedSliModel"),
		}
	}

	if apiModel.ComparisonOperator == "" {
		m.ComparisonOperator = types.StringNull()
	} else {
		m.ComparisonOperator = types.StringValue(string(apiModel.ComparisonOperator))
	}

	if apiModel.MetricThreshold == nil {
		m.MetricThreshold = types.Float64Null()
	} else {
		m.MetricThreshold = types.Float64Value(*apiModel.MetricThreshold)
	}

	if apiModel.RequestBasedSliMetric == nil {
		m.RequestBasedSliMetric = fwtypes.NewObjectValueOfNull[requestBasedSliMetricModel](ctx)
	} else {
		var nestedModel requestBasedSliMetricModel
		innerDiags := flex.Flatten(ctx, apiModel.RequestBasedSliMetric, &nestedModel)
		diags.Append(innerDiags...)
		if !innerDiags.HasError() {
			m.RequestBasedSliMetric = fwtypes.NewObjectValueOfMust(ctx, &nestedModel)
		}
	}

	return diags
}

func (m resourceServiceLevelObjectiveModel) ExpandTo(ctx context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[applicationsignals.UpdateServiceLevelObjectiveInput]():
		return m.expandToUpdateServiceLevelObjectiveInput(ctx)

	case reflect.TypeFor[applicationsignals.CreateServiceLevelObjectiveInput]():
		return m.expandToCreateServiceLevelObjectiveInput(ctx)
	}
	return nil, diags
}

func (m resourceServiceLevelObjectiveModel) expandToUpdateServiceLevelObjectiveInput(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var input applicationsignals.UpdateServiceLevelObjectiveInput

	input.Id = stringPtr(m.Name)
	input.Description = stringPtr(m.Description)
	input.BurnRateConfigurations = expandBurnRateConfigurations(ctx, m.BurnRateConfigurations, &diags)
	input.Goal = expandGoal(ctx, m.Goal, &diags)
	input.SliConfig = expandSli(ctx, m.Sli, &diags)
	input.RequestBasedSliConfig = expandRequestBasedSli(ctx, m.RequestBasedSli, &diags)

	return &input, diags
}

func (m resourceServiceLevelObjectiveModel) expandToCreateServiceLevelObjectiveInput(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var input applicationsignals.CreateServiceLevelObjectiveInput

	input.Name = stringPtr(m.Name)
	input.Description = stringPtr(m.Description)
	input.BurnRateConfigurations = expandBurnRateConfigurations(ctx, m.BurnRateConfigurations, &diags)
	input.Goal = expandGoal(ctx, m.Goal, &diags)
	input.SliConfig = expandSli(ctx, m.Sli, &diags)
	input.RequestBasedSliConfig = expandRequestBasedSli(ctx, m.RequestBasedSli, &diags)

	return &input, diags
}

func (m *resourceServiceLevelObjectiveModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	var apiModel *awstypes.ServiceLevelObjective

	if ptr, ok := v.(*awstypes.ServiceLevelObjective); ok {
		apiModel = ptr
	} else if val, ok := v.(awstypes.ServiceLevelObjective); ok {
		apiModel = &val
	} else {
		diags.AddError("Flatten Error", fmt.Sprintf("Invalid type: expected *ServiceLevelObjective or ServiceLevelObjective, got %T", v))
		return diags
	}

	m.ARN = flattenStringPtr(apiModel.Arn)
	m.Description = flattenStringPtr(apiModel.Description)
	m.Name = flattenStringPtr(apiModel.Name)

	m.CreatedTime = flattenTimePtr(apiModel.CreatedTime)
	m.LastUpdatedTime = flattenTimePtr(apiModel.LastUpdatedTime)

	if apiModel.EvaluationType != "" {
		m.EvaluationType = types.StringValue(string(apiModel.EvaluationType))
	} else {
		m.EvaluationType = types.StringNull()
	}
	if apiModel.MetricSourceType != "" {
		m.MetricSourceType = types.StringValue(string(apiModel.MetricSourceType))
	} else {
		m.MetricSourceType = types.StringNull()
	}

	if apiModel.BurnRateConfigurations != nil {

		models := make([]burnRateConfigurationModel, 0, len(apiModel.BurnRateConfigurations))

		for _, apiValue := range apiModel.BurnRateConfigurations {
			var model burnRateConfigurationModel
			diags.Append(flex.Flatten(ctx, apiValue, &model)...)
			if diags.HasError() {
				return diags
			}
			models = append(models, model)
		}

		listValue, listDiags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, models)
		diags.Append(listDiags...)

		m.BurnRateConfigurations = listValue
	} else {
		m.BurnRateConfigurations = fwtypes.NewListNestedObjectValueOfNull[burnRateConfigurationModel](ctx)
	}

	if apiModel.Goal != nil {
		var goalModel goalModel
		diags.Append(flex.Flatten(ctx, *apiModel.Goal, &goalModel)...)
		m.Goal = fwtypes.NewObjectValueOfMust(ctx, &goalModel)
	}

	if apiModel.Sli != nil {
		var sliModel sliModel
		diags.Append(flex.Flatten(ctx, *apiModel.Sli, &sliModel)...)
		if !diags.HasError() {
			m.Sli = fwtypes.NewObjectValueOfMust(ctx, &sliModel)
		}
	} else {
		m.Sli = fwtypes.NewObjectValueOfNull[sliModel](ctx)
	}

	if apiModel.RequestBasedSli != nil {
		var reqSliModel requestBasedSliModel
		diags.Append(flex.Flatten(ctx, *apiModel.RequestBasedSli, &reqSliModel)...)
		if !diags.HasError() {
			m.RequestBasedSli = fwtypes.NewObjectValueOfMust(ctx, &reqSliModel)
		}
	} else {
		m.RequestBasedSli = fwtypes.NewObjectValueOfNull[requestBasedSliModel](ctx)
	}

	return diags
}

func (m sliModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var config awstypes.ServiceLevelIndicatorConfig

	if !m.ComparisonOperator.IsNull() {
		config.ComparisonOperator = awstypes.ServiceLevelIndicatorComparisonOperator(m.ComparisonOperator.ValueString())
	}

	if !m.MetricThreshold.IsNull() {
		val := m.MetricThreshold.ValueFloat64()
		config.MetricThreshold = &val
	}

	if !m.SliMetric.IsNull() {
		sliMetricData, d := m.SliMetric.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var metric awstypes.ServiceLevelIndicatorMetricConfig
		diags.Append(flex.Expand(ctx, sliMetricData, &metric)...)
		if diags.HasError() {
			return nil, diags
		}

		config.SliMetricConfig = &metric
	}

	return &config, diags
}

type resourceServiceLevelObjectiveModel struct {
	framework.WithRegionModel
	ARN                    types.String                                                `tfsdk:"arn"`
	BurnRateConfigurations fwtypes.ListNestedObjectValueOf[burnRateConfigurationModel] `tfsdk:"burn_rate_configurations"`
	CreatedTime            timetypes.RFC3339                                           `tfsdk:"created_time"`
	Description            types.String                                                `tfsdk:"description"`
	EvaluationType         types.String                                                `tfsdk:"evaluation_type"`
	Goal                   fwtypes.ObjectValueOf[goalModel]                            `tfsdk:"goal"`
	LastUpdatedTime        timetypes.RFC3339                                           `tfsdk:"last_updated_time"`
	MetricSourceType       types.String                                                `tfsdk:"metric_source_type"`
	Name                   types.String                                                `tfsdk:"name"`
	RequestBasedSli        fwtypes.ObjectValueOf[requestBasedSliModel]                 `tfsdk:"request_based_sli"`
	Sli                    fwtypes.ObjectValueOf[sliModel]                             `tfsdk:"sli"`
	Timeouts               timeouts.Value                                              `tfsdk:"timeouts"`
}

type burnRateConfigurationModel struct {
	LookBackWindowMinutes types.Int32 `tfsdk:"look_back_window_minutes"`
}

type goalModel struct {
	AttainmentGoal   types.Float64                        `tfsdk:"attainment_goal"`
	Interval         fwtypes.ObjectValueOf[intervalModel] `tfsdk:"interval"`
	WarningThreshold types.Float64                        `tfsdk:"warning_threshold"`
}

type intervalModel struct {
	CalendarInterval fwtypes.ObjectValueOf[calendarIntervalModel] `tfsdk:"calendar_interval"`
	RollingInterval  fwtypes.ObjectValueOf[rollingIntervalModel]  `tfsdk:"rolling_interval"`
}

type calendarIntervalModel struct {
	Duration     types.Int32       `tfsdk:"duration"`
	DurationUnit types.String      `tfsdk:"duration_unit"`
	StartTime    timetypes.RFC3339 `tfsdk:"start_time"`
}

type rollingIntervalModel struct {
	Duration     types.Int32  `tfsdk:"duration"`
	DurationUnit types.String `tfsdk:"duration_unit"`
}

type requestBasedSliModel struct {
	ComparisonOperator    types.String                                      `tfsdk:"comparison_operator"`
	MetricThreshold       types.Float64                                     `tfsdk:"metric_threshold"`
	RequestBasedSliMetric fwtypes.ObjectValueOf[requestBasedSliMetricModel] `tfsdk:"request_based_sli_metric"`
}

type requestBasedSliMetricModel struct {
	DependencyConfig            fwtypes.ObjectValueOf[dependencyConfigModel]            `tfsdk:"dependency_config"`
	KeyAttributes               fwtypes.MapOfString                                     `tfsdk:"key_attributes"`
	MetricType                  types.String                                            `tfsdk:"metric_type" autoflex:",omitempty"`
	MonitoredRequestCountMetric fwtypes.ObjectValueOf[monitoredRequestCountMetricModel] `tfsdk:"monitored_request_count_metric"`
	OperationName               types.String                                            `tfsdk:"operation_name"`
	TotalRequestCountMetric     fwtypes.ListNestedObjectValueOf[metricDataQueryModel]   `tfsdk:"total_request_count_metric"`
}

type monitoredRequestCountMetricModel struct {
	GoodCountMetric fwtypes.ListNestedObjectValueOf[metricDataQueryModel] `tfsdk:"good_count_metric"`
	BadCountMetric  fwtypes.ListNestedObjectValueOf[metricDataQueryModel] `tfsdk:"bad_count_metric"`
}

type sliModel struct {
	ComparisonOperator types.String                          `tfsdk:"comparison_operator"`
	MetricThreshold    types.Float64                         `tfsdk:"metric_threshold"`
	SliMetric          fwtypes.ObjectValueOf[sliMetricModel] `tfsdk:"sli_metric"`
}

type sliMetricModel struct {
	DependencyConfig  fwtypes.ObjectValueOf[dependencyConfigModel]          `tfsdk:"dependency_config"`
	KeyAttributes     fwtypes.MapOfString                                   `tfsdk:"key_attributes"`
	MetricDataQueries fwtypes.ListNestedObjectValueOf[metricDataQueryModel] `tfsdk:"metric_data_queries"`
	MetricName        types.String                                          `tfsdk:"metric_name"`
	MetricType        types.String                                          `tfsdk:"metric_type" autoflex:",omitempty"`
	OperationName     types.String                                          `tfsdk:"operation_name"`
	PeriodSeconds     types.Int32                                           `tfsdk:"period_seconds"`
	Statistic         types.String                                          `tfsdk:"statistic"`
}

type dependencyConfigModel struct {
	DependencyKeyAttributes fwtypes.MapOfString `tfsdk:"dependency_key_attributes"`
	DependencyOperationName types.String        `tfsdk:"dependency_operation_name"`
}

type metricDataQueryModel struct {
	AccountId  types.String                           `tfsdk:"account_id"`
	Expression types.String                           `tfsdk:"expression"`
	Id         types.String                           `tfsdk:"id"`
	Label      types.String                           `tfsdk:"label"`
	MetricStat fwtypes.ObjectValueOf[metricStatModel] `tfsdk:"metric_stat"`
	Period     types.Int32                            `tfsdk:"period"`
	ReturnData types.Bool                             `tfsdk:"return_data"`
}

type metricStatModel struct {
	Metric fwtypes.ObjectValueOf[metricModel] `tfsdk:"metric"`
	Period types.Int32                        `tfsdk:"period"`
	Stat   types.String                       `tfsdk:"stat"`
	Unit   types.String                       `tfsdk:"unit" autoflex:",omitempty"`
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

func sweepServiceLevelObjectives(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := applicationsignals.ListServiceLevelObjectivesInput{}
	conn := client.ApplicationSignalsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := applicationsignals.NewListServiceLevelObjectivesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.SloSummaries {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceServiceLevelObjective, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.Name))),
			)
		}
	}

	return sweepResources, nil
}
