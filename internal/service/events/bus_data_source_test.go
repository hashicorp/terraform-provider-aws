// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsBusDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_event_bus.test"
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBusDataSourceConfig_basic(busName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccEventsBusDataSource_kmsKeyIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_event_bus.test"
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBusDataSourceConfig_kmsKeyIdentifier(busName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_identifier", resourceName, "kms_key_identifier"),
				),
			},
		},
	})
}

func TestAccEventsBusDataSource_deadLetterConfig(t *testing.T) {
	ctx := acctest.Context(t)
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_event_bus.test"
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBusDataSourceConfig_deadLetterConfig(busName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "dead_letter_config.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dead_letter_config.0.arn", resourceName, "dead_letter_config.0.arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
				),
			},
		},
	})
}

func TestAccEventsBusDataSource_logConfig(t *testing.T) {
	ctx := acctest.Context(t)
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudwatch_event_bus.test"
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBusDataSourceConfig_logConfig(busName, "FULL", "TRACE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "log_config.0.include_detail", resourceName, "log_config.0.include_detail"),
					resource.TestCheckResourceAttrPair(dataSourceName, "log_config.0.level", resourceName, "log_config.0.level"),
				),
			},
		},
	})
}

func testAccBusDataSourceConfig_basic(busName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

data "aws_cloudwatch_event_bus" "test" {
  name = aws_cloudwatch_event_bus.test.name
}
`, busName)
}

func testAccBusDataSourceConfig_kmsKeyIdentifier(busName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

data "aws_iam_policy_document" "key_policy" {
  statement {
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]

    resources = [
      aws_kms_key.test.arn,
    ]

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }

  statement {
    actions = [
      "kms:*",
    ]

    resources = [
      aws_kms_key.test.arn
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id
  policy = data.aws_iam_policy_document.key_policy.json
}

resource "aws_cloudwatch_event_bus" "test" {
  name               = %[1]q
  description        = "Test event bus"
  kms_key_identifier = aws_kms_key.test.arn
}

data "aws_cloudwatch_event_bus" "test" {
  name = aws_cloudwatch_event_bus.test.name
}
`, busName)
}

func testAccBusDataSourceConfig_deadLetterConfig(busName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
  dead_letter_config {
    arn = aws_sqs_queue.test.arn
  }
}

data "aws_cloudwatch_event_bus" "test" {
  name = aws_cloudwatch_event_bus.test.name
}
`, busName)
}

func testAccBusDataSourceConfig_logConfig(name, includeDetail, level string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
  log_config {
    include_detail = %[2]q
    level          = %[3]q
  }
}

data "aws_cloudwatch_event_bus" "test" {
  name = aws_cloudwatch_event_bus.test.name
}
`, name, includeDetail, level)
}
