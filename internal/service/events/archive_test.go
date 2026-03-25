// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsArchive_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeArchiveOutput
	archiveName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_archive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_basic(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, archiveName),
					resource.TestCheckResourceAttr(resourceName, "retention_days", "0"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", fmt.Sprintf("archive/%s", archiveName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_key_identifier", ""),
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

func TestAccEventsArchive_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeArchiveOutput
	resourceName := "aws_cloudwatch_event_archive.test"
	archiveName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_basic(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName, &v1),
				),
			},
			{
				Config: testAccArchiveConfig_updateAttributes(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "retention_days", "7"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"company.team.service\"]}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
				),
			},
		},
	})
}

func TestAccEventsArchive_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeArchiveOutput
	archiveName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_archive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_basic(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfevents.ResourceArchive(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEventsArchive_kmsKeyIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeArchiveOutput
	archiveName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_archive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_kmsKeyIdentifier(archiveName, "${aws_kms_key.test_1.id}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, archiveName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_key.test_1", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccArchiveConfig_kmsKeyIdentifier(archiveName, "${aws_kms_key.test_2.arn}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, archiveName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_key.test_2", names.AttrARN),
				),
			},
			{
				Config: testAccArchiveConfig_kmsKeyIdentifier(archiveName, "${aws_kms_alias.test_1.name}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, archiveName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_alias.test_1", names.AttrName),
				),
			},
			{
				Config: testAccArchiveConfig_kmsKeyIdentifier(archiveName, "${aws_kms_alias.test_1.arn}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, archiveName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_alias.test_1", names.AttrARN),
				),
			},
		},
	})
}

func TestAccEventsArchive_retentionSetOnCreation(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 eventbridge.DescribeArchiveOutput
	archiveName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_archive.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveConfig_retentionOnCreation(archiveName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, archiveName),
					resource.TestCheckResourceAttr(resourceName, "retention_days", "1"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", fmt.Sprintf("archive/%s", archiveName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", ""),
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

func testAccCheckArchiveDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_event_archive" {
				continue
			}

			_, err := tfevents.FindArchiveByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Archive %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckArchiveExists(ctx context.Context, t *testing.T, n string, v *eventbridge.DescribeArchiveOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		output, err := tfevents.FindArchiveByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccArchiveConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_archive" "test" {
  name             = %[1]q
  event_source_arn = aws_cloudwatch_event_bus.test.arn
}
`, name)
}

func testAccArchiveConfig_updateAttributes(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_archive" "test" {
  name             = %[1]q
  event_source_arn = aws_cloudwatch_event_bus.test.arn
  retention_days   = 7
  description      = "test"
  event_pattern    = <<PATTERN
{
  "source": ["company.team.service"]
}
PATTERN
}
`, name)
}

func testAccArchiveConfig_kmsKeyIdentifier(name, kmsKeyIdentifier string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_kms_key" "test_1" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "key-policy-example"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow describing of the key"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        },
        Action = [
          "kms:DescribeKey"
        ],
        Resource = "*"
      },
      {
        Sid    = "Allow use of the key"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        },
        Action = [
          "kms:GenerateDataKey",
          "kms:Decrypt",
          "kms:ReEncrypt*"
        ],
        Resource = "*"
        Condition = {
          StringEquals = {
            "kms:EncryptionContext:aws:events:event-bus:arn" = aws_cloudwatch_event_bus.test.arn
          }
        }
      }
    ]
  })
  tags = {
    EventBridgeApiDestinations = "true"
  }
}

resource "aws_kms_alias" "test_1" {
  name          = "alias/test-1"
  target_key_id = aws_kms_key.test_1.key_id
}

resource "aws_kms_key" "test_2" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "key-policy-example"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow describing of the key"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        },
        Action = [
          "kms:DescribeKey"
        ],
        Resource = "*"
      },
      {
        Sid    = "Allow use of the key"
        Effect = "Allow"
        Principal = {
          Service = "events.amazonaws.com"
        },
        Action = [
          "kms:GenerateDataKey",
          "kms:Decrypt",
          "kms:ReEncrypt*"
        ],
        Resource = "*"
        Condition = {
          StringEquals = {
            "kms:EncryptionContext:aws:events:event-bus:arn" = aws_cloudwatch_event_bus.test.arn
          }
        }
      }
    ]
  })
  tags = {
    EventBridgeApiDestinations = "true"
  }
}

resource "aws_cloudwatch_event_archive" "test" {
  name               = %[1]q
  event_source_arn   = aws_cloudwatch_event_bus.test.arn
  kms_key_identifier = %[2]q
}
`, name, kmsKeyIdentifier)
}

func testAccArchiveConfig_retentionOnCreation(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_archive" "test" {
  name             = %[1]q
  event_source_arn = aws_cloudwatch_event_bus.test.arn
  retention_days   = 1
}
`, name)
}
