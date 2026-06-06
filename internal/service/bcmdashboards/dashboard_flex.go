// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdashboards

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/bcmdashboards"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bcmdashboards/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// AutoFlex handles the leaf query/filter/group structures field-for-field. The
// QueryParameters and DisplayConfig SDK unions (Go interfaces) require custom
// expansion and flattening, which is implemented on widgetConfigModel and the
// helpers below.
var (
	_ flex.Expander  = widgetConfigModel{}
	_ flex.Flattener = (*widgetConfigModel)(nil)
)

// flattenDashboard copies a GetDashboardOutput into the resource model. AutoFlex
// handles name, description, timestamps, and the widget tree; the ARN-derived id
// and the renamed dashboard_type attribute are set explicitly.
func flattenDashboard(ctx context.Context, out *bcmdashboards.GetDashboardOutput, model *dashboardResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(flex.Flatten(ctx, out, model)...)
	if diags.HasError() {
		return diags
	}

	model.ARN = flex.StringToFramework(ctx, out.Arn)
	model.ID = model.ARN
	model.DashboardType = fwtypes.StringEnumValue(out.Type)

	return diags
}

func (m widgetConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := &awstypes.WidgetConfig{}

	queryParameters, d := m.QueryParameters.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	if queryParameters != nil {
		v, d := expandQueryParameters(ctx, *queryParameters)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out.QueryParameters = v
	}

	displayConfig, d := m.DisplayConfig.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	if displayConfig != nil {
		v, d := expandDisplayConfig(ctx, *displayConfig)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out.DisplayConfig = v
	}

	return out, diags
}

func (m *widgetConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	var widgetConfig awstypes.WidgetConfig
	switch t := v.(type) {
	case awstypes.WidgetConfig:
		widgetConfig = t
	case *awstypes.WidgetConfig:
		widgetConfig = *t
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("flattening widget config: %T", v))
		return diags
	}

	queryParameters, d := flattenQueryParameters(ctx, widgetConfig.QueryParameters)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.QueryParameters = queryParameters

	displayConfig, d := flattenDisplayConfig(ctx, widgetConfig.DisplayConfig)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.DisplayConfig = displayConfig

	return diags
}

func expandQueryParameters(ctx context.Context, m queryParametersModel) (awstypes.QueryParameters, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.CostAndUsage.IsNull():
		query, d := m.CostAndUsage.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var value awstypes.CostAndUsageQuery
		// WithNoIgnoredFieldNames so the filter expression's "tags" field (a cost
		// filter dimension, not resource tags) is not skipped by AutoFlex.
		diags.Append(flex.Expand(ctx, query, &value, flex.WithNoIgnoredFieldNames())...)
		return &awstypes.QueryParametersMemberCostAndUsage{Value: value}, diags

	case !m.ReservationCoverage.IsNull():
		query, d := m.ReservationCoverage.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var value awstypes.ReservationCoverageQuery
		// WithNoIgnoredFieldNames so the filter expression's "tags" field (a cost
		// filter dimension, not resource tags) is not skipped by AutoFlex.
		diags.Append(flex.Expand(ctx, query, &value, flex.WithNoIgnoredFieldNames())...)
		return &awstypes.QueryParametersMemberReservationCoverage{Value: value}, diags

	case !m.ReservationUtilization.IsNull():
		query, d := m.ReservationUtilization.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var value awstypes.ReservationUtilizationQuery
		// WithNoIgnoredFieldNames so the filter expression's "tags" field (a cost
		// filter dimension, not resource tags) is not skipped by AutoFlex.
		diags.Append(flex.Expand(ctx, query, &value, flex.WithNoIgnoredFieldNames())...)
		return &awstypes.QueryParametersMemberReservationUtilization{Value: value}, diags

	case !m.SavingsPlansCoverage.IsNull():
		query, d := m.SavingsPlansCoverage.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var value awstypes.SavingsPlansCoverageQuery
		// WithNoIgnoredFieldNames so the filter expression's "tags" field (a cost
		// filter dimension, not resource tags) is not skipped by AutoFlex.
		diags.Append(flex.Expand(ctx, query, &value, flex.WithNoIgnoredFieldNames())...)
		return &awstypes.QueryParametersMemberSavingsPlansCoverage{Value: value}, diags

	case !m.SavingsPlansUtilization.IsNull():
		query, d := m.SavingsPlansUtilization.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var value awstypes.SavingsPlansUtilizationQuery
		// WithNoIgnoredFieldNames so the filter expression's "tags" field (a cost
		// filter dimension, not resource tags) is not skipped by AutoFlex.
		diags.Append(flex.Expand(ctx, query, &value, flex.WithNoIgnoredFieldNames())...)
		return &awstypes.QueryParametersMemberSavingsPlansUtilization{Value: value}, diags
	}

	diags.AddError(
		"Invalid query_parameters",
		"exactly one query type (cost_and_usage, reservation_coverage, reservation_utilization, savings_plans_coverage, or savings_plans_utilization) must be specified",
	)
	return nil, diags
}

func flattenQueryParameters(ctx context.Context, queryParameters awstypes.QueryParameters) (fwtypes.ListNestedObjectValueOf[queryParametersModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	// The zero value of a nested object list has no element type, so every union
	// alternative must be explicitly null before the active one is populated.
	m := queryParametersModel{
		CostAndUsage:            fwtypes.NewListNestedObjectValueOfNull[costAndUsageQueryModel](ctx),
		ReservationCoverage:     fwtypes.NewListNestedObjectValueOfNull[reservationCoverageQueryModel](ctx),
		ReservationUtilization:  fwtypes.NewListNestedObjectValueOfNull[reservationUtilizationQueryModel](ctx),
		SavingsPlansCoverage:    fwtypes.NewListNestedObjectValueOfNull[savingsPlansCoverageQueryModel](ctx),
		SavingsPlansUtilization: fwtypes.NewListNestedObjectValueOfNull[savingsPlansUtilizationQueryModel](ctx),
	}

	switch v := queryParameters.(type) {
	case *awstypes.QueryParametersMemberCostAndUsage:
		var query costAndUsageQueryModel
		diags.Append(flex.Flatten(ctx, v.Value, &query, flex.WithNoIgnoredFieldNames())...)
		value, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &query)
		diags.Append(d...)
		m.CostAndUsage = value

	case *awstypes.QueryParametersMemberReservationCoverage:
		var query reservationCoverageQueryModel
		diags.Append(flex.Flatten(ctx, v.Value, &query, flex.WithNoIgnoredFieldNames())...)
		value, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &query)
		diags.Append(d...)
		m.ReservationCoverage = value

	case *awstypes.QueryParametersMemberReservationUtilization:
		var query reservationUtilizationQueryModel
		diags.Append(flex.Flatten(ctx, v.Value, &query, flex.WithNoIgnoredFieldNames())...)
		value, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &query)
		diags.Append(d...)
		m.ReservationUtilization = value

	case *awstypes.QueryParametersMemberSavingsPlansCoverage:
		var query savingsPlansCoverageQueryModel
		diags.Append(flex.Flatten(ctx, v.Value, &query, flex.WithNoIgnoredFieldNames())...)
		value, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &query)
		diags.Append(d...)
		m.SavingsPlansCoverage = value

	case *awstypes.QueryParametersMemberSavingsPlansUtilization:
		var query savingsPlansUtilizationQueryModel
		diags.Append(flex.Flatten(ctx, v.Value, &query, flex.WithNoIgnoredFieldNames())...)
		value, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &query)
		diags.Append(d...)
		m.SavingsPlansUtilization = value
	}

	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfNull[queryParametersModel](ctx), diags
	}

	result, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &m)
	diags.Append(d...)
	return result, diags
}

func expandDisplayConfig(ctx context.Context, m displayConfigModel) (awstypes.DisplayConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.Graph.IsNull():
		graphs, d := m.Graph.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		value := make(map[string]awstypes.GraphDisplayConfig, len(graphs))
		for _, g := range graphs {
			value[g.Metric.ValueString()] = awstypes.GraphDisplayConfig{
				VisualType: g.VisualType.ValueEnum(),
			}
		}
		return &awstypes.DisplayConfigMemberGraph{Value: value}, diags

	case !m.Table.IsNull():
		return &awstypes.DisplayConfigMemberTable{Value: awstypes.TableDisplayConfigStruct{}}, diags
	}

	diags.AddError(
		"Invalid display_config",
		"exactly one of graph or table must be specified",
	)
	return nil, diags
}

func flattenDisplayConfig(ctx context.Context, displayConfig awstypes.DisplayConfig) (fwtypes.ListNestedObjectValueOf[displayConfigModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	// The zero value of a nested object list has no element type, so both union
	// alternatives must be explicitly null before the active one is populated.
	m := displayConfigModel{
		Graph: fwtypes.NewListNestedObjectValueOfNull[graphDisplayConfigModel](ctx),
		Table: fwtypes.NewListNestedObjectValueOfNull[tableDisplayConfigModel](ctx),
	}

	switch v := displayConfig.(type) {
	case *awstypes.DisplayConfigMemberGraph:
		// Map iteration order is non-deterministic; sort the keys so the
		// flattened list is stable across reads.
		keys := make([]string, 0, len(v.Value))
		for k := range v.Value {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		graphs := make([]graphDisplayConfigModel, 0, len(keys))
		for _, k := range keys {
			graphs = append(graphs, graphDisplayConfigModel{
				Metric:     types.StringValue(k),
				VisualType: fwtypes.StringEnumValue(v.Value[k].VisualType),
			})
		}
		value, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, graphs)
		diags.Append(d...)
		m.Graph = value

	case *awstypes.DisplayConfigMemberTable:
		value, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &tableDisplayConfigModel{})
		diags.Append(d...)
		m.Table = value
	}

	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfNull[displayConfigModel](ctx), diags
	}

	result, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &m)
	diags.Append(d...)
	return result, diags
}
