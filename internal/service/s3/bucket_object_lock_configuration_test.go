// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketObjectLockConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectLockConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", string(types.ObjectLockRetentionModeCompliance)),
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

func TestAccS3BucketObjectLockConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectLockConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketObjectLockConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketObjectLockConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(ctx, resourceName),
				),
			},
			{
				Config: testAccBucketObjectLockConfigurationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.years", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", string(types.ObjectLockRetentionModeGovernance)),
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

func TestAccS3BucketObjectLockConfiguration_migrate_noChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectLockConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledDefaultRetention(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.rule.0.default_retention.0.mode", string(types.ObjectLockRetentionModeCompliance)),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.rule.0.default_retention.0.days", acctest.Ct3),
				),
			},
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", string(types.ObjectLockRetentionModeCompliance)),
				),
			},
		},
	})
}

func TestAccS3BucketObjectLockConfiguration_migrate_withChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectLockConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledNoDefaultRetention(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.rule.#", acctest.Ct0),
				),
			},
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", string(types.ObjectLockRetentionModeCompliance)),
				),
			},
		},
	})
}

func TestAccS3BucketObjectLockConfiguration_noRule(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectLockConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_noRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", string(types.ObjectLockEnabledEnabled)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
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

func TestAccS3BucketObjectLockConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketObjectLockConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketObjectLockConfigurationConfig_directoryBucket(rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketObjectLockConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_object_lock_configuration" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindObjectLockConfiguration(ctx, conn, bucket, expectedBucketOwner)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Object Lock configuration (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketObjectLockConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		_, err = tfs3.FindObjectLockConfiguration(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketObjectLockConfigurationConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    default_retention {
      mode = %[2]q
      days = 3
    }
  }
}
`, bucketName, types.ObjectLockRetentionModeCompliance)
}

func testAccBucketObjectLockConfigurationConfig_update(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    default_retention {
      mode  = %[2]q
      years = 1
    }
  }
}
`, bucketName, types.ObjectLockModeGovernance)
}

func testAccBucketObjectLockConfigurationConfig_noRule(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket.test.id
}
`, bucketName)
}

func testAccBucketObjectLockConfigurationConfig_directoryBucket(bucketName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket

  rule {
    default_retention {
      mode = %[1]q
      days = 3
    }
  }
}
`, types.ObjectLockRetentionModeCompliance))
}
