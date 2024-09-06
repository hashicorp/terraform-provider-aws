// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package computeoptimizer_test

import (
	// "context"
	// "errors"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer"
	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	// "github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"

	// "github.com/hashicorp/terraform-provider-aws/internal/create"

	// "github.com/hashicorp/terraform-provider-aws/internal/conns"
	// "github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfcomputeoptimizer "github.com/hashicorp/terraform-provider-aws/internal/service/computeoptimizer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccComputeOptimizerRecommendationPreferences_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var recommendationpreferences computeoptimizer.GetRecommendationPreferencesOutput
	resourceType := "Ec2Instance"
	resourceName := "aws_computeoptimizer_recommendation_preferences.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t, resourceType) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ComputeOptimizerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecommendationPreferencesDestroy(ctx, resourceType),
		Steps: []resource.TestStep{
			{
				Config: testAccRecommendationPreferencesConfig_basic(resourceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecommendationPreferencesExists(ctx, resourceName, &recommendationpreferences),
				),
			},
			// {
			// 	ResourceName:            resourceName,
			// 	ImportState:             true,
			// 	ImportStateVerify:       true,
			// 	ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			// },
		},
	})
}

func TestAccComputeOptimizerRecommendationPreferences_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var recommendationpreferences computeoptimizer.GetRecommendationPreferencesOutput
	resourceType := "Ec2Instance"
	resourceName := "aws_computeoptimizer_recommendation_preferences.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t, resourceType) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ComputeOptimizerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecommendationPreferencesDestroy(ctx, resourceType),
		Steps: []resource.TestStep{
			{
				// Config: testAccRecommendationPreferencesConfig_basic(rName, testAccRecommendationPreferencesVersionNewer),
				Config: testAccRecommendationPreferencesConfig_basic(resourceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecommendationPreferencesExists(ctx, resourceName, &recommendationpreferences),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcomputeoptimizer.ResourceRecommendationPreferences(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRecommendationPreferencesDestroy(ctx context.Context, resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ComputeOptimizerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_computeoptimizer_recommendation_preferences" {
				continue
			}

			input := &computeoptimizer.GetRecommendationPreferencesInput{
				ResourceType: types.ResourceType(resourceType),
			}
			_, err := conn.GetRecommendationPreferences(ctx, input)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ComputeOptimizer, create.ErrActionCheckingDestroyed, tfcomputeoptimizer.ResNameRecommendationPreferences, rs.Primary.ID, err)
			}

			return create.Error(names.ComputeOptimizer, create.ErrActionCheckingDestroyed, tfcomputeoptimizer.ResNameRecommendationPreferences, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRecommendationPreferencesExists(ctx context.Context, name string, recommendationpreferences *computeoptimizer.GetRecommendationPreferencesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ComputeOptimizer, create.ErrActionCheckingExistence, tfcomputeoptimizer.ResNameRecommendationPreferences, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ComputeOptimizer, create.ErrActionCheckingExistence, tfcomputeoptimizer.ResNameRecommendationPreferences, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ComputeOptimizerClient(ctx)
		resp, err := conn.GetRecommendationPreferences(ctx, &computeoptimizer.GetRecommendationPreferencesInput{
			ResourceType: types.ResourceType(rs.Primary.Attributes["resource_type"]),
		})

		if err != nil {
			return create.Error(names.ComputeOptimizer, create.ErrActionCheckingExistence, tfcomputeoptimizer.ResNameRecommendationPreferences, rs.Primary.ID, err)
		}

		*recommendationpreferences = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T, resourceType string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ComputeOptimizerClient(ctx)

	input := &computeoptimizer.GetRecommendationPreferencesInput{
		ResourceType: types.ResourceType(resourceType),
	}

	_, err := conn.GetRecommendationPreferences(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// func testAccCheckRecommendationPreferencesNotRecreated(before, after *computeoptimizer.DescribeRecommendationPreferencesResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.RecommendationPreferencesId), aws.ToString(after.RecommendationPreferencesId); before != after {
// 			return create.Error(names.ComputeOptimizer, create.ErrActionCheckingNotRecreated, tfcomputeoptimizer.ResNameRecommendationPreferences, aws.ToString(before.RecommendationPreferencesId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccRecommendationPreferencesConfig_basic(resourceType string) string {
	return fmt.Sprintf(`
resource "aws_computeoptimizer_recommendation_preferences" "test" {
  resource_type = %[1]q
}
`, resourceType)
}
