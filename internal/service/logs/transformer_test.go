// Copyright (c) HashiCorp, Inc.
// Copyright 2025 Twilio Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

func TestAccLogsTransformer_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_identifier", logGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "transformer_config.parse_json.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsTransformer_disappears(t *testing.T) {
	t.Skip("temporarily disabled")

	// ctx := acctest.Context(t)
	// var transformer cloudwatchlogs.GetTransformerOutput
	// rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// resourceName := "aws_logs_transformer.test"

	// resource.ParallelTest(t, resource.TestCase{
	// 	PreCheck: func() {
	// 		acctest.PreCheck(ctx, t)
	// 		acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
	// 	},
	// 	ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
	// 	ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
	// 	CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
	// 	Steps: []resource.TestStep{
	// 		{
	// 			Config: testAccTransformerConfig_basic(rName, testAccTransformerVersionNewer),
	// 			Check: resource.ComposeAggregateTestCheckFunc(
	// 				testAccCheckTransformerExists(ctx, resourceName, &transformer),
	// 				// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
	// 				// but expects a new resource factory function as the third argument. To expose this
	// 				// private function to the testing package, you may need to add a line like the following
	// 				// to exports_test.go:
	// 				//
	// 				//   var ResourceTransformer = newResourceTransformer
	// 				acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflogs.ResourceTransformer, resourceName),
	// 			),
	// 			ExpectNonEmptyPlan: true,
	// 			ConfigPlanChecks: resource.ConfigPlanChecks{
	// 				PostApplyPostRefresh: []plancheck.PlanCheck{
	// 					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
	// 				},
	// 			},
	// 		},
	// 	},
	// })
}

func testAccTransformerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["log_group_identifier"], nil
	}
}

func testAccCheckTransformerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_logs_transformer" {
				continue
			}

			_, err := tflogs.FindTransformerByLogGroupIdentifier(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Logs, create.ErrActionCheckingDestroyed, tflogs.ResNameTransformer, rs.Primary.ID, err)
			}

			return create.Error(names.Logs, create.ErrActionCheckingDestroyed, tflogs.ResNameTransformer, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTransformerExists(ctx context.Context, t *testing.T, name string, transformer *cloudwatchlogs.GetTransformerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameTransformer, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameTransformer, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		resp, err := tflogs.FindTransformerByLogGroupIdentifier(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameTransformer, rs.Primary.ID, err)
		}

		*transformer = *resp

		return nil
	}
}

// func testAccCheckTransformerNotRecreated(before, after *cloudwatchlogs.DescribeTransformerResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.TransformerId), aws.ToString(after.TransformerId); before != after {
// 			return create.Error(names.Logs, create.ErrActionCheckingNotRecreated, tflogs.ResNameTransformer, aws.ToString(before.TransformerId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccTransformerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_transformer" "test" {
  log_group_identifier = aws_cloudwatch_log_group.test.name

  transformer_config {
    parse_json {}
  }
}
`, rName)
}
