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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketServerSideEncryptionConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.bucket_key_enabled",
				},
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySEEByDefault_AES256(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, string(types.ServerSideEncryptionAes256)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAes256)),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.bucket_key_enabled",
				},
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_KMS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, string(types.ServerSideEncryptionAwsKms)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.bucket_key_enabled",
				},
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_KMSDSSE(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, string(types.ServerSideEncryptionAwsKmsDsse)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKmsDsse)),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.bucket_key_enabled",
				},
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_UpdateSSEAlgorithm(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, string(types.ServerSideEncryptionAwsKms)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.bucket_key_enabled",
				},
			},
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, string(types.ServerSideEncryptionAes256)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAes256)),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.bucket_key_enabled",
				},
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_KMSWithMasterKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKMSMasterKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.bucket_key_enabled",
				},
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_KMSWithMasterKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKMSMasterKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "aws_kms_key.test", names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"rule.0.bucket_key_enabled",
				},
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_BucketKeyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_keyEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_keyEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", acctest.CtFalse),
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

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_BucketKeyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKeyEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "aws_kms_key.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKeyEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "aws_kms_key.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", acctest.CtFalse),
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

func TestAccS3BucketServerSideEncryptionConfiguration_migrate_noChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_defaultEncryptionDefaultKey(rName, string(types.ServerSideEncryptionAwsKms)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.bucket_key_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_migrateNoChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_migrate_withChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_defaultEncryptionDefaultKey(rName, string(types.ServerSideEncryptionAwsKms)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAwsKms)),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.bucket_key_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, string(types.ServerSideEncryptionAes256)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, bucketResourceName, names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", string(types.ServerSideEncryptionAes256)),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketServerSideEncryptionConfigurationConfig_directoryBucket(rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketServerSideEncryptionConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
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

		_, err = tfs3.FindServerSideEncryptionConfiguration(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketServerSideEncryptionConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    # This is Amazon S3 bucket default encryption.
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
`, rName)
}

func testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKMSMasterKeyARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "KMS Key for Bucket %[1]s"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
  }
}
`, rName)
}

func testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKMSMasterKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "KMS Key for Bucket %[1]s"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.id
      sse_algorithm     = "aws:kms"
    }
  }
}
`, rName)
}

func testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, sseAlgorithm string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = %[2]q
    }
  }
}
`, rName, sseAlgorithm)
}

func testAccBucketServerSideEncryptionConfigurationConfig_keyEnabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    # This is Amazon S3 bucket default encryption.
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }

    bucket_key_enabled = %[2]t
  }
}
`, rName, enabled)
}

func testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKeyEnabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "KMS Key for Bucket %[1]s"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.id
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = %[2]t
  }
}
`, rName, enabled)
}

func testAccBucketServerSideEncryptionConfigurationConfig_migrateNoChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}
`, rName)
}

func testAccBucketServerSideEncryptionConfigurationConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket

  rule {
    # This is Amazon S3 bucket default encryption.
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
`)
}
