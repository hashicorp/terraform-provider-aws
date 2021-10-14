package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccDataSourceS3Bucket_basic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()
	hostedZoneID, _ := tfs3.HostedZoneIDForRegion(region)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("data.aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttrPair("data.aws_s3_bucket.bucket", "arn", "aws_s3_bucket.bucket", "arn"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "region", region),
					testAccCheckS3BucketDomainName("data.aws_s3_bucket.bucket", "bucket_domain_name", bucketName),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "bucket_regional_domain_name", testAccBucketRegionalDomainName(bucketName, region)),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "hosted_zone_id", hostedZoneID),
					resource.TestCheckNoResourceAttr("data.aws_s3_bucket.bucket", "website_endpoint"),
				),
			},
		},
	})
}

func TestAccDataSourceS3Bucket_website(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteDataSourceConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("data.aws_s3_bucket.bucket"),
					testAccCheckBucketWebsite("data.aws_s3_bucket.bucket", "index.html", "error.html", "", ""),
					testAccCheckS3BucketWebsiteEndpoint("data.aws_s3_bucket.bucket", "website_endpoint", bucketName, region),
				),
			},
		},
	})
}

func testAccBucketDataSourceConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

data "aws_s3_bucket" "bucket" {
  bucket = aws_s3_bucket.bucket.id
}
`, bucketName)
}

func testAccBucketWebsiteDataSourceConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}

data "aws_s3_bucket" "bucket" {
  bucket = aws_s3_bucket.bucket.id
}
`, bucketName)
}
