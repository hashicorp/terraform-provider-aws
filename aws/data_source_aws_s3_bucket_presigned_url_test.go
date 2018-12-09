package aws

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsS3BucketPresignedURL_get(t *testing.T) {
	rInt := acctest.RandInt()

	bucket := fmt.Sprintf("tf-s3-presigned-url-test-%d", rInt)
	key := fmt.Sprintf("tf-s3-presigned-url-test-get-%d", rInt)
	expirationTime := "30"

	conf := testAccDataSourceAwsS3BucketPresignedURL_get(bucket, key, expirationTime)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: conf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_s3_bucket_presigned_url.presigned_get_url", "bucket", bucket),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_presigned_url.presigned_get_url", "key", key),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_presigned_url.presigned_get_url", "expiration_time", expirationTime),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_presigned_url.presigned_get_url", "put", "false"),
					resource.TestCheckResourceAttrSet("data.aws_s3_bucket_presigned_url.presigned_get_url", "url"),
					testAccDataCheckS3BucketPresignedGetURL("data.aws_s3_bucket_presigned_url.presigned_get_url"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsS3BucketPresignedURL_put(t *testing.T) {
	rInt := acctest.RandInt()

	bucket := fmt.Sprintf("tf-s3-presigned-url-test-%d", rInt)
	key := fmt.Sprintf("tf-s3-presigned-url-test-put-%d", rInt)
	expirationTime := "30"

	conf := testAccDataSourceAwsS3BucketPresignedURL_put(bucket, key, expirationTime)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: conf,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_s3_bucket_presigned_url.presigned_post_url", "bucket", bucket),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_presigned_url.presigned_post_url", "key", key),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_presigned_url.presigned_post_url", "expiration_time", expirationTime),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_presigned_url.presigned_post_url", "put", "true"),
					resource.TestCheckResourceAttrSet("data.aws_s3_bucket_presigned_url.presigned_post_url", "url"),
					testAccDataCheckS3BucketPresignedPutURL("data.aws_s3_bucket_presigned_url.presigned_post_url"),
				),
			},
		},
	})
}

func testAccDataCheckS3BucketPresignedGetURL(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find S3 Presigned URL data source: %s", n)
		}

		resp, err := http.Get(rs.Primary.Attributes["url"])
		if err != nil {
			return fmt.Errorf("Error making get request to: %s %s", rs.Primary.Attributes["url"], err)
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("Expected status code 200 but got: %v", resp.StatusCode)
		}

		return nil
	}
}

func testAccDataCheckS3BucketPresignedPutURL(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find S3 Presigned URL data source: %s", n)
		}

		req, err := http.NewRequest("PUT", rs.Primary.Attributes["url"], strings.NewReader("test"))
		if err != nil {
			return fmt.Errorf("Error creating request: %s %s", rs.Primary.Attributes["url"], err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("Expected status code 200 but got: %v", resp.StatusCode)
		}

		return nil
	}
}

func testAccDataSourceAwsS3BucketPresignedURL_get(bucketName string, key string, expirationTime string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "%s"
}

resource "aws_s3_bucket_object" "object" {
	bucket = "${aws_s3_bucket.bucket.id}"
	key = "%s"
	content = <<CONTENT
{"msg": "Howdy!"}
CONTENT
}

data "aws_s3_bucket_presigned_url" "presigned_get_url" {
	bucket = "${aws_s3_bucket.bucket.id}"
	key = "%s"
	expiration_time = %s
}
	`, bucketName, key, key, expirationTime)
}

func testAccDataSourceAwsS3BucketPresignedURL_put(bucketName string, key string, expirationTime string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "%s"
	force_destroy = true
}

data "aws_s3_bucket_presigned_url" "presigned_post_url" {
	bucket = "${aws_s3_bucket.bucket.id}"
	key = "%s"
	expiration_time = %s
	put = true
}
	`, bucketName, key, expirationTime)
}
