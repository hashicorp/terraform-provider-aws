// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDestinationARN, bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "destination_type", "S3")),
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

func testAccPublishingDestinationConfig_basic(bucketName string) string {
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

resource "aws_guardduty_publishing_destination" "test" {
  detector_id     = aws_guardduty_detector.test_gd.id
  destination_arn = aws_s3_bucket.gd_bucket.arn
  kms_key_arn     = aws_kms_key.gd_key.arn

  depends_on = [
    aws_s3_bucket_policy.gd_bucket_policy,
  ]
}`, bucketName)
}

func testAccCheckPublishingDestinationExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		destination_id, detector_id, err_state_read := tfguardduty.DecodePublishDestinationID(rs.Primary.ID)

		if err_state_read != nil {
			return err_state_read
		}

		input := &guardduty.DescribePublishingDestinationInput{
			DetectorId:    aws.String(detector_id),
			DestinationId: aws.String(destination_id),
		}

		conn := acctest.ProviderMeta(ctx, t).GuardDutyClient(ctx)
		_, err := conn.DescribePublishingDestination(ctx, input)
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

			destination_id, detector_id, err_state_read := tfguardduty.DecodePublishDestinationID(rs.Primary.ID)

			if err_state_read != nil {
				return err_state_read
			}

			input := &guardduty.DescribePublishingDestinationInput{
				DetectorId:    aws.String(detector_id),
				DestinationId: aws.String(destination_id),
			}

			_, err := conn.DescribePublishingDestination(ctx, input)
			// Catch expected error.
			if err == nil {
				return fmt.Errorf("Resource still exists.")
			}
		}
		return nil
	}
}
