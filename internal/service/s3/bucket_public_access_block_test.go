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

func TestAccS3BucketPublicAccessBlock_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config types.PublicAccessBlockConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_public_access_block.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPublicAccessBlockDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, t, "aws_s3_bucket.test"),
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", acctest.CtFalse),
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

func TestAccS3BucketPublicAccessBlock_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var config types.PublicAccessBlockConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_public_access_block.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPublicAccessBlockDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3.ResourceBucketPublicAccessBlock(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_Disappears_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	var config types.PublicAccessBlockConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_public_access_block.test"
	bucketResourceName := "aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPublicAccessBlockDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					acctest.CheckSDKResourceDisappears(ctx, t, tfs3.ResourceBucket(), bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_blockPublicACLs(t *testing.T) {
	ctx := acctest.Context(t)
	var config types.PublicAccessBlockConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_public_access_block.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPublicAccessBlockDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, true, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, true, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_blockPublicPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var config types.PublicAccessBlockConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_public_access_block.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPublicAccessBlockDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, true, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, true, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_ignorePublicACLs(t *testing.T) {
	ctx := acctest.Context(t)
	var config types.PublicAccessBlockConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_public_access_block.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPublicAccessBlockDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_restrictPublicBuckets(t *testing.T) {
	ctx := acctest.Context(t)
	var config types.PublicAccessBlockConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_public_access_block.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPublicAccessBlockDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig_basic(rName, false, false, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", acctest.CtTrue),
				),
			},
		},
	})
}

// This test can be safely run at all times as the dangling public access
// block left behind by skipped destruction will ultimately be cleaned up
// by destruction of the associated bucket.
func TestAccS3BucketPublicAccessBlock_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var config types.PublicAccessBlockConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_public_access_block.test"
	bucketResourceName := "aws_s3_bucket.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketPublicAccessBlockDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig_skipDestroy(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig_skipDestroy(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					testAccCheckBucketPublicAccessBlockExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
			// Remove the public access block resource from configuration
			{
				Config: testAccBucketPublicAccessBlockConfig_skipDestroy_postRemoval(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, t, bucketResourceName),
					testAccCheckBucketPublicAccessBlockExistsByName(ctx, t, rName),
				),
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	name := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt(t))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketPublicAccessBlockConfig_directoryBucket(name, acctest.CtFalse, acctest.CtFalse, acctest.CtFalse, acctest.CtFalse),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketPublicAccessBlockDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)

			if rs.Type != "aws_s3_bucket_public_access_block" {
				continue
			}

			if tfs3.IsDirectoryBucket(rs.Primary.ID) {
				conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
			}

			_, err := tfs3.FindPublicAccessBlockConfiguration(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Public Access Block %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketPublicAccessBlockExists(ctx context.Context, t *testing.T, n string, v *types.PublicAccessBlockConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)
		if tfs3.IsDirectoryBucket(rs.Primary.ID) {
			conn = acctest.ProviderMeta(ctx, t).S3ExpressClient(ctx)
		}

		output, err := tfs3.FindPublicAccessBlockConfiguration(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccCheckProvisionedConcurrencyConfigExistsByName is a helper to verify a
// public access block is in place for a given bucket.
//
// This variant of the test check exists function which accepts bucket name
// directly to support skip_destroy checks where the public access block
// resource is removed from state, but should still exist remotely.
func testAccCheckBucketPublicAccessBlockExistsByName(ctx context.Context, t *testing.T, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).S3Client(ctx)

		_, err := tfs3.FindPublicAccessBlockConfiguration(ctx, conn, rName)
		return err
	}
}

func testAccBucketPublicAccessBlockConfig_basic(bucketName string, blockPublicAcls, blockPublicPolicy, ignorePublicAcls, restrictPublicBuckets bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.bucket

  block_public_acls       = %[2]t
  block_public_policy     = %[3]t
  ignore_public_acls      = %[4]t
  restrict_public_buckets = %[5]t
}
`, bucketName, blockPublicAcls, blockPublicPolicy, ignorePublicAcls, restrictPublicBuckets)
}

func testAccBucketPublicAccessBlockConfig_skipDestroy(bucketName string, skipDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.bucket

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true

  skip_destroy = %[2]t
}
`, bucketName, skipDestroy)
}

func testAccBucketPublicAccessBlockConfig_skipDestroy_postRemoval(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, bucketName)
}

func testAccBucketPublicAccessBlockConfig_directoryBucket(bucketName, blockPublicAcls, blockPublicPolicy, ignorePublicAcls, restrictPublicBuckets string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_baseAZ(bucketName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket
  location {
    name = local.location_name
  }
}
resource "aws_s3_bucket_public_access_block" "bucket" {
  bucket                  = aws_s3_directory_bucket.test.bucket
  block_public_acls       = %[1]q
  block_public_policy     = %[2]q
  ignore_public_acls      = %[3]q
  restrict_public_buckets = %[4]q
}
`, blockPublicAcls, blockPublicPolicy, ignorePublicAcls, restrictPublicBuckets))
}
