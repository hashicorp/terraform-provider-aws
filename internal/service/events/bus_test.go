// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsBus_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3, v4, v5 eventbridge.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	busNameModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"
	description := "Test event bus"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_basic(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, busName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig_description(busName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v2),
					testAccCheckBusNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
				),
			},
			{
				Config: testAccBusConfig_basic(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v3),
					testAccCheckBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				Config: testAccBusConfig_basic(busNameModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v4),
					testAccCheckBusRecreated(&v3, &v4),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", fmt.Sprintf("event-bus/%s", busNameModified)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, busNameModified),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccBusConfig_tags1(busNameModified, names.AttrKey, names.AttrValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v5),
					testAccCheckBusNotRecreated(&v4, &v5),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", names.AttrValue),
				),
			},
		},
	})
}

func TestAccEventsBus_kmsKeyIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 eventbridge.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_kmsKeyIdentifier1(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_key.test1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig_kmsKeyIdentifier2(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_key.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccEventsBus_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_tags1(busName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig_tags2(busName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v2),
					testAccCheckBusNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBusConfig_tags1(busName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v3),
					testAccCheckBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEventsBus_default(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccBusConfig_basic("default"),
				ExpectError: regexache.MustCompile(`cannot be 'default'`),
			},
		},
	})
}

func TestAccEventsBus_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_basic(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfevents.ResourceBus(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEventsBus_partnerEventSource(t *testing.T) {
	ctx := acctest.Context(t)
	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	busName := os.Getenv(key)
	if busName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var busOutput eventbridge.DescribeEventBusOutput
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_partnerSource(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &busOutput),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckResourceAttr(resourceName, "event_source_name", busName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, busName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccEventsBus_deadLetterConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 eventbridge.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_deadLetterConfig1(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "dead_letter_config.0.arn", "aws_sqs_queue.test1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig_deadLetterConfig2(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "dead_letter_config.0.arn", "aws_sqs_queue.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccEventsBus_logConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeEventBusOutput
	busName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_logConfig(busName, "FULL", "TRACE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.include_detail", "FULL"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.level", "TRACE"),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, busName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig_logConfig(busName, "NONE", "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, t, resourceName, &v1),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "log_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.include_detail", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "log_config.0.level", "OFF"),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, busName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func testAccCheckBusDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_event_bus" {
				continue
			}

			_, err := tfevents.FindEventBusByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Event Bus %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBusExists(ctx context.Context, t *testing.T, n string, v *eventbridge.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		output, err := tfevents.FindEventBusByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBusRecreated(i, j *eventbridge.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.Arn) == aws.ToString(j.Arn) {
			return fmt.Errorf("EventBridge Event Bus not recreated")
		}
		return nil
	}
}

func testAccCheckBusNotRecreated(i, j *eventbridge.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.Arn) != aws.ToString(j.Arn) {
			return fmt.Errorf("EventBridge Event Bus was recreated")
		}
		return nil
	}
}

func testAccBusConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}
`, name)
}

func testAccBusConfig_description(name, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}

func testAccBusConfig_tags1(name, key, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, name, key, value)
}

func testAccBusConfig_tags2(name, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, key1, value1, key2, value2)
}

func testAccBusConfig_partnerSource(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name              = %[1]q
  event_source_name = %[1]q
}
`, name)
}

func testAccBusConfig_kmsKeyIdentifierBase() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "test2" {
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
      aws_kms_key.test1.arn,
      aws_kms_key.test2.arn,
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
      aws_kms_key.test1.arn,
      aws_kms_key.test2.arn,
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

resource "aws_kms_key_policy" "test1" {
  key_id = aws_kms_key.test1.id
  policy = data.aws_iam_policy_document.key_policy.json
}

resource "aws_kms_key_policy" "test2" {
  key_id = aws_kms_key.test2.id
  policy = data.aws_iam_policy_document.key_policy.json
}
`
}

func testAccBusConfig_kmsKeyIdentifier1(name string) string {
	return acctest.ConfigCompose(
		testAccBusConfig_kmsKeyIdentifierBase(),
		fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name               = %[1]q
  kms_key_identifier = aws_kms_key.test1.arn
}
`, name))
}

func testAccBusConfig_kmsKeyIdentifier2(name string) string {
	return acctest.ConfigCompose(
		testAccBusConfig_kmsKeyIdentifierBase(),
		fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name               = %[1]q
  kms_key_identifier = aws_kms_key.test2.arn
}
`, name))
}

func testAccBusConfig_deadLetterConfig1(name string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test1" {
  name = "%[1]s-test1"
}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
  dead_letter_config {
    arn = aws_sqs_queue.test1.arn
  }
}
`, name)
}

func testAccBusConfig_deadLetterConfig2(name string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test2" {
  name = "%[1]s-test2"
}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
  dead_letter_config {
    arn = aws_sqs_queue.test2.arn
  }
}
`, name)
}

func testAccBusConfig_logConfig(name, includeDetail, level string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
  log_config {
    include_detail = %[2]q
    level          = %[3]q
  }
}
`, name, includeDetail, level)
}
