// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketCORSConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCORSConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": acctest.Ct1,
						"allowed_origins.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
				},
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCORSConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketCorsConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx),
		Steps: []resource.TestStep{

			{
				Config: testAccBucketCORSConfigurationConfig_completeSingleRule(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": acctest.Ct1,
						"allowed_methods.#": acctest.Ct3,
						"allowed_origins.#": acctest.Ct1,
						"expose_headers.#":  acctest.Ct1,
						names.AttrID:        rName,
						"max_age_seconds":   "3000",
					}),
				),
			},
			{
				Config: testAccBucketCORSConfigurationConfig_multipleRules(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": acctest.Ct1,
						"allowed_methods.#": acctest.Ct3,
						"allowed_origins.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": acctest.Ct1,
						"allowed_origins.#": acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
					"cors_rule.1.max_age_seconds",
				},
			},
			{
				Config: testAccBucketCORSConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": acctest.Ct1,
						"allowed_origins.#": acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
				},
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_SingleRule(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCORSConfigurationConfig_completeSingleRule(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": acctest.Ct1,
						"allowed_methods.#": acctest.Ct3,
						"allowed_origins.#": acctest.Ct1,
						"expose_headers.#":  acctest.Ct1,
						names.AttrID:        rName,
						"max_age_seconds":   "3000",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_headers.*", "*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.expose_headers.*", "ETag"),
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

func TestAccS3BucketCORSConfiguration_MultipleRules(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCORSConfigurationConfig_multipleRules(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": acctest.Ct1,
						"allowed_methods.#": acctest.Ct3,
						"allowed_origins.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_headers.*", "*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": acctest.Ct1,
						"allowed_origins.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "*"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cors_rule.0.max_age_seconds",
					"cors_rule.1.max_age_seconds",
				},
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_migrate_corsRuleNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_cors(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_methods.#", acctest.Ct2),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_origins.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.expose_headers.#", acctest.Ct2),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
			{
				Config: testAccBucketCORSConfigurationConfig_migrateRuleNoChange(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": acctest.Ct1,
						"allowed_methods.#": acctest.Ct2,
						"allowed_origins.#": acctest.Ct1,
						"expose_headers.#":  acctest.Ct2,
						"max_age_seconds":   "3000",
					}),
				),
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_migrate_corsRuleWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_cors(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_headers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_methods.#", acctest.Ct2),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_origins.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.expose_headers.#", acctest.Ct2),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
			{
				Config: testAccBucketCORSConfigurationConfig_migrateRuleChange(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketCORSConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": acctest.Ct1,
						"allowed_origins.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
				),
			},
		},
	})
}

func TestAccS3BucketCORSConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketCORSConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketCORSConfigurationConfig_directoryBucket(rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketCORSConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_cors_configuration" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindCORSRules(ctx, conn, bucket, expectedBucketOwner)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Website Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketCORSConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
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

		_, err = tfs3.FindCORSRules(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketCORSConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_completeSingleRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST", "DELETE"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["ETag"]
    id              = %[1]q
    max_age_seconds = 3000
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_multipleRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST", "DELETE"]
    allowed_origins = ["https://www.example.com"]
  }

  cors_rule {
    allowed_methods = ["GET"]
    allowed_origins = ["*"]
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_migrateRuleNoChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_migrateRuleChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}
`, rName)
}

func testAccBucketCORSConfigurationConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}
`)
}
