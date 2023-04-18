package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3BucketServerSideEncryptionConfigurationDataSource_aes256(t *testing.T) {
	ctx := acctest.Context(t)

	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket-aes256")
	dataSourceName := "data.aws_s3_bucket_server_side_encryption_configuration.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceServerSideEncryptionConfig_aes256(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "sse_algorithm", "AES256"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_key_enabled", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "kms_master_key_id", ""),
				),
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfigurationDataSource_kms(t *testing.T) {
	ctx := acctest.Context(t)

	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket-kms")
	dataSourceName := "data.aws_s3_bucket_server_side_encryption_configuration.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceServerSideEncryptionConfig_kms(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "sse_algorithm", "aws:kms"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_key_enabled", "false"),
					resource.TestCheckResourceAttrSet(dataSourceName, "kms_master_key_id"),
				),
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfigurationDataSource_default_encryption(t *testing.T) {
	ctx := acctest.Context(t)

	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket-default-encryption")
	dataSourceName := "data.aws_s3_bucket_server_side_encryption_configuration.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceServerSideEncryptionConfig_default_encryption(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "sse_algorithm", "AES256"),
					resource.TestCheckResourceAttr(dataSourceName, "bucket_key_enabled", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "kms_master_key_id", ""),
				),
			},
		},
	})
}

func testAccBucketDataSourceServerSideEncryptionConfig_aes256(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "aes256" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "AES256"  
    }
  }
}

data "aws_s3_bucket_server_side_encryption_configuration" "bucket" {
  bucket = aws_s3_bucket.bucket.id
}
`, bucketName)
}

func testAccBucketDataSourceServerSideEncryptionConfig_kms(bucketName string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "s3" {
  name = "alias/aws/s3"
}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "kms" {
  bucket = aws_s3_bucket.bucket.id
  
  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = data.aws_kms_alias.s3.id
      sse_algorithm     = "aws:kms"
    }
  }
}

data "aws_s3_bucket_server_side_encryption_configuration" "bucket" {
  bucket = aws_s3_bucket.bucket.id
}
`, bucketName)
}

func testAccBucketDataSourceServerSideEncryptionConfig_default_encryption(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

data "aws_s3_bucket_server_side_encryption_configuration" "bucket" {
  bucket = aws_s3_bucket.bucket.id
}
`, bucketName)
}
