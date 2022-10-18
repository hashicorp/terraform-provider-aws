package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccInstanceStorageConfigDataSource_S3Config(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance_storage_config.test"
	datasourceName := "data.aws_connect_instance_storage_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStorageConfigDataSourceConfig_S3Config(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "association_id", resourceName, "association_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "resource_type", resourceName, "resource_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.#", resourceName, "storage_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.#", resourceName, "storage_config.0.s3_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.bucket_name", resourceName, "storage_config.0.s3_config.0.bucket_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.bucket_prefix", resourceName, "storage_config.0.s3_config.0.bucket_prefix"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.encryption_config.#", resourceName, "storage_config.0.s3_config.0.encryption_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.encryption_config.0.encryption_type", resourceName, "storage_config.0.s3_config.0.encryption_config.0.encryption_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.encryption_config.0.key_id", resourceName, "storage_config.0.s3_config.0.encryption_config.0.key_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.storage_type", resourceName, "storage_config.0.storage_type"),
				),
			},
		},
	})
}

func testAccInstanceStorageConfigDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccInstanceStorageConfigDataSourceConfig_S3Config(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageConfigDataSourceConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "KMS Key for Bucket"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_connect_instance_storage_config" "test" {
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.test]

  instance_id   = aws_connect_instance.test.id
  resource_type = "CHAT_TRANSCRIPTS"

  storage_config {
    s3_config {
      bucket_name   = aws_s3_bucket.test.id
      bucket_prefix = "tf-test-Chat-Transcripts"

      encryption_config {
        encryption_type = "KMS"
        key_id          = aws_kms_key.test.arn
      }
    }
    storage_type = "S3"
  }
}

data "aws_connect_instance_storage_config" "test" {
  association_id = aws_connect_instance_storage_config.test.association_id
  instance_id    = aws_connect_instance.test.id
  resource_type  = aws_connect_instance_storage_config.test.resource_type
}
`, rName2))
}
