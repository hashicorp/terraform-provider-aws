// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "logs", fmt.Sprintf("log-group:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "log_group_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "skip_destroy"},
			},
		},
	})
}

func TestAccLogsGroup_nameGenerate(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "skip_destroy"},
			},
		},
	})
}

func TestAccLogsGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "skip_destroy"},
			},
		},
	})
}

func TestAccLogsGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "skip_destroy"},
			},
			{
				Config: testAccGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccLogsGroup_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"
	kmsKey1ResourceName := "aws_kms_key.test.0"
	kmsKey2ResourceName := "aws_kms_key.test.1"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_kmsKey(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKey1ResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "skip_destroy"},
			},
			{
				Config: testAccGroupConfig_kmsKey(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKey2ResourceName, "arn"),
				),
			},
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func TestAccLogsGroup_logGroupClass(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_logGroupClass(rName, "INFREQUENT_ACCESS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "log_group_class", "INFREQUENT_ACCESS"),
				),
			},
		},
	})
}

func TestAccLogsGroup_retentionPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_retentionPolicy(rName, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "365"),
				),
			},
			{
				Config: testAccGroupConfig_retentionPolicy(rName, 1096),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "1096"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "skip_destroy"},
			},
			{
				Config: testAccGroupConfig_retentionPolicy(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
				),
			},
		},
	})
}

func TestAccLogsGroup_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resource1Name := "aws_cloudwatch_log_group.test.0"
	resource2Name := "aws_cloudwatch_log_group.test.1"
	resource3Name := "aws_cloudwatch_log_group.test.2"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resource1Name, &v1),
					testAccCheckGroupExists(ctx, t, resource2Name, &v2),
					testAccCheckGroupExists(ctx, t, resource3Name, &v3),
				),
			},
		},
	})
}

func TestAccLogsGroup_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchLogsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupNoDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "true"),
				),
			},
		},
	})
}

func testAccCheckGroupExists(ctx context.Context, t *testing.T, n string, v *types.LogGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(t).LogsClient(ctx)

		output, err := tflogs.FindLogGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_group" {
				continue
			}

			_, err := tflogs.FindLogGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Log Group still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGroupNoDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_group" {
				continue
			}

			_, err := tflogs.FindLogGroupByName(ctx, conn, rs.Primary.ID)

			return err
		}

		return nil
	}
}

func testAccGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccGroupConfig_nameGenerated() string {
	return `
resource "aws_cloudwatch_log_group" "test" {}
`
}

func testAccGroupConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name_prefix = %[1]q
}
`, namePrefix)
}

func testAccGroupConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccGroupConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccGroupConfig_kmsKey(rName string, idx int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count = 2

  description             = "%[1]s-${count.index}"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": "kms:*",
    "Resource": "*"
  }]
}
POLICY
}

resource "aws_cloudwatch_log_group" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test[%[2]d].arn
}
`, rName, idx)
}

func testAccGroupConfig_logGroupClass(rName string, val string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name            = %[1]q
  log_group_class = %[2]q
}
`, rName, val)
}

func testAccGroupConfig_retentionPolicy(rName string, val int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = %[2]d
}
`, rName, val)
}

func testAccGroupConfig_multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  count = 3

  name = "%[1]s-${count.index}"
}
`, rName)
}

func testAccGroupConfig_skipDestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name         = %[1]q
  skip_destroy = true
}
`, rName)
}
