package autoscaling

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

func TestFlattenEnabledMetrics(t *testing.T) {
	expanded := []*autoscaling.EnabledMetric{
		{Granularity: aws.String("1Minute"), Metric: aws.String("GroupTotalInstances")},
		{Granularity: aws.String("1Minute"), Metric: aws.String("GroupMaxSize")},
	}

	result := flattenASGEnabledMetrics(expanded)

	if len(result) != 2 {
		t.Fatalf("expected result had %d elements, but got %d", 2, len(result))
	}

	if result[0] != "GroupTotalInstances" {
		t.Fatalf("expected id to be GroupTotalInstances, but was %s", result[0])
	}

	if result[1] != "GroupMaxSize" {
		t.Fatalf("expected id to be GroupMaxSize, but was %s", result[1])
	}
}

func TestExpandStepAdjustments(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"metric_interval_lower_bound": "1.0",
			"metric_interval_upper_bound": "2.0",
			"scaling_adjustment":          1,
		},
	}
	parameters, err := ExpandStepAdjustments(expanded)
	if err != nil {
		t.Fatalf("bad: %#v", err)
	}

	expected := &autoscaling.StepAdjustment{
		MetricIntervalLowerBound: aws.Float64(1.0),
		MetricIntervalUpperBound: aws.Float64(2.0),
		ScalingAdjustment:        aws.Int64(int64(1)),
	}

	if !reflect.DeepEqual(parameters[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			parameters[0],
			expected)
	}
}

func TestFlattenStepAdjustments(t *testing.T) {
	expanded := []*autoscaling.StepAdjustment{
		{
			MetricIntervalLowerBound: aws.Float64(1.0),
			MetricIntervalUpperBound: aws.Float64(2.5),
			ScalingAdjustment:        aws.Int64(int64(1)),
		},
	}

	result := FlattenStepAdjustments(expanded)[0]
	if result == nil {
		t.Fatal("expected result to have value, but got nil")
	}
	if result["metric_interval_lower_bound"] != "1" {
		t.Fatalf("expected metric_interval_lower_bound to be 1, but got %s", result["metric_interval_lower_bound"])
	}
	if result["metric_interval_upper_bound"] != "2.5" {
		t.Fatalf("expected metric_interval_upper_bound to be 2.5, but got %s", result["metric_interval_upper_bound"])
	}
	if result["scaling_adjustment"] != int64(1) {
		t.Fatalf("expected scaling_adjustment to be 1, but got %d", result["scaling_adjustment"])
	}
}
