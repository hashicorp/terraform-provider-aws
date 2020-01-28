package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSS3BucketAnalyticsConfiguration_basic(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketAnalyticsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "storage_class_analysis.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketAnalyticsConfiguration_removed(t *testing.T) {
	var ac s3.AnalyticsConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_analytics_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketAnalyticsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketAnalyticsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationExists(resourceName, &ac),
				),
			},
			{
				Config: testAccAWSS3BucketAnalyticsConfiguration_removed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketAnalyticsConfigurationRemoved(rName, rName),
				),
			},
		},
	})
}

func testAccCheckAWSS3BucketAnalyticsConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_analytics_configuration" {
			continue
		}

		bucket, name, err := resourceAwsS3BucketAnalyticsConfigurationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		return waitForDeleteS3BucketAnalyticsConfiguration(conn, bucket, name, 1*time.Minute)

	}
	return nil
}

func testAccCheckAWSS3BucketAnalyticsConfigurationExists(n string, ac *s3.AnalyticsConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).s3conn
		output, err := conn.GetBucketAnalyticsConfiguration(&s3.GetBucketAnalyticsConfigurationInput{
			Bucket: aws.String(rs.Primary.Attributes["bucket"]),
			Id:     aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if output == nil || output.AnalyticsConfiguration == nil {
			return fmt.Errorf("error reading S3 Bucket Analytics Configuration %q: empty response", rs.Primary.ID)
		}

		*ac = *output.AnalyticsConfiguration

		return nil
	}
}
func testAccCheckAWSS3BucketAnalyticsConfigurationRemoved(bucket, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).s3conn
		return waitForDeleteS3BucketAnalyticsConfiguration(conn, bucket, name, 1*time.Minute)
	}
}

func testAccAWSS3BucketAnalyticsConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_analytics_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
}
`, rName) + testAccAWSS3BucketAnalyticsConfiguration_base(rName)
}

func testAccAWSS3BucketAnalyticsConfiguration_removed(rName string) string {
	return testAccAWSS3BucketAnalyticsConfiguration_base(rName)
}

func testAccAWSS3BucketAnalyticsConfiguration_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}
