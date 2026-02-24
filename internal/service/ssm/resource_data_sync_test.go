// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMResourceDataSync_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssm_resource_data_sync.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDataSyncDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSyncConfig_basic(acctest.RandInt(t), sdkacctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
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

func TestAccSSMResourceDataSync_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssm_resource_data_sync.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDataSyncDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSyncConfig_basic(acctest.RandInt(t), sdkacctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfssm.ResourceResourceDataSync(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMResourceDataSync_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(5)
	resourceName := "aws_ssm_resource_data_sync.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDataSyncDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSyncConfig_basic(acctest.RandInt(t), rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceDataSyncConfig_update(acctest.RandInt(t), rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceDataSyncExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func testAccCheckResourceDataSyncDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_resource_data_sync" {
				continue
			}

			_, err := tfssm.FindResourceDataSyncByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSM Resource Data Sync %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourceDataSyncExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)

		_, err := tfssm.FindResourceDataSyncByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccResourceDataSyncConfig_basic(rInt int, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "hoge" {
  bucket        = "tf-test-bucket-%[1]d"
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "hoge" {
  bucket = aws_s3_bucket.hoge.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SSMBucketPermissionsCheck",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-bucket-%[1]d"
    },
    {
      "Sid": " SSMBucketDelivery",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:PutObject",
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::tf-test-bucket-%[1]d/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
      EOF

}

resource "aws_ssm_resource_data_sync" "test" {
  name = "tf-test-ssm-%[2]s"

  s3_destination {
    bucket_name = aws_s3_bucket.hoge.bucket
    region      = aws_s3_bucket.hoge.region
  }
}
`, rInt, rName)
}

func testAccResourceDataSyncConfig_update(rInt int, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "hoge" {
  bucket        = "tf-test-bucket-%[1]d"
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "hoge" {
  bucket = aws_s3_bucket.hoge.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SSMBucketPermissionsCheck",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-bucket-%[1]d"
    },
    {
      "Sid": " SSMBucketDelivery",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:PutObject",
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::tf-test-bucket-%[1]d/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
      EOF

}

resource "aws_ssm_resource_data_sync" "test" {
  name = "tf-test-ssm-%[2]s"

  s3_destination {
    bucket_name = aws_s3_bucket.hoge.bucket
    region      = aws_s3_bucket.hoge.region
    prefix      = "test-"
  }
}
`, rInt, rName)
}
