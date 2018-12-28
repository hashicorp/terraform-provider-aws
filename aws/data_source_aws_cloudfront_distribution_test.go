package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCloudfrontDistribution_dataSource_basic(t *testing.T) {
	ri := acctest.RandInt()
	testConfig := fmt.Sprintf(testAccAWSCloudFrontDistributionDsConfig, ri, distOriginBucket, distLogBucket, testAccAWSCloudFrontDistributionRetainConfig())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontDistributionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.aws_cloudfront_distribution.s3_distribution", "arn",
						regexp.MustCompile(`^arn:aws:cloudfront::\d+:distribution\/[A-Z]\w+$`)),
					resource.TestMatchResourceAttr("data.aws_cloudfront_distribution.s3_distribution", "status",
						regexp.MustCompile(`^InProgress|Deployed$`)),
					resource.TestMatchResourceAttr("data.aws_cloudfront_distribution.s3_distribution", "domain_name",
						regexp.MustCompile(`^[a-z0-9]+\.cloudfront\.net$`)),
					resource.TestMatchResourceAttr("data.aws_cloudfront_distribution.s3_distribution", "etag",
						regexp.MustCompile(`^[A-Z0-9]+$`)),
					resource.TestMatchResourceAttr("data.aws_cloudfront_distribution.s3_distribution", "in_progress_validation_batches",
						regexp.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

var distOriginBucket = fmt.Sprintf(`
resource "aws_s3_bucket" "s3_bucket_origin" {
	bucket = "mybucket.${var.rand_id}"
	acl = "public-read"
}
`)

var distLogBucket = fmt.Sprintf(`
resource "aws_s3_bucket" "s3_bucket_logs" {
	bucket = "mylogs.${var.rand_id}"
	acl = "public-read"
}
`)

var testAccAWSCloudFrontDistributionDsConfig = `
variable rand_id {
	default = %d
}

# origin bucket
%s

# log bucket
%s

resource "aws_cloudfront_distribution" "s3_distribution" {
	origin {
		domain_name = "${aws_s3_bucket.s3_bucket_origin.id}.s3.amazonaws.com"
		origin_id = "myS3Origin"
	}
	enabled = true
	default_root_object = "index.html"
	logging_config {
		include_cookies = false
		bucket = "${aws_s3_bucket.s3_bucket_logs.id}.s3.amazonaws.com"
		prefix = "myprefix"
	}
	aliases = [ "mysite.${var.rand_id}.example.com", "yoursite.${var.rand_id}.example.com" ]
	default_cache_behavior {
		allowed_methods = [ "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT" ]
		cached_methods = [ "GET", "HEAD" ]
		target_origin_id = "myS3Origin"
		forwarded_values {
			query_string = false
			cookies {
				forward = "none"
			}
		}
		viewer_protocol_policy = "allow-all"
		min_ttl = 0
		default_ttl = 3600
		max_ttl = 86400
	}
	price_class = "PriceClass_200"
	restrictions {
		geo_restriction {
			restriction_type = "whitelist"
			locations = [ "US", "CA", "GB", "DE" ]
		}
	}
	viewer_certificate {
		cloudfront_default_certificate = true
	}
	%s
}

data "aws_cloudfront_distribution" "s3_distribution" {
	id = "${aws_cloudfront_distribution.s3_distribution.id}"
}
`
