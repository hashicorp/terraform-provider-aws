package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSCloudFrontDataSourceResponseHeadersPolicy_basic(t *testing.T) {
	rName := fmt.Sprintf("test-policy%d", sdkacctest.RandInt())
	dataSourceName := "data.aws_cloudfront_response_headers_policy.example"
	resourceName := "aws_cloudfront_response_headers_policy.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontResponseHeadersPolicyDataSourceNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "comment", "test comment"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_config.0.access_control_allow_credentials", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_config.0.access_control_allow_headers.0.items.0", "test"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_config.0.access_control_allow_methods.0.items.0", "GET"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_config.0.access_control_allow_origins.0.items.0", "test.example.comtest"),
					resource.TestCheckResourceAttr(dataSourceName, "cors_config.0.origin_override", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccAWSCloudFrontResponseHeadersPolicyDataSourceNameConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_cloudfront_response_headers_policy" "example" {
  name = aws_cloudfront_response_headers_policy.example.name
}

resource "aws_cloudfront_response_headers_policy" "example" {
  name    = %[1]q
  comment = "test comment"
  cors_config {
    access_control_allow_credentials = true
    access_control_allow_headers {
      items = ["test"]
    }
    access_control_allow_methods {
      items = ["GET"]
    }
    access_control_allow_origins {
      items = ["test.example.comtest"]
    }
    origin_override = true
  }
}
`, rName)
}
