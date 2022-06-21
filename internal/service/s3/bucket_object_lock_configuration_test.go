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

func TestAccS3BucketObjectLockConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketObjectLockConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", s3.ObjectLockEnabledEnabled),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", s3.ObjectLockRetentionModeCompliance),
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

func TestAccS3BucketObjectLockConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketObjectLockConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketObjectLockConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketObjectLockConfiguration_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(resourceName),
				),
			},
			{
				Config: testAccBucketObjectLockConfigurationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", s3.ObjectLockEnabledEnabled),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.years", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", s3.ObjectLockRetentionModeGovernance),
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

func TestAccS3BucketObjectLockConfiguration_migrate_noChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketObjectLockConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledDefaultRetention(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.object_lock_enabled", s3.ObjectLockEnabledEnabled),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.rule.0.default_retention.0.mode", s3.ObjectLockRetentionModeCompliance),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.rule.0.default_retention.0.days", "3"),
				),
			},
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", s3.ObjectLockEnabledEnabled),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", s3.ObjectLockRetentionModeCompliance),
				),
			},
		},
	})
}

func TestAccS3BucketObjectLockConfiguration_migrate_withChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_object_lock_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketObjectLockConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledNoDefaultRetention(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.object_lock_enabled", s3.ObjectLockEnabledEnabled),
					resource.TestCheckResourceAttr(bucketResourceName, "object_lock_configuration.0.rule.#", "0"),
				),
			},
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", s3.ObjectLockEnabledEnabled),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", s3.ObjectLockRetentionModeCompliance),
				),
			},
		},
	})
}

func testAccCheckBucketObjectLockConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_object_lock_configuration" {
			continue
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetObjectLockConfigurationInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetObjectLockConfiguration(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, tfs3.ErrCodeObjectLockConfigurationNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket Object Lock configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("S3 Bucket Object Lock configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketObjectLockConfigurationExists(resourceName string) resource.TestCheckFunc {
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

		input := &s3.GetObjectLockConfigurationInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetObjectLockConfiguration(input)

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket Object Lock configuration (%s): %w", rs.Primary.ID, err)
		}

		if output == nil || output.ObjectLockConfiguration == nil {
			return fmt.Errorf("S3 Bucket Object Lock configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketObjectLockConfigurationConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    default_retention {
      mode = %[2]q
      days = 3
    }
  }
}
`, bucketName, s3.ObjectLockRetentionModeCompliance)
}

func testAccBucketObjectLockConfigurationConfig_update(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "aws_s3_bucket_object_lock_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    default_retention {
      mode  = %[2]q
      years = 1
    }
  }
}
`, bucketName, s3.ObjectLockModeGovernance)
}
