package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSCloudFrontResponseHeadersPolicy_basic(t *testing.T) {
	rName := fmt.Sprintf("test-policy%d", sdkacctest.RandInt())
	resourceName := "aws_cloudfront_response_headers_policy.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontResponseHeadersPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "comment", "test comment"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_credentials", "true"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_headers.0.items.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_methods.0.items.0", "GET"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_origins.0.items.0", "test.example.comtest"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.origin_override", "true"),
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

func TestAccAWSCloudFrontResponseHeadersPolicy_update(t *testing.T) {
	rName := fmt.Sprintf("test-policy%d", sdkacctest.RandInt())
	resourceName := "aws_cloudfront_response_headers_policy.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontPublicKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontResponseHeadersPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "comment", "test comment"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_credentials", "true"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_headers.0.items.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_methods.0.items.0", "GET"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_origins.0.items.0", "test.example.comtest"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.origin_override", "true"),
				),
			},
			{
				Config: testAccAWSCloudFrontResponseHeadersPolicyConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "comment", "test comment updated"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_credentials", "false"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_headers.0.items.0", "test2"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_methods.0.items.0", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.access_control_allow_origins.0.items.0", "test2.example.comtest"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.0.origin_override", "false"),
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

func testAccAWSCloudFrontResponseHeadersPolicyConfig(rName string) string {
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

func testAccAWSCloudFrontResponseHeadersPolicyConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_response_headers_policy" "example" {
  name    = %[1]q
  comment = "test comment updated"
  cors_config {
    access_control_allow_credentials = false
    access_control_allow_headers {
      items = ["test2"]
    }
    access_control_allow_methods {
      items = ["POST"]
    }
    access_control_allow_origins {
      items = ["test2.example.comtest"]
    }
    origin_override = false
  }
}
`, rName)
}
