// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestExpandStepAdjustments(t *testing.T) {
	t.Parallel()

	tfList := []interface{}{
		map[string]interface{}{
			"metric_interval_lower_bound": "1.0",
			"metric_interval_upper_bound": "2.0",
			"scaling_adjustment":          1,
		},
	}
	stepAdjustments := expandStepAdjustments(tfList)

	stepAdjustment := &awstypes.StepAdjustment{
		MetricIntervalLowerBound: aws.Float64(1.0),
		MetricIntervalUpperBound: aws.Float64(2.0),
		ScalingAdjustment:        aws.Int32(1),
	}

	var got, want any = &stepAdjustments[0], stepAdjustment
	if diff := cmp.Diff(got, want, cmpopts.IgnoreUnexported(awstypes.StepAdjustment{})); diff != "" {
		t.Fatalf("unexpected expandStepAdjustments diff (+wanted, -got): %s", diff)
	}
}

func TestFlattenStepAdjustments(t *testing.T) {
	t.Parallel()

	stepAdjustments := []awstypes.StepAdjustment{
		{
			MetricIntervalLowerBound: aws.Float64(1.0),
			MetricIntervalUpperBound: aws.Float64(2.5),
			ScalingAdjustment:        aws.Int32(1),
		},
	}

	got := flattenStepAdjustments(stepAdjustments)[0]
	want := map[string]interface{}{
		"metric_interval_lower_bound": "1",
		"metric_interval_upper_bound": "2.5",
		"scaling_adjustment":          int32(1),
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("unexpected flattenStepAdjustments diff (+wanted, -got): %s", diff)
	}
}
