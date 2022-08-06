package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnectInstanceStorageConfig_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":                     testAccInstanceStorageConfig_basic,
		"disappears":                testAccInstanceStorageConfig_disappears,
		"S3Config_BucketName":       testAccInstanceStorageConfig_S3Config_BucketName,
		"S3Config_BucketPrefix":     testAccInstanceStorageConfig_S3Config_BucketPrefix,
		"S3Config_EncryptionConfig": testAccInstanceStorageConfig_S3Config_EncryptionConfig,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccInstanceStorageConfig_basic(t *testing.T) {
	var v connect.DescribeInstanceStorageConfigOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance_storage_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceStorageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStorageConfigConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStorageConfigExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "association_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", connect.InstanceStorageResourceTypeChatTranscripts),
					resource.TestCheckResourceAttr(resourceName, "storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_config.0.s3_config.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.bucket_prefix", "tf-test-Chat-Transcripts"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.storage_type", connect.StorageTypeS3),
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

func testAccInstanceStorageConfig_S3Config_BucketName(t *testing.T) {
	var v connect.DescribeInstanceStorageConfigOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance_storage_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceStorageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStorageConfigConfig_S3Config_bucketName(rName, rName2, rName3, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStorageConfigExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_config.0.s3_config.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.bucket_prefix", "tf-test-Chat-Transcripts"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.storage_type", connect.StorageTypeS3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceStorageConfigConfig_S3Config_bucketName(rName, rName2, rName3, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStorageConfigExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_config.0.s3_config.0.bucket_name", "aws_s3_bucket.test2", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.bucket_prefix", "tf-test-Chat-Transcripts"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.storage_type", connect.StorageTypeS3),
				),
			},
		},
	})
}

func testAccInstanceStorageConfig_S3Config_BucketPrefix(t *testing.T) {
	var v connect.DescribeInstanceStorageConfigOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance_storage_config.test"

	originalBucketPrefix := "originalBucketPrefix"
	updatedBucketPrefix := "updatedBucketPrefix"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceStorageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStorageConfigConfig_S3Config_bucketPrefix(rName, rName2, originalBucketPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStorageConfigExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_config.0.s3_config.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.bucket_prefix", originalBucketPrefix),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.storage_type", connect.StorageTypeS3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceStorageConfigConfig_S3Config_bucketPrefix(rName, rName2, updatedBucketPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStorageConfigExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_config.0.s3_config.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.bucket_prefix", updatedBucketPrefix),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.storage_type", connect.StorageTypeS3),
				),
			},
		},
	})
}

func testAccInstanceStorageConfig_S3Config_EncryptionConfig(t *testing.T) {
	var v connect.DescribeInstanceStorageConfigOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance_storage_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceStorageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStorageConfigConfig_S3Config_encryptionConfig(rName, rName2, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStorageConfigExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_config.0.s3_config.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.bucket_prefix", "tf-test-Chat-Transcripts"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.encryption_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.encryption_config.0.encryption_type", connect.EncryptionTypeKms),
					resource.TestCheckResourceAttrPair(resourceName, "storage_config.0.s3_config.0.encryption_config.0.key_id", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.storage_type", connect.StorageTypeS3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceStorageConfigConfig_S3Config_encryptionConfig(rName, rName2, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStorageConfigExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "storage_config.0.s3_config.0.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.bucket_prefix", "tf-test-Chat-Transcripts"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.encryption_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.s3_config.0.encryption_config.0.encryption_type", connect.EncryptionTypeKms),
					resource.TestCheckResourceAttrPair(resourceName, "storage_config.0.s3_config.0.encryption_config.0.key_id", "aws_kms_key.test2", "arn"),
					resource.TestCheckResourceAttr(resourceName, "storage_config.0.storage_type", connect.StorageTypeS3),
				),
			},
		},
	})
}

func testAccInstanceStorageConfig_disappears(t *testing.T) {
	var v connect.DescribeInstanceStorageConfigOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance_storage_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceStorageConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStorageConfigConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceStorageConfigExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceInstanceStorageConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInstanceStorageConfigExists(resourceName string, function *connect.DescribeInstanceStorageConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Instance Storage Config not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Instance Storage Config ID not set")
		}
		instanceId, associationId, resourceType, err := tfconnect.InstanceStorageConfigParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeInstanceStorageConfigInput{
			AssociationId: aws.String(associationId),
			InstanceId:    aws.String(instanceId),
			ResourceType:  aws.String(resourceType),
		}

		getFunction, err := conn.DescribeInstanceStorageConfig(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckInstanceStorageConfigDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_instance_storage_config" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceId, associationId, resourceType, err := tfconnect.InstanceStorageConfigParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeInstanceStorageConfigInput{
			AssociationId: aws.String(associationId),
			InstanceId:    aws.String(instanceId),
			ResourceType:  aws.String(resourceType),
		}

		_, err = conn.DescribeInstanceStorageConfig(params)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccInstanceStorageConfigConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccInstanceStorageConfigConfig_basic(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageConfigConfig_base(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_connect_instance_storage_config" "test" {
  instance_id   = aws_connect_instance.test.id
  resource_type = "CHAT_TRANSCRIPTS"

  storage_config {
    s3_config {
      bucket_name   = aws_s3_bucket.test.id
      bucket_prefix = "tf-test-Chat-Transcripts"
    }
    storage_type = "S3"
  }
}
`, rName2))
}

func testAccInstanceStorageConfigConfig_S3Config_bucketName(rName, rName2, rName3, selectBucket string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageConfigConfig_base(rName),
		fmt.Sprintf(`
locals {
  select_bucket = %[3]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket" "test2" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_connect_instance_storage_config" "test" {
  instance_id   = aws_connect_instance.test.id
  resource_type = "CHAT_TRANSCRIPTS"

  storage_config {
    s3_config {
      bucket_name   = local.select_bucket == "first" ? aws_s3_bucket.test.id : aws_s3_bucket.test2.id
      bucket_prefix = "tf-test-Chat-Transcripts"
    }
    storage_type = "S3"
  }
}
`, rName2, rName3, selectBucket))
}

func testAccInstanceStorageConfigConfig_S3Config_bucketPrefix(rName, rName2, bucketPrefix string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageConfigConfig_base(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_connect_instance_storage_config" "test" {
  instance_id   = aws_connect_instance.test.id
  resource_type = "CHAT_TRANSCRIPTS"

  storage_config {
    s3_config {
      bucket_name   = aws_s3_bucket.test.id
      bucket_prefix = %[2]q
    }
    storage_type = "S3"
  }
}
`, rName2, bucketPrefix))
}

func testAccInstanceStorageConfigConfig_S3Config_encryptionConfig(rName, rName2, selectKey string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageConfigConfig_base(rName),
		fmt.Sprintf(`
locals {
  select_key = %[2]q
}

resource "aws_kms_key" "test" {
  description             = "KMS Key for Bucket 1"
  deletion_window_in_days = 10
}

resource "aws_kms_key" "test2" {
  description             = "KMS Key for Bucket 2"
  deletion_window_in_days = 10
}


resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = local.select_key == "first" ? aws_kms_key.test.arn : aws_kms_key.test2.arn
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
        key_id          = local.select_key == "first" ? aws_kms_key.test.arn : aws_kms_key.test2.arn
      }
    }
    storage_type = "S3"
  }
}
`, rName2, selectKey))
}
