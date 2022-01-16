package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccS3BucketAccelerateConfiguration_basic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketAccelerateConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(bucketName, s3.BucketAccelerateStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "status", s3.BucketAccelerateStatusEnabled),
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

func TestAccS3BucketAccelerateConfiguration_update(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketAccelerateConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(bucketName, s3.BucketAccelerateStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "status", s3.BucketAccelerateStatusEnabled),
				),
			},
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(bucketName, s3.BucketAccelerateStatusSuspended),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "status", s3.BucketAccelerateStatusSuspended),
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

func TestAccS3BucketAccelerateConfiguration_disappears(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketAccelerateConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(bucketName, s3.BucketAccelerateStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketAccelerateConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBucketAccelerateConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_accelerate_configuration" {
			continue
		}

		input := &s3.GetBucketAccelerateConfigurationInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetBucketAccelerateConfiguration(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket accelerate configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("S3 Bucket accelerate configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketAccelerateConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		input := &s3.GetBucketAccelerateConfigurationInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetBucketAccelerateConfiguration(input)

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket accelerate configuration (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("S3 Bucket accelerate configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketAccelerateConfigurationBasicConfig(bucketName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  lifecycle {
    ignore_changes = [
	  acceleration_status
    ]
  }
}

resource "aws_s3_bucket_accelerate_configuration" "test" {
  bucket = aws_s3_bucket.test.id
  status = %[2]q
}
`, bucketName, status)
}
