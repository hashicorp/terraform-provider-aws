package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDataSourceCloudFrontDistribution_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDistributionData,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.aws_cloudfront_distribution.test",
						"domain_name",
						regexp.MustCompile(`^[a-z0-9]+\.cloudfront\.net$`),
					),
					resource.TestCheckResourceAttrSet("data.aws_cloudfront_distribution.test", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_cloudfront_distribution.test", "domain_name"),
					resource.TestCheckResourceAttrSet("data.aws_cloudfront_distribution.test", "etag"),
					resource.TestCheckResourceAttrSet("data.aws_cloudfront_distribution.test", "hosted_zone_id"),
					resource.TestCheckResourceAttrSet("data.aws_cloudfront_distribution.test", "in_progress_validation_batches"),
					resource.TestCheckResourceAttrSet("data.aws_cloudfront_distribution.test", "last_modified_time"),
					resource.TestCheckResourceAttrSet("data.aws_cloudfront_distribution.test", "status"),
				),
			},
		},
	})
}

var testAccAWSCloudFrontDistributionData = fmt.Sprintf(`
%s

data "aws_cloudfront_distribution" "test" {
	id = "${aws_cloudfront_distribution.s3_distribution.id}"
}
`, fmt.Sprintf(testAccAWSCloudFrontDistributionS3ConfigWithTags, acctest.RandInt(), originBucket, logBucket, testAccAWSCloudFrontDistributionRetainConfig()))
