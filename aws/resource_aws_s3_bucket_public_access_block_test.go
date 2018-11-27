package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSS3BucketPublicAccessBlock_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketPublicAccessBlockConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketPublicAccessBlock("aws_s3_bucket.bucket", &s3.PublicAccessBlockConfiguration{
						BlockPublicAcls:       aws.Bool(false),
						BlockPublicPolicy:     aws.Bool(false),
						IgnorePublicAcls:      aws.Bool(true),
						RestrictPublicBuckets: aws.Bool(false),
					}),
				),
			},
			{
				ResourceName:      "aws_s3_bucket_public_access_block.bucket",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSS3BucketPublicAccessBlock_update(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketPublicAccessBlockConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketPublicAccessBlock("aws_s3_bucket.bucket", &s3.PublicAccessBlockConfiguration{
						BlockPublicAcls:       aws.Bool(false),
						BlockPublicPolicy:     aws.Bool(false),
						IgnorePublicAcls:      aws.Bool(true),
						RestrictPublicBuckets: aws.Bool(false),
					}),
				),
			},
			{
				Config: testAccAWSS3BucketPublicAccessBlockConfig_updated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketPublicAccessBlock("aws_s3_bucket.bucket", &s3.PublicAccessBlockConfiguration{
						BlockPublicAcls:       aws.Bool(true),
						BlockPublicPolicy:     aws.Bool(true),
						IgnorePublicAcls:      aws.Bool(false),
						RestrictPublicBuckets: aws.Bool(true),
					}),
				),
			},
			{
				ResourceName:      "aws_s3_bucket_public_access_block.bucket",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSS3BucketPublicAccessBlock(n string, config *s3.PublicAccessBlockConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).s3conn

		resp, err := conn.GetPublicAccessBlock(&s3.GetPublicAccessBlockInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetPublicAccessBlock error: %v", err)
		}

		if !reflect.DeepEqual(resp.PublicAccessBlockConfiguration, config) {
			return fmt.Errorf("Non-equivalent config error:\n\nexpected: %s\n\n	got:%s\n",
				config, resp.PublicAccessBlockConfiguration)
		}

		return nil
	}
}

func testAccAWSS3BucketPublicAccessBlockConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "%s"
	tags {
		TestName = "TestACCAWSS3BucketPublicAccessBlock_basic"
	}
}

resource "aws_s3_bucket_public_access_block" "bucket" {
	bucket = "${aws_s3_bucket.bucket.bucket}"

	block_public_acls				= false
	block_public_policy			= false
	ignore_public_acls			= true
	restrict_public_buckets = false
}
`, bucketName)
}

func testAccAWSS3BucketPublicAccessBlockConfig_updated(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "%s"
	tags {
		TestName = "TestACCAWSS3BucketPublicAccessBlock_update"
	}
}

resource "aws_s3_bucket_public_access_block" "bucket" {
	bucket = "${aws_s3_bucket.bucket.bucket}"

	block_public_acls				= true
	block_public_policy			= true
	ignore_public_acls			= false
	restrict_public_buckets = true
}
`, bucketName)
}
