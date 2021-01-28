package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCloudFrontDataSourceOriginRequestPolicy_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontDataSourceOriginRequestPolicyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudfront_origin_request_policy.example", "comment", "test comment"),
					resource.TestCheckResourceAttr("data.aws_cloudfront_origin_request_policy.example", "cookies_config.0.cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr("data.aws_cloudfront_origin_request_policy.example", "cookies_config.0.cookies.0.items.0", "test"),
					resource.TestCheckResourceAttr("data.aws_cloudfront_origin_request_policy.example", "headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr("data.aws_cloudfront_origin_request_policy.example", "headers_config.0.headers.0.items.0", "test"),
					resource.TestCheckResourceAttr("data.aws_cloudfront_origin_request_policy.example", "query_strings_config.0.query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr("data.aws_cloudfront_origin_request_policy.example", "query_strings_config.0.query_strings.0.items.0", "test"),
				),
			},
		},
	})
}
func testAccAWSCloudFrontDataSourceOriginRequestPolicyConfig(rInt int) string {
	return fmt.Sprintf(`
data "aws_cloudfront_origin_request_policy" "example" {
  name = aws_cloudfront_origin_request_policy.example.name
}

resource "aws_cloudfront_origin_request_policy" "example" {
  name    = "test-policy%[1]d"
  comment = "test comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["test"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["test"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["test"]
    }
  }
}
`, rInt)
}
