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
					resource.TestCheckResourceAttr(resourceName, "transformer_config.0.parse_json.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTransformerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "log_group_identifier",
			},
		},
	})
}

func TestAccLogsTransformer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_transformer.test"

	resource.ParallelTest(t, resource.TestCase{
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
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflogs.ResourceTransformer, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsTransformer_disappears_logGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var transformer cloudwatchlogs.GetTransformerOutput
	resourceName := "aws_cloudwatch_log_transformer.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransformerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransformerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransformerExists(ctx, t, resourceName, &transformer),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceGroup(), logGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
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

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		resp, err := tflogs.FindTransformerByLogGroupIdentifier(ctx, conn, rs.Primary.Attributes["log_group_identifier"])
		if err != nil {
			return create.Error(names.Logs, create.ErrActionCheckingExistence, tflogs.ResNameTransformer, rs.Primary.Attributes["log_group_identifier"], err)
		}

		*transformer = *resp

		return nil
	}
}

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
