// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPublishingDestination_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_publishing_destination.test"
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	detectorResourceName := "aws_guardduty_detector.test_gd"
	bucketResourceName := "aws_s3_bucket.gd_bucket"
	kmsKeyResourceName := "aws_kms_key.gd_key"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublishingDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublishingDestinationConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublishingDestinationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "guardduty", "detector/{detector_id}/publishingdestination/{destination_id}"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestinationARN, bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "destination_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
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

func testAccPublishingDestination_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_publishing_destination.test"
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublishingDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublishingDestinationConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublishingDestinationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfguardduty.ResourcePublishingDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPublishingDestination_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_publishing_destination.test"
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPublishingDestinationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPublishingDestinationConfig_tags1(bucketName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublishingDestinationExists(ctx, t, resourceName),
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
				Config: testAccPublishingDestinationConfig_tags2(bucketName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublishingDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPublishingDestinationConfig_tags1(bucketName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublishingDestinationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckPublishingDestinationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		_, err := tfguardduty.FindPublishingDestinationByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes["destination_id"])

		return err
	}
}

func testAccCheckPublishingDestinationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_guardduty_publishing_destination" {
				continue
			}

			_, err := tfguardduty.FindPublishingDestinationByTwoPartKey(ctx, conn, rs.Primary.Attributes["detector_id"], rs.Primary.Attributes["destination_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("GuardDuty Publishing Destination %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPublishingDestinationConfig_base(bucketName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "bucket_pol" {
  statement {
    sid = "Allow PutObject"
    actions = [
      "s3:PutObject"
    ]

    resources = [
      "${aws_s3_bucket.gd_bucket.arn}/*"
    ]

    principals {
      type        = "Service"
      identifiers = ["guardduty.${data.aws_partition.current.dns_suffix}"]
    }
  }

  statement {
    sid = "Allow GetBucketLocation"
    actions = [
      "s3:GetBucketLocation"
    ]

    resources = [
      aws_s3_bucket.gd_bucket.arn
    ]

    principals {
      type        = "Service"
      identifiers = ["guardduty.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

data "aws_iam_policy_document" "kms_pol" {
  statement {
    sid = "Allow GuardDuty to encrypt findings"
    actions = [
      "kms:GenerateDataKey"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:kms:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:key/*"
    ]

    principals {
      type        = "Service"
      identifiers = ["guardduty.${data.aws_partition.current.dns_suffix}"]
    }
  }

  statement {
    sid = "Allow all users to modify/delete key (test only)"
    actions = [
      "kms:*"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:kms:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:key/*"
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

resource "aws_guardduty_detector" "test_gd" {
  enable = true
}

resource "aws_s3_bucket" "gd_bucket" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_policy" "gd_bucket_policy" {
  bucket = aws_s3_bucket.gd_bucket.id
  policy = data.aws_iam_policy_document.bucket_pol.json
}

resource "aws_kms_key" "gd_key" {
  description             = "Temporary key for AccTest of TF"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  policy                  = data.aws_iam_policy_document.kms_pol.json
}
`, bucketName)
}

func testAccPublishingDestinationConfig_basic(bucketName string) string {
	return acctest.ConfigCompose(
		testAccPublishingDestinationConfig_base(bucketName),
		`
resource "aws_guardduty_publishing_destination" "test" {
  detector_id     = aws_guardduty_detector.test_gd.id
  destination_arn = aws_s3_bucket.gd_bucket.arn
  kms_key_arn     = aws_kms_key.gd_key.arn

  depends_on = [
    aws_s3_bucket_policy.gd_bucket_policy,
  ]
}
`)
}

func testAccPublishingDestinationConfig_tags1(bucketName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccPublishingDestinationConfig_base(bucketName),
		fmt.Sprintf(`
resource "aws_guardduty_publishing_destination" "test" {
  detector_id     = aws_guardduty_detector.test_gd.id
  destination_arn = aws_s3_bucket.gd_bucket.arn
  kms_key_arn     = aws_kms_key.gd_key.arn

  tags = {
    %[1]q = %[2]q
  }

  depends_on = [
    aws_s3_bucket_policy.gd_bucket_policy,
  ]
}
`, tagKey1, tagValue1))
}

func testAccPublishingDestinationConfig_tags2(bucketName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccPublishingDestinationConfig_base(bucketName),
		fmt.Sprintf(`
resource "aws_guardduty_publishing_destination" "test" {
  detector_id     = aws_guardduty_detector.test_gd.id
  destination_arn = aws_s3_bucket.gd_bucket.arn
  kms_key_arn     = aws_kms_key.gd_key.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  depends_on = [
    aws_s3_bucket_policy.gd_bucket_policy,
  ]
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
