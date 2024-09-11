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

	expectedLogGroupClass := "STANDARD"
	if acctest.Partition() != names.StandardPartitionID {
		expectedLogGroupClass = ""
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "logs", fmt.Sprintf("log-group:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "log_group_class", expectedLogGroupClass),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccLogsGroup_nameGenerate(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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

func TestAccLogsGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccLogsGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflogs.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_kmsKey(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKey1ResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupConfig_kmsKey(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKey2ResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// CloudWatch Logs IA is available in all AWS Commercial regions.
			acctest.PreCheckPartition(t, names.StandardPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_logGroupClass(rName, "INFREQUENT_ACCESS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_retentionPolicy(rName, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "365"),
				),
			},
			{
				Config: testAccGroupConfig_retentionPolicy(rName, 1096),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "1096"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupConfig_retentionPolicy(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", acctest.Ct0),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resource1Name, &v1),
					testAccCheckLogGroupExists(ctx, t, resource2Name, &v2),
					testAccCheckLogGroupExists(ctx, t, resource3Name, &v3),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupNoDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccLogsGroup_skipDestroyInconsistentPlan(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
				),
			},
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckLogGroupExists(ctx context.Context, t *testing.T, n string, v *types.LogGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindLogGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLogGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

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
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

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
