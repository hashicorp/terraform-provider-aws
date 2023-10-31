// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccLogsStream_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ls cloudwatchlogs.LogStream
	resourceName := "aws_cloudwatch_log_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &ls),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStreamImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsStream_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ls cloudwatchlogs.LogStream
	resourceName := "aws_cloudwatch_log_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &ls),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsStream_Disappears_logGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var ls cloudwatchlogs.LogStream
	resourceName := "aws_cloudwatch_log_stream.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamExists(ctx, resourceName, &ls),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceGroup(), logGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStreamExists(ctx context.Context, n string, v *cloudwatchlogs.LogStream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Logs Log Stream ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn(ctx)

		output, err := tflogs.FindLogStreamByTwoPartKey(ctx, conn, rs.Primary.Attributes["log_group_name"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStreamDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_stream" {
				continue
			}

			_, err := tflogs.FindLogStreamByTwoPartKey(ctx, conn, rs.Primary.Attributes["log_group_name"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Log Stream still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStreamImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", n)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["log_group_name"], rs.Primary.ID), nil
	}
}

func testAccStreamConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.test.id
}
`, rName)
}
