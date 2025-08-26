// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package computeoptimizer_test

import (
	"context"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/computeoptimizer/types"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcomputeoptimizer "github.com/hashicorp/terraform-provider-aws/internal/service/computeoptimizer"
)

func TestAccComputeOptimizer_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"EnrollmentStatus": {
			acctest.CtBasic:         testAccEnrollmentStatus_basic,
			"includeMemberAccounts": testAccEnrollmentStatus_includeMemberAccounts,
		},
		"RecommendationPreferences": {
			acctest.CtBasic:          testAccRecommendationPreferences_basic,
			acctest.CtDisappears:     testAccRecommendationPreferences_disappears,
			"preferredResources":     testAccRecommendationPreferences_preferredResources,
			"utilizationPreferences": testAccRecommendationPreferences_utilizationPreferences,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheckEnrollmentStatus(ctx context.Context, t *testing.T, want awstypes.Status) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ComputeOptimizerClient(ctx)

	output, err := tfcomputeoptimizer.FindEnrollmentStatus(ctx, conn)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if got := output.Status; got != want {
		t.Fatalf("Compute Optimizer enrollment status: %s", got)
	}
}
