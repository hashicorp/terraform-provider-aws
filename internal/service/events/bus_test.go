// Copyright (c) HashiCorp, Inc.
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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsBus_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeEventBusOutput
	busName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	busNameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_basic(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, resourceName, &v1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, busName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig_basic(busNameModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, resourceName, &v2),
					testAccCheckBusRecreated(&v1, &v2),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "events", fmt.Sprintf("event-bus/%s", busNameModified)),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, busNameModified),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccBusConfig_tags1(busNameModified, names.AttrKey, names.AttrValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, resourceName, &v3),
					testAccCheckBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.key", names.AttrValue),
				),
			},
		},
	})
}

func TestAccEventsBus_kmsKeyIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 eventbridge.DescribeEventBusOutput
	busName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_kmsKeyIdentifier1(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, resourceName, &v1),
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
					testAccCheckBusExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_key.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccEventsBus_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeEventBusOutput
	busName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_tags1(busName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					testAccCheckBusExists(ctx, resourceName, &v2),
					testAccCheckBusNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBusConfig_tags1(busName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, resourceName, &v3),
					testAccCheckBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEventsBus_default(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx),
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
	busName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_basic(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfevents.ResourceBus(), resourceName),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_partnerSource(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBusExists(ctx, resourceName, &busOutput),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckResourceAttr(resourceName, "event_source_name", busName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, busName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckBusDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_event_bus" {
				continue
			}

			_, err := tfevents.FindEventBusByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckBusExists(ctx context.Context, n string, v *eventbridge.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsClient(ctx)

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
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
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
