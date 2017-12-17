package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceS3Bucket_basic(t *testing.T) {
	rInt := acctest.RandInt()
	arnRegexp := regexp.MustCompile(
		"^arn:aws:s3:::")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("data.aws_s3_bucket.bucket"),
					resource.TestMatchResourceAttr("data.aws_s3_bucket.bucket", "arn", arnRegexp),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "region", "us-west-2"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "bucket_domain_name", testAccBucketDomainName(rInt)),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "hosted_zone_id", HostedZoneIDForRegion("us-west-2")),
					resource.TestCheckNoResourceAttr("data.aws_s3_bucket.bucket", "website_endpoint"),
				),
			},
		},
	})
}

func TestAccDataSourceS3Bucket_website(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketWebsiteConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("data.aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"data.aws_s3_bucket.bucket", "index.html", "error.html", "", ""),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "website_endpoint", testAccWebsiteEndpoint(rInt)),
				),
			},
		},
	})
}

func TestAccDataSourceS3Bucket_whenDefaultEncryptionNotEnabled(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("data.aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.rule.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceS3Bucket_whenDefaultEncryptionEnabledWithAES256(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketConfigWithDefaultEncryptionAES256(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
				),
			},
		},
	})
}

func TestAccDataSourceS3Bucket_whenDefaultEncryptionEnabledWithAWSKMS(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketConfigWithDefaultEncryptionAWSKMS(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "aws:kms"),
					resource.TestMatchResourceAttr(
						"data.aws_s3_bucket.bucket", "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", regexp.MustCompile("^arn")),
				),
			},
		},
	})
}

func testAccAWSDataSourceS3BucketConfigWithDefaultEncryptionAWSKMS(randInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "arbitrary" {
  description             = "KMS Key for Bucket Testing %d"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "arbitrary" {
  bucket = "tf-test-bucket-%d"
  server_side_encryption_configuration {
	rule {
	  apply_server_side_encryption_by_default {
		kms_master_key_id = "${aws_kms_key.arbitrary.arn}"
	  	sse_algorithm     = "aws:kms"
	  }
	}
  }
}

data "aws_s3_bucket" "bucket" {
  bucket = "${aws_s3_bucket.arbitrary.id}"
}`, randInt, randInt)
}

func testAccAWSDataSourceS3BucketConfigWithDefaultEncryptionAES256(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  server_side_encryption_configuration {
	rule {
	  apply_server_side_encryption_by_default {
	  	sse_algorithm = "AES256"
	  }
	}
  }
}

data "aws_s3_bucket" "bucket" {
  bucket = "${aws_s3_bucket.bucket.id}"
}`, randInt)
}

func testAccAWSDataSourceS3BucketConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
}

data "aws_s3_bucket" "bucket" {
	bucket = "${aws_s3_bucket.bucket.id}"
}`, randInt)
}

func testAccAWSDataSourceS3BucketWebsiteConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
	bucket = "tf-test-bucket-%d"
	acl = "public-read"

	website {
		index_document = "index.html"
		error_document = "error.html"
	}
}

data "aws_s3_bucket" "bucket" {
	bucket = "${aws_s3_bucket.bucket.id}"
}`, randInt)
}
