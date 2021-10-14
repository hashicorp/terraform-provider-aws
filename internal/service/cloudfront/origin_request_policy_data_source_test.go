package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSCloudFrontDataSourceOriginRequestPolicy_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dataSourceName := "data.aws_cloudfront_origin_request_policy.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOriginRequestPolicyDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "comment", "test comment"),
					resource.TestCheckResourceAttr(dataSourceName, "cookies_config.0.cookie_behavior", "whitelist"),
					resource.TestCheckResourceAttr(dataSourceName, "cookies_config.0.cookies.0.items.0", "test"),
					resource.TestCheckResourceAttr(dataSourceName, "headers_config.0.header_behavior", "whitelist"),
					resource.TestCheckResourceAttr(dataSourceName, "headers_config.0.headers.0.items.0", "test"),
					resource.TestCheckResourceAttr(dataSourceName, "query_strings_config.0.query_string_behavior", "whitelist"),
					resource.TestCheckResourceAttr(dataSourceName, "query_strings_config.0.query_strings.0.items.0", "test"),
				),
			},
		},
	})
}

func testAccOriginRequestPolicyDataSourceConfig(rInt int) string {
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
