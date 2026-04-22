// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketAccelerateConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAccelerateConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationConfig_basic(bucketName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, names.AttrExpectedBucketOwner, ""),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.BucketAccelerateStatusEnabled)),
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

func TestAccS3BucketAccelerateConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAccelerateConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationConfig_basic(bucketName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.BucketAccelerateStatusEnabled)),
				),
			},
			{
				Config: testAccBucketAccelerateConfigurationConfig_basic(bucketName, string(types.BucketAccelerateStatusSuspended)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.BucketAccelerateStatusSuspended)),
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

func TestAccS3BucketAccelerateConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAccelerateConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationConfig_basic(bucketName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3.ResourceBucketAccelerateConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketAccelerateConfiguration_migrate_noChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAccelerateConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_acceleration(rName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acceleration_status", string(types.BucketAccelerateStatusEnabled)),
				),
			},
			{
				Config: testAccBucketAccelerateConfigurationConfig_basic(rName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.BucketAccelerateStatusEnabled)),
				),
			},
		},
	})
}

func TestAccS3BucketAccelerateConfiguration_migrate_withChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAccelerateConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_acceleration(rName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acceleration_status", string(types.BucketAccelerateStatusEnabled)),
				),
			},
			{
				Config: testAccBucketAccelerateConfigurationConfig_basic(rName, string(types.BucketAccelerateStatusSuspended)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.BucketAccelerateStatusSuspended)),
				),
			},
		},
	})
}

func TestAccS3BucketAccelerateConfiguration_expectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAccelerateConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationConfig_expectedBucketOwner(bucketName, string(types.BucketAccelerateStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrExpectedBucketOwner),
					acctest.CheckResourceAttrFormat(ctx, resourceName, names.AttrID, "{bucket},{expected_bucket_owner}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.BucketAccelerateStatusEnabled)),
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

func TestAccS3BucketAccelerateConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketAccelerateConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketAccelerateConfigurationConfig_directoryBucket(bucketName, string(types.BucketAccelerateStatusEnabled)),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketAccelerateConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)

			if rs.Type != "aws_s3_bucket_accelerate_configuration" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			if tfs3.IsDirectoryBucket(bucket) {
				conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
			}

			_, err = tfs3.FindBucketAccelerateConfiguration(ctx, conn, bucket, expectedBucketOwner)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Accelerate Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketAccelerateConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		if tfs3.IsDirectoryBucket(bucket) {
			conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
		}

		_, err = tfs3.FindBucketAccelerateConfiguration(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketAccelerateConfigurationConfig_basic(bucketName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_accelerate_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  status = %[2]q
}
`, bucketName, status)
}

func testAccBucketAccelerateConfigurationConfig_expectedBucketOwner(bucketName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_accelerate_configuration" "test" {
  bucket                = aws_s3_bucket.test.bucket
  expected_bucket_owner = data.aws_caller_identity.current.account_id

  status = %[2]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_caller_identity" "current" {}
`, bucketName, status)
}

func testAccBucketAccelerateConfigurationConfig_directoryBucket(bucketName, status string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(bucketName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_accelerate_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  status = %[1]q
}
`, status))
}
