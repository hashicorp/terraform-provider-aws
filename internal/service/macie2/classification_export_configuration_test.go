// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccClassificationExportConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetClassificationExportConfigurationOutput
	resourceName := "aws_macie2_classification_export_configuration.test"
	kmsKeyResourceName := "aws_kms_key.test"
	macieAccountResourceName := "aws_macie2_account.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClassificationExportConfigurationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccClassificationExportConfigurationConfig_basic("macieprefix/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassificationExportConfigurationExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(macieAccountResourceName, names.AttrStatus, macie2.MacieStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "s3_destination.0.bucket_name", s3BucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "s3_destination.0.key_prefix", "macieprefix/"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_destination.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClassificationExportConfigurationConfig_basic(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClassificationExportConfigurationExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(macieAccountResourceName, names.AttrStatus, macie2.MacieStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "s3_destination.0.bucket_name", s3BucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "s3_destination.0.key_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "s3_destination.0.kms_key_arn", kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckClassificationExportConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie2_classification_export_configuration" {
				continue
			}

			input := macie2.GetClassificationExportConfigurationInput{}
			resp, err := conn.GetClassificationExportConfigurationWithContext(ctx, &input)

			if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException, "Macie is not enabled") {
				continue
			}

			if err != nil {
				return err
			}

			if (macie2.GetClassificationExportConfigurationOutput{}) != *resp || resp != nil { // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-conditional
				return fmt.Errorf("macie classification export configuration %q still configured", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckClassificationExportConfigurationExists(ctx context.Context, resourceName string, macie2CEConfig *macie2.GetClassificationExportConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn(ctx)
		input := macie2.GetClassificationExportConfigurationInput{}

		resp, err := conn.GetClassificationExportConfigurationWithContext(ctx, &input)

		if err != nil {
			return err
		}

		if (macie2.GetClassificationExportConfigurationOutput{}) == *resp || resp == nil { // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-conditional
			return fmt.Errorf("macie classification export configuration %q does not exist", rs.Primary.ID)
		}

		*macie2CEConfig = *resp

		return nil
	}
}

func testAccClassificationExportConfigurationConfig_basic(prefix string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Id" : "allow_macie",
    "Statement" : [
      {
        "Sid" : "Allow Macie to use the key",
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "macie.${data.aws_partition.current.dns_suffix}"
        },
        "Action" : [
          "kms:GenerateDataKey",
          "kms:Encrypt"
        ],
        "Resource" : "*"
      },
      {
        "Sid" : "Enable IAM User Permissions",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action" : "kms:*",
        "Resource" : "*"
      }
    ]
  })
}

resource "aws_s3_bucket" "test" {
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = jsonencode(
    {
      "Version" : "2012-10-17",
      "Statement" : [
        {
          "Sid" : "Deny non-HTTPS access",
          "Effect" : "Deny",
          "Principal" : "*",
          "Action" : "s3:*",
          "Resource" : "${aws_s3_bucket.test.arn}/*",
          "Condition" : {
            "Bool" : {
              "aws:SecureTransport" : "false"
            }
          }
        },
        {
          "Sid" : "Allow Macie to upload objects to the bucket",
          "Effect" : "Allow",
          "Principal" : {
            "Service" : "macie.${data.aws_partition.current.dns_suffix}"
          },
          "Action" : "s3:PutObject",
          "Resource" : "${aws_s3_bucket.test.arn}/*"
        },
        {
          "Sid" : "Allow Macie to use the getBucketLocation operation",
          "Effect" : "Allow",
          "Principal" : {
            "Service" : "macie.${data.aws_partition.current.dns_suffix}"
          },
          "Action" : "s3:GetBucketLocation",
          "Resource" : aws_s3_bucket.test.arn,
          "Condition" : {
            "StringEquals" : {
              "aws:SourceAccount" : data.aws_caller_identity.current.account_id
            },
            "ArnLike" : {
              "aws:SourceArn" : [
                "arn:${data.aws_partition.current.partition}:macie2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:export-configuration:*",
                "arn:${data.aws_partition.current.partition}:macie2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:classification-job/*"
              ]
            }
          }
        }
      ]
    }
  )
}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_classification_export_configuration" "test" {
  depends_on = [
    aws_macie2_account.test,
    aws_kms_key.test,
    aws_s3_bucket.test,
    aws_s3_bucket_policy.test
  ]
  s3_destination {
    bucket_name = aws_s3_bucket.test.bucket
    key_prefix  = (%[1]q == "") ? null : %[1]q
    kms_key_arn = aws_kms_key.test.arn
  }
}
`, prefix)
}
