package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAWSCloudFrontResponseHeadersPolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontResponseHeadersPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontResponseHeadersPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontResponseHeadersPolicyExists(resourceName),
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

func TestAccAWSCloudFrontResponseHeadersPolicy_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontResponseHeadersPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontResponseHeadersPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontResponseHeadersPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceResponseHeadersPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFrontResponseHeadersPolicy_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_response_headers_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontResponseHeadersPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontResponseHeadersPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontResponseHeadersPolicyExists(resourceName),
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
					testAccCheckCloudFrontResponseHeadersPolicyExists(resourceName),
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

func testAccCheckCloudFrontResponseHeadersPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_response_headers_policy" {
			continue
		}

		_, err := tfcloudfront.FindResponseHeadersPolicyByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudFront Response Headers Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCloudFrontResponseHeadersPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudFront Response Headers Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		_, err := tfcloudfront.FindResponseHeadersPolicyByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSCloudFrontResponseHeadersPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_response_headers_policy" "test" {
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
resource "aws_cloudfront_response_headers_policy" "test" {
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
