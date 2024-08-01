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

func TestAccS3BucketLogging_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrExpectedBucketOwner, ""),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "log/"),
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

func TestAccS3BucketLogging_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketLogging(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketLogging_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
				),
			},
			{
				// Test updating "target_prefix".
				Config: testAccBucketLoggingConfig_update(rName, "tmp/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "tmp/"),
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

func TestAccS3BucketLogging_TargetGrantByID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_targetGrantByID(rName, string(types.BucketLogsPermissionFullControl)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.BucketLogsPermissionFullControl),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "target_grant.*.grantee.0.id", "data.aws_canonical_user_id.current", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "target_grant.*.grantee.0.display_name", "data.aws_canonical_user_id.current", names.AttrDisplayName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_targetGrantByID(rName, string(types.BucketLogsPermissionRead)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeCanonicalUser),
						"permission":     string(types.BucketLogsPermissionRead),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "target_grant.*.grantee.0.display_name", "data.aws_canonical_user_id.current", names.AttrDisplayName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_TargetGrantByEmail(t *testing.T) {
	ctx := acctest.Context(t)
	rEmail := acctest.SkipIfEnvVarNotSet(t, "AWS_S3_BUCKET_LOGGING_AMAZON_CUSTOMER_BY_EMAIL")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_targetGrantByEmail(rName, rEmail, string(types.BucketLogsPermissionFullControl)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":               acctest.Ct1,
						"grantee.0.email_address": rEmail,
						"grantee.0.type":          string(types.TypeAmazonCustomerByEmail),
						"permission":              string(types.BucketLogsPermissionFullControl),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_targetGrantByEmail(rName, rEmail, string(types.BucketLogsPermissionRead)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":       acctest.Ct1,
						"grantee.0.email": rEmail,
						"grantee.0.type":  string(types.TypeAmazonCustomerByEmail),
						"permission":      string(types.BucketLogsPermissionRead),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_TargetGrantByGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_targetGrantByGroup(rName, string(types.BucketLogsPermissionFullControl)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeGroup),
						"permission":     string(types.BucketLogsPermissionFullControl),
					}),
					testAccCheckBucketLoggingTargetGrantGranteeURI(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_targetGrantByGroup(rName, string(types.BucketLogsPermissionRead)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":      acctest.Ct1,
						"grantee.0.type": string(types.TypeGroup),
						"permission":     string(types.BucketLogsPermissionRead),
					}),
					testAccCheckBucketLoggingTargetGrantGranteeURI(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_migrate_loggingNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_logging(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "logging.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(bucketResourceName, "logging.0.target_bucket", "aws_s3_bucket.log_bucket", names.AttrID),
					resource.TestCheckResourceAttr(bucketResourceName, "logging.0.target_prefix", "log/"),
				),
			},
			{
				Config: testAccBucketLoggingConfig_migrate(bucketName, "log/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "log/"),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_migrate_loggingWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_logging(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "logging.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(bucketResourceName, "logging.0.target_bucket", "aws_s3_bucket.log_bucket", names.AttrID),
					resource.TestCheckResourceAttr(bucketResourceName, "logging.0.target_prefix", "log/"),
				),
			},
			{
				Config: testAccBucketLoggingConfig_migrate(bucketName, "tmp/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "tmp/"),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_withExpectedBucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_withExpectedBucketOwner(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrExpectedBucketOwner),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "log/"),
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

func TestAccS3BucketLogging_withTargetObjectKeyFormat(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_withTargetObjectKeyFormatPartitionedPrefix(rName, "EventTime"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.0.partitioned_prefix.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.0.partitioned_prefix.0.partition_date_source", "EventTime"),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.0.simple_prefix.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_withTargetObjectKeyFormatPartitionedPrefix(rName, "DeliveryTime"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.0.partitioned_prefix.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.0.partitioned_prefix.0.partition_date_source", "DeliveryTime"),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.0.simple_prefix.#", acctest.Ct0),
				),
			},
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.#", acctest.Ct0),
				),
			},
			{
				Config: testAccBucketLoggingConfig_withTargetObjectKeyFormatSimplePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.0.partitioned_prefix.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_object_key_format.0.simple_prefix.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketLoggingConfig_directoryBucket(rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketLoggingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_logging" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindLoggingEnabled(ctx, conn, bucket, expectedBucketOwner)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Logging %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketLoggingExists(ctx context.Context, n string) resource.TestCheckFunc {
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

		_, err = tfs3.FindLoggingEnabled(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccCheckBucketLoggingTargetGrantGranteeURI(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		uri := fmt.Sprintf("http://acs.%s/groups/s3/LogDelivery", acctest.PartitionDNSSuffix())
		return resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
			"grantee.0.uri": uri,
		})(s)
	}
}

func testAccBucketLoggingConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[1]s-log"
}

resource "aws_s3_bucket_ownership_controls" "log_bucket_ownership" {
  bucket = aws_s3_bucket.log_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  depends_on = [aws_s3_bucket_ownership_controls.log_bucket_ownership]

  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  depends_on = [aws_s3_bucket_acl.log_bucket_acl]
}
`, rName)
}

func testAccBucketLoggingConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), `
resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"
}
`)
}

func testAccBucketLoggingConfig_update(rName, targetPrefix string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = %[1]q
}
`, targetPrefix))
}

func testAccBucketLoggingConfig_targetGrantByID(rName, permission string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"

  target_grant {
    grantee {
      id   = data.aws_canonical_user_id.current.id
      type = "CanonicalUser"
    }
    permission = %[1]q
  }
}
`, permission))
}

func testAccBucketLoggingConfig_targetGrantByEmail(rName, email, permission string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"

  target_grant {
    grantee {
      email_address = %[1]q
      type          = "AmazonCustomerByEmail"
    }
    permission = %[2]q
  }
}
`, email, permission))
}

func testAccBucketLoggingConfig_targetGrantByGroup(rName, permission string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"

  target_grant {
    grantee {
      type = "Group"
      # Test with S3 log delivery group
      uri = "http://acs.${data.aws_partition.current.dns_suffix}/groups/s3/LogDelivery"
    }
    permission = %[1]q
  }
}
`, permission))
}

func testAccBucketLoggingConfig_migrate(rName, targetPrefix string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = %[1]q
}
`, targetPrefix))
}

func testAccBucketLoggingConfig_withExpectedBucketOwner(rName string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), `
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  expected_bucket_owner = data.aws_caller_identity.current.account_id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"
}
`)
}

func testAccBucketLoggingConfig_withTargetObjectKeyFormatPartitionedPrefix(rName, partitionDateSource string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"

  target_object_key_format {
    partitioned_prefix {
      partition_date_source = %[1]q
    }
  }
}
`, partitionDateSource))
}

func testAccBucketLoggingConfig_withTargetObjectKeyFormatSimplePrefix(rName string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), `
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"

  target_object_key_format {
    simple_prefix {}
  }
}
`)
}

func testAccBucketLoggingConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccBucketLoggingConfig_base(rName), testAccDirectoryBucketConfig_base(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket
  location {
    name = local.location_name
  }
}
resource "aws_s3_bucket_logging" "test" {
  bucket        = aws_s3_directory_bucket.test.bucket
  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"
}
`)
}
