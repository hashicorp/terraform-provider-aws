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

func TestAccS3BucketVersioning_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusEnabled)),
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

func TestAccS3BucketVersioning_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketVersioning(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketVersioning_disappears_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucket(), bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketVersioning_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
				),
			},
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusSuspended)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusSuspended)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusEnabled)),
				),
			},
		},
	})
}

// TestAccBucketVersioning_MFADelete can only test for a "Disabled"
// mfa_delete configuration as the "mfa" argument is required if it's enabled
func TestAccS3BucketVersioning_MFADelete(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_mfaDelete(rName, string(types.MFADeleteDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.mfa_delete", string(types.MFADeleteDisabled)),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusEnabled)),
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

func TestAccS3BucketVersioning_migrate_versioningDisabledNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateEnabled(bucketName, tfs3.BucketVersioningStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", tfs3.BucketVersioningStatusDisabled),
				),
			},
		},
	})
}

func TestAccS3BucketVersioning_migrate_versioningDisabledWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateEnabled(bucketName, string(types.BucketVersioningStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusEnabled)),
				),
			},
		},
	})
}

func TestAccS3BucketVersioning_migrate_versioningEnabledNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateEnabled(bucketName, string(types.BucketVersioningStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusEnabled)),
				),
			},
		},
	})
}

func TestAccS3BucketVersioning_migrate_versioningEnabledWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateEnabled(bucketName, string(types.BucketVersioningStatusSuspended)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusSuspended)),
				),
			},
		},
	})
}

// TestAccS3BucketVersioning_migrate_mfaDeleteNoChange can only test for a "Disabled"
// mfa_delete configuration as the "mfa" argument is required if it's enabled
func TestAccS3BucketVersioning_migrate_mfaDeleteNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioningMFADelete(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.mfa_delete", acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateMFADelete(bucketName, string(types.MFADeleteDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.mfa_delete", string(types.MFADeleteDisabled)),
				),
			},
		},
	})
}

func TestAccS3BucketVersioning_Status_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", tfs3.BucketVersioningStatusDisabled),
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

func TestAccS3BucketVersioning_Status_disabledToEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", tfs3.BucketVersioningStatusDisabled),
				),
			},
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusEnabled)),
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

func TestAccS3BucketVersioning_Status_disabledToSuspended(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", tfs3.BucketVersioningStatusDisabled),
				),
			},
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusSuspended)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusSuspended)),
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

func TestAccS3BucketVersioning_Status_enabledToDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusEnabled)),
				),
			},
			{
				Config:      testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				ExpectError: regexache.MustCompile(`versioning_configuration.status cannot be updated from 'Enabled' to 'Disabled'`),
			},
		},
	})
}

func TestAccS3BucketVersioning_Status_suspendedToDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, string(types.BucketVersioningStatusSuspended)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", string(types.BucketVersioningStatusSuspended)),
				),
			},
			{
				Config:      testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				ExpectError: regexache.MustCompile(`versioning_configuration.status cannot be updated from 'Suspended' to 'Disabled'`),
			},
		},
	})
}

func TestAccS3BucketVersioning_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketVersioningDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketVersioningConfig_directoryBucket(rName, string(types.BucketVersioningStatusEnabled)),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketVersioningDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_versioning" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindBucketVersioning(ctx, conn, bucket, expectedBucketOwner)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Versioning %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketVersioningExists(ctx context.Context, n string) resource.TestCheckFunc {
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

		_, err = tfs3.FindBucketVersioning(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketVersioningConfig_basic(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = %[2]q
  }
}
`, rName, status)
}

func testAccBucketVersioningConfig_mfaDelete(rName, mfaDelete string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    mfa_delete = %[2]q
    status     = "Enabled"
  }
}
`, rName, mfaDelete)
}

func testAccBucketVersioningConfig_migrateEnabled(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = %[2]q
  }
}
`, rName, status)
}

func testAccBucketVersioningConfig_migrateMFADelete(rName, mfaDelete string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    mfa_delete = %[2]q
    status     = "Enabled"
  }
}
`, rName, mfaDelete)
}

func testAccBucketVersioningConfig_directoryBucket(rName, status string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  versioning_configuration {
    status = %[1]q
  }
}
`, status))
}
