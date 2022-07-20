package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccS3BucketServerSideEncryptionConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
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

func TestAccS3BucketServerSideEncryptionConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketServerSideEncryptionConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySEEByDefault_AES256(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, s3.ServerSideEncryptionAes256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAes256),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
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

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_KMS(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, s3.ServerSideEncryptionAwsKms),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAwsKms),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
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

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_UpdateSSEAlgorithm(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, s3.ServerSideEncryptionAwsKms),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAwsKms),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, s3.ServerSideEncryptionAes256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAes256),
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

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_KMSWithMasterKeyARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKMSMasterKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAwsKms),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "aws_kms_key.test", "arn"),
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

func TestAccS3BucketServerSideEncryptionConfiguration_ApplySSEByDefault_KMSWithMasterKeyID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKMSMasterKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAwsKms),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "aws_kms_key.test", "id"),
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

func TestAccS3BucketServerSideEncryptionConfiguration_BucketKeyEnabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_keyEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", "true"),
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
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", "false"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultKeyEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAwsKms),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "aws_kms_key.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", "true"),
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
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAwsKms),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", "aws_kms_key.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.bucket_key_enabled", "false"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_defaultEncryptionDefaultKey(rName, s3.ServerSideEncryptionAwsKms),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAwsKms),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.bucket_key_enabled", "false"),
				),
			},
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_migrateNoChange(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAwsKms),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
		},
	})
}

func TestAccS3BucketServerSideEncryptionConfiguration_migrate_withChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_server_side_encryption_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketServerSideEncryptionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_defaultEncryptionDefaultKey(rName, s3.ServerSideEncryptionAwsKms),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAwsKms),
					resource.TestCheckResourceAttr(bucketResourceName, "server_side_encryption_configuration.0.rule.0.bucket_key_enabled", "false"),
				),
			},
			{
				Config: testAccBucketServerSideEncryptionConfigurationConfig_applySSEByDefaultSSEAlgorithm(rName, s3.ServerSideEncryptionAes256),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", s3.ServerSideEncryptionAes256),
					resource.TestCheckNoResourceAttr(resourceName, "rule.0.bucket_key_enabled"),
				),
			},
		},
	})
}

func testAccCheckBucketServerSideEncryptionConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_server_side_encryption_configuration" {
			continue
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketEncryptionInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketEncryption(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, tfs3.ErrCodeServerSideEncryptionConfigurationNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket server-side encryption configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil && output.ServerSideEncryptionConfiguration != nil {
			return fmt.Errorf("S3 Bucket server-side encryption configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketServerSideEncryptionConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketEncryptionInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketEncryption(input)

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket server-side encryption configuration (%s): %w", rs.Primary.ID, err)
		}

		if output == nil || output.ServerSideEncryptionConfiguration == nil {
			return fmt.Errorf("S3 Bucket server-side encryption configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketServerSideEncryptionConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {}
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
