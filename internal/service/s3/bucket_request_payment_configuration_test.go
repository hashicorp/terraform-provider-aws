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

func TestAccS3BucketRequestPaymentConfiguration_Basic_BucketOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketRequestPaymentConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, string(types.PayerBucketOwner)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "payer", string(types.PayerBucketOwner)),
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

func TestAccS3BucketRequestPaymentConfiguration_Basic_Requester(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketRequestPaymentConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, string(types.PayerRequester)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "payer", string(types.PayerRequester)),
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

func TestAccS3BucketRequestPaymentConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketRequestPaymentConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, string(types.PayerRequester)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "payer", string(types.PayerRequester)),
				),
			},
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, string(types.PayerBucketOwner)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "payer", string(types.PayerBucketOwner)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, string(types.PayerRequester)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "payer", string(types.PayerRequester)),
				),
			},
		},
	})
}

func TestAccS3BucketRequestPaymentConfiguration_migrate_noChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketRequestPaymentConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_requestPayer(rName, string(types.PayerRequester)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "request_payer", string(types.PayerRequester)),
				),
			},
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, string(types.PayerRequester)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "payer", string(types.PayerRequester)),
				),
			},
		},
	})
}

func TestAccS3BucketRequestPaymentConfiguration_migrate_withChange(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketRequestPaymentConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_requestPayer(rName, string(types.PayerRequester)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "request_payer", string(types.PayerRequester)),
				),
			},
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, string(types.PayerBucketOwner)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "payer", string(types.PayerBucketOwner)),
				),
			},
		},
	})
}

func TestAccS3BucketRequestPaymentConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketRequestPaymentConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketRequestPaymentConfigurationConfig_directoryBucket(rName, string(types.PayerBucketOwner)),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketRequestPaymentConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_request_payment_configuration" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindBucketRequestPayment(ctx, conn, bucket, expectedBucketOwner)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Request Payment Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketRequestPaymentConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
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

		_, err = tfs3.FindBucketRequestPayment(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketRequestPaymentConfigurationConfig_basic(rName, payer string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_request_payment_configuration" "test" {
  bucket = aws_s3_bucket.test.id
  payer  = %[2]q
}
`, rName, payer)
}

func testAccBucketRequestPaymentConfigurationConfig_directoryBucket(rName, payer string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_request_payment_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  payer  = %[1]q
}
`, payer))
}
