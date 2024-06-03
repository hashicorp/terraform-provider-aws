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

func TestAccS3BucketInventory_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.InventoryConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_inventory.test"
	inventoryName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketInventoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketInventoryConfig_basic(rName, inventoryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketInventoryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrBucket, rName),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", "documents/"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, inventoryName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "included_object_versions", "All"),

					resource.TestCheckResourceAttr(resourceName, "optional_fields.#", acctest.Ct2),

					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.frequency", "Weekly"),

					resource.TestCheckResourceAttr(resourceName, "destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.#", acctest.Ct1),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "destination.0.bucket.0.bucket_arn", "s3", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "destination.0.bucket.0.account_id"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.0.format", "ORC"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.0.prefix", "inventory"),
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

func TestAccS3BucketInventory_encryptWithSSES3(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.InventoryConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_inventory.test"
	inventoryName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketInventoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketInventoryConfig_encryptSSE(rName, inventoryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketInventoryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.0.encryption.0.sse_s3.#", acctest.Ct1),
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

func TestAccS3BucketInventory_encryptWithSSEKMS(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.InventoryConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_inventory.test"
	inventoryName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketInventoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketInventoryConfig_encryptSSEKMS(rName, inventoryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketInventoryExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.0.encryption.0.sse_kms.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "destination.0.bucket.0.encryption.0.sse_kms.0.key_id", regexache.MustCompile(fmt.Sprintf("^arn:%s:kms:", acctest.Partition()))),
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

func TestAccS3BucketInventory_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	inventoryName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketInventoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketInventoryConfig_directoryBucket(rName, inventoryName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketInventoryExists(ctx context.Context, n string, v *types.InventoryConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		bucket, name, err := tfs3.BucketInventoryParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		output, err := tfs3.FindInventoryConfiguration(ctx, conn, bucket, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBucketInventoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_inventory" {
				continue
			}

			bucket, name, err := tfs3.BucketInventoryParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindInventoryConfiguration(ctx, conn, bucket, name)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Inventory %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketInventoryConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketInventoryConfig_basic(bucketName, inventoryName string) string {
	return acctest.ConfigCompose(testAccBucketInventoryConfig_base(bucketName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket_inventory" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[1]q

  included_object_versions = "All"

  optional_fields = [
    "Size",
    "LastModifiedDate",
  ]

  filter {
    prefix = "documents/"
  }

  schedule {
    frequency = "Weekly"
  }

  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_bucket.test.arn
      account_id = data.aws_caller_identity.current.account_id
      prefix     = "inventory"
    }
  }
}
`, inventoryName))
}

func testAccBucketInventoryConfig_encryptSSE(bucketName, inventoryName string) string {
	return acctest.ConfigCompose(testAccBucketInventoryConfig_base(bucketName), fmt.Sprintf(`
resource "aws_s3_bucket_inventory" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[1]q

  included_object_versions = "Current"

  schedule {
    frequency = "Daily"
  }

  destination {
    bucket {
      format     = "CSV"
      bucket_arn = aws_s3_bucket.test.arn

      encryption {
        sse_s3 {}
      }
    }
  }
}
`, inventoryName))
}

func testAccBucketInventoryConfig_encryptSSEKMS(bucketName, inventoryName string) string {
	return acctest.ConfigCompose(testAccBucketInventoryConfig_base(bucketName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_s3_bucket_inventory" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[2]q

  included_object_versions = "Current"

  schedule {
    frequency = "Daily"
  }

  destination {
    bucket {
      format     = "Parquet"
      bucket_arn = aws_s3_bucket.test.arn

      encryption {
        sse_kms {
          key_id = aws_kms_key.test.arn
        }
      }
    }
  }
}
`, bucketName, inventoryName))
}

func testAccBucketInventoryConfig_directoryBucket(bucketName, inventoryName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(bucketName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_inventory" "test" {
  bucket = aws_s3_directory_bucket.test.id
  name   = %[1]q

  included_object_versions = "All"

  optional_fields = [
    "Size",
    "LastModifiedDate",
  ]

  filter {
    prefix = "documents/"
  }

  schedule {
    frequency = "Weekly"
  }

  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_directory_bucket.test.arn
      account_id = data.aws_caller_identity.current.account_id
      prefix     = "inventory"
    }
  }
}
`, inventoryName))
}
