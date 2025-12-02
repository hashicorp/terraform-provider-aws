// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsGroupRetention_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group_retention.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupRetentionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRetentionConfig_basic(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, logGroupResourceName, &v),
					testAccCheckGroupRetentionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrLogGroupName, rName),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "7"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsGroupRetention_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group_retention.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupRetentionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRetentionConfig_basic(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, logGroupResourceName, &v),
					testAccCheckGroupRetentionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "7"),
				),
			},
			{
				Config: testAccGroupRetentionConfig_basic(rName, 14),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, logGroupResourceName, &v),
					testAccCheckGroupRetentionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "14"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsGroupRetention_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group_retention.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupRetentionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRetentionConfig_basic(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, logGroupResourceName, &v),
					testAccCheckGroupRetentionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceGroupRetention(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsGroupRetention_retentionValues(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group_retention.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupRetentionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRetentionConfig_basic(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, logGroupResourceName, &v),
					testAccCheckGroupRetentionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "30"),
				),
			},
			{
				Config: testAccGroupRetentionConfig_basic(rName, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, logGroupResourceName, &v),
					testAccCheckGroupRetentionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "365"),
				),
			},
			{
				Config: testAccGroupRetentionConfig_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, logGroupResourceName, &v),
					testAccCheckGroupRetentionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsGroupRetention_disappears_LogGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group_retention.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupRetentionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupRetentionConfig_basic(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, logGroupResourceName, &v),
					testAccCheckGroupRetentionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceGroup(), logGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGroupRetentionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_group_retention" {
				continue
			}

			lg, err := tflogs.FindLogGroupByName(ctx, conn, rs.Primary.ID)
			if err != nil {
				// If the log group is gone, the retention policy is also gone
				continue
			}

			// Check if retention policy still exists
			if lg.RetentionInDays != nil {
				return fmt.Errorf("CloudWatch Logs Log Group retention policy %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckGroupRetentionExists(ctx context.Context, n string, v *types.LogGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		lg, err := tflogs.FindLogGroupByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if lg.RetentionInDays == nil {
			return fmt.Errorf("CloudWatch Logs Log Group retention policy %s not found", rs.Primary.ID)
		}

		*v = *lg

		return nil
	}
}

func testAccGroupRetentionConfig_basic(rName string, retentionInDays int) string {
	return fmt.Sprintf(`
# Create the log group first without any retention policy
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
  # Explicitly manage lifecycle to prevent retention conflicts
  lifecycle {
    ignore_changes = [retention_in_days]
  }
}

resource "aws_cloudwatch_log_group_retention" "test" {
  log_group_name    = aws_cloudwatch_log_group.test.name
  retention_in_days = %[2]d
}
`, rName, retentionInDays)
}
