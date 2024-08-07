// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/autoscaling"
)

func TestGroupStateUpgradeV0(t *testing.T) {
	ctx := acctest.Context(t)
	t.Parallel()

	testCases := []struct {
		testName string
		rawState map[string]interface{}
		want     map[string]interface{}
	}{
		{
			testName: "empty state",
			rawState: map[string]interface{}{},
			want: map[string]interface{}{
				"ignore_failed_scaling_activities": acctest.CtFalse,
			},
		},
		{
			testName: "non-empty state",
			rawState: map[string]interface{}{
				"capacity_rebalance":        acctest.CtTrue,
				"health_check_grace_period": "600",
				"max_instance_lifetime":     "3600",
			},
			want: map[string]interface{}{
				"capacity_rebalance":               acctest.CtTrue,
				"health_check_grace_period":        "600",
				"ignore_failed_scaling_activities": acctest.CtFalse,
				"max_instance_lifetime":            "3600",
			},
		},
		{
			testName: "ignore_failed_scaling_activities set",
			rawState: map[string]interface{}{
				"capacity_rebalance":               acctest.CtFalse,
				"health_check_grace_period":        "400",
				"ignore_failed_scaling_activities": acctest.CtTrue,
				"max_instance_lifetime":            "36000",
			},
			want: map[string]interface{}{
				"capacity_rebalance":               acctest.CtFalse,
				"health_check_grace_period":        "400",
				"ignore_failed_scaling_activities": acctest.CtTrue,
				"max_instance_lifetime":            "36000",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got, err := tfautoscaling.GroupStateUpgradeV0(ctx, testCase.rawState, nil)

			if err != nil {
				t.Errorf("err = %q", err)
			} else if diff := cmp.Diff(got, testCase.want); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
