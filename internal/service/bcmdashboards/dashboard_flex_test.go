// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdashboards

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bcmdashboards/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestWidgetConfigRoundTrip exercises the custom Expander/Flattener glue for the
// QueryParameters and DisplayConfig SDK unions and the depth-capped filter
// expression. Each case is flattened from an SDK WidgetConfig into the resource
// model and expanded back, and the result must equal the original.
func TestWidgetConfigRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]awstypes.WidgetConfig{
		"cost_and_usage graph": {
			QueryParameters: &awstypes.QueryParametersMemberCostAndUsage{
				Value: awstypes.CostAndUsageQuery{
					Granularity: awstypes.GranularityMonthly,
					Metrics:     []awstypes.MetricName{awstypes.MetricNameUnblendedCost},
					TimeRange: &awstypes.DateTimeRange{
						StartTime: &awstypes.DateTimeValue{Type: awstypes.DateTimeTypeAbsolute, Value: aws.String("2025-01-01")},
						EndTime:   &awstypes.DateTimeValue{Type: awstypes.DateTimeTypeAbsolute, Value: aws.String("2025-03-31")},
					},
				},
			},
			DisplayConfig: &awstypes.DisplayConfigMemberGraph{
				Value: map[string]awstypes.GraphDisplayConfig{
					"UnblendedCost": {VisualType: awstypes.VisualTypeBar},
				},
			},
		},
		"cost_and_usage with group_by and filter, table": {
			QueryParameters: &awstypes.QueryParametersMemberCostAndUsage{
				Value: awstypes.CostAndUsageQuery{
					Granularity: awstypes.GranularityDaily,
					Metrics:     []awstypes.MetricName{awstypes.MetricNameUnblendedCost, awstypes.MetricNameUsageQuantity},
					TimeRange: &awstypes.DateTimeRange{
						StartTime: &awstypes.DateTimeValue{Type: awstypes.DateTimeTypeRelative, Value: aws.String("-P3M")},
						EndTime:   &awstypes.DateTimeValue{Type: awstypes.DateTimeTypeRelative, Value: aws.String("P0D")},
					},
					GroupBy: []awstypes.GroupDefinition{
						{Key: aws.String("SERVICE"), Type: awstypes.GroupDefinitionTypeDimension},
					},
					Filter: &awstypes.Expression{
						And: []awstypes.Expression{
							{Tags: &awstypes.TagValues{Key: aws.String("Environment"), Values: []string{"production"}}},
							{Dimensions: &awstypes.DimensionValues{Key: awstypes.DimensionUsageType, Values: []string{"DataTransfer-In-Bytes"}}},
						},
					},
				},
			},
			DisplayConfig: &awstypes.DisplayConfigMemberTable{
				Value: awstypes.TableDisplayConfigStruct{},
			},
		},
		"savings_plans_coverage": {
			QueryParameters: &awstypes.QueryParametersMemberSavingsPlansCoverage{
				Value: awstypes.SavingsPlansCoverageQuery{
					Granularity: awstypes.GranularityMonthly,
					Metrics:     []awstypes.MetricName{awstypes.MetricNameSpendCoveredBySavingsPlans},
					TimeRange: &awstypes.DateTimeRange{
						StartTime: &awstypes.DateTimeValue{Type: awstypes.DateTimeTypeAbsolute, Value: aws.String("2025-01-01")},
						EndTime:   &awstypes.DateTimeValue{Type: awstypes.DateTimeTypeAbsolute, Value: aws.String("2025-02-01")},
					},
				},
			},
			DisplayConfig: &awstypes.DisplayConfigMemberGraph{
				Value: map[string]awstypes.GraphDisplayConfig{
					"SpendCoveredBySavingsPlans": {VisualType: awstypes.VisualTypeLine},
				},
			},
		},
		"reservation_utilization not-filter": {
			QueryParameters: &awstypes.QueryParametersMemberReservationUtilization{
				Value: awstypes.ReservationUtilizationQuery{
					Granularity: awstypes.GranularityMonthly,
					TimeRange: &awstypes.DateTimeRange{
						StartTime: &awstypes.DateTimeValue{Type: awstypes.DateTimeTypeAbsolute, Value: aws.String("2025-01-01")},
						EndTime:   &awstypes.DateTimeValue{Type: awstypes.DateTimeTypeAbsolute, Value: aws.String("2025-02-01")},
					},
					Filter: &awstypes.Expression{
						Not: &awstypes.Expression{
							Dimensions: &awstypes.DimensionValues{Key: awstypes.DimensionRegion, Values: []string{"us-east-1"}},
						},
					},
				},
			},
			DisplayConfig: &awstypes.DisplayConfigMemberGraph{
				Value: map[string]awstypes.GraphDisplayConfig{
					"Cost": {VisualType: awstypes.VisualTypeStack},
				},
			},
		},
	}

	for name, want := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var model widgetConfigModel
			if diags := model.Flatten(ctx, want); diags.HasError() {
				t.Fatalf("Flatten: %v", diags)
			}

			got, diags := model.Expand(ctx)
			if diags.HasError() {
				t.Fatalf("Expand: %v", diags)
			}

			if diff := cmp.Diff(&want, got, cmpopts.IgnoreUnexported(
				awstypes.WidgetConfig{},
				awstypes.QueryParametersMemberCostAndUsage{},
				awstypes.QueryParametersMemberReservationCoverage{},
				awstypes.QueryParametersMemberReservationUtilization{},
				awstypes.QueryParametersMemberSavingsPlansCoverage{},
				awstypes.QueryParametersMemberSavingsPlansUtilization{},
				awstypes.DisplayConfigMemberGraph{},
				awstypes.DisplayConfigMemberTable{},
				awstypes.CostAndUsageQuery{},
				awstypes.ReservationCoverageQuery{},
				awstypes.ReservationUtilizationQuery{},
				awstypes.SavingsPlansCoverageQuery{},
				awstypes.SavingsPlansUtilizationQuery{},
				awstypes.DateTimeRange{},
				awstypes.DateTimeValue{},
				awstypes.GroupDefinition{},
				awstypes.Expression{},
				awstypes.DimensionValues{},
				awstypes.TagValues{},
				awstypes.CostCategoryValues{},
				awstypes.GraphDisplayConfig{},
				awstypes.TableDisplayConfigStruct{},
			)); diff != "" {
				t.Errorf("round-trip mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestWidgetConfigExpandValidation verifies that an empty query_parameters or
// display_config selection produces a diagnostic rather than a silent zero value.
func TestWidgetConfigExpandValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var empty queryParametersModel
	if _, diags := expandQueryParameters(ctx, empty); !diags.HasError() {
		t.Error("expected error expanding empty query_parameters, got none")
	}

	var emptyDisplay displayConfigModel
	if _, diags := expandDisplayConfig(ctx, emptyDisplay); !diags.HasError() {
		t.Error("expected error expanding empty display_config, got none")
	}
}
