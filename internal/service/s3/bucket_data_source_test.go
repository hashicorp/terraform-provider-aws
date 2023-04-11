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

func TestAccS3BucketDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()
	hostedZoneID, _ := tfs3.HostedZoneIDForRegion(region)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, "data.aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttrPair("data.aws_s3_bucket.bucket", "arn", "aws_s3_bucket.bucket", "arn"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "region", region),
					testAccCheckBucketDomainName("data.aws_s3_bucket.bucket", "bucket_domain_name", bucketName),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "bucket_regional_domain_name", testAccBucketRegionalDomainName(bucketName, region)),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "hosted_zone_id", hostedZoneID),
					resource.TestCheckNoResourceAttr("data.aws_s3_bucket.bucket", "website_endpoint"),
				),
			},
		},
	})
}

func TestAccS3BucketDataSource_website(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceConfig_website(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, "data.aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttrPair("data.aws_s3_bucket.bucket", "bucket", "aws_s3_bucket.bucket", "id"),
					resource.TestCheckResourceAttrPair("data.aws_s3_bucket.bucket", "website_domain", "aws_s3_bucket_website_configuration.test", "website_domain"),
					resource.TestCheckResourceAttrPair("data.aws_s3_bucket.bucket", "website_endpoint", "aws_s3_bucket_website_configuration.test", "website_endpoint"),
				),
			},
		},
	})
}

func TestAccS3BucketDataSource_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceConfig_encryption(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, "data.aws_s3_bucket.bucket"),
					//resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "server_side_encryption_configuration", aws_s3_bucket_server_side_encryption_configuration),
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

func testAccBucketDataSourceConfig_website(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.bucket.id
  index_document {
    suffix = "index.html"
  }
  error_document {
    key = "error.html"
  }
}

data "aws_s3_bucket" "bucket" {
  # Must have bucket website configured first
  bucket = aws_s3_bucket_website_configuration.test.id
}
`, bucketName)
}

func testAccBucketDataSourceConfig_encryption(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "AES256"
    }
  }
}

data "aws_s3_bucket" "server_side_encryption_configuration" {
  bucket = aws_s3_bucket.bucket.server_side_encryption_configuration
}
`, bucketName)
}
