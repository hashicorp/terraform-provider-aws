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

func init() {
	acctest.RegisterServiceErrorCheckFunc(cloudfront.EndpointsID, testAccErrorCheckSkipFunction)
}

func testAccErrorCheckSkipFunction(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"InvalidParameterValueException: Unsupported source arn",
	)
}

func TestAccCloudFrontFunction_basic(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("function/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "code"),
					resource.TestCheckResourceAttr(resourceName, "comment", ""),
					resource.TestCheckResourceAttr(resourceName, "etag", "ETVPDKIKX0DER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "runtime", "cloudfront-js-1.0"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNASSOCIATED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func TestAccCloudFrontFunction_disappears(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudfront_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontFunction_publish(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publish(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "ETVPDKIKX0DER"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", ""),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNPUBLISHED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
			{
				Config: testAccFunctionConfig_publish(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "ETVPDKIKX0DER"),
					resource.TestCheckResourceAttr(resourceName, "live_stage_etag", "ETVPDKIKX0DER"),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNASSOCIATED"),
				),
			},
		},
	})
}

// If you are testing manually and can't wait for deletion, set the
// TF_TEST_CLOUDFRONT_RETAIN environment variable.
func TestAccCloudFrontFunction_associated(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_associated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					// After creation the function will be in UNASSOCIATED status.
					// Apply the same configuration and it will move to DEPLOYED status.
					resource.TestCheckResourceAttr(resourceName, "status", "UNASSOCIATED"),
				),
			},
			{
				Config: testAccFunctionConfig_associated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "status", "DEPLOYED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
			{
				Config: testAccFunctionConfig_unassociated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccCloudFrontFunction_Update_code(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "ETVPDKIKX0DER"),
				),
			},
			{
				Config: testAccFunctionConfig_codeUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "E3UN6WX5RRO2AG"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func TestAccCloudFrontFunction_UpdateCodeAndPublish(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publish(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "ETVPDKIKX0DER"),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNPUBLISHED"),
				),
			},
			{
				Config: testAccFunctionConfig_codeUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "etag", "E3UN6WX5RRO2AG"),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", "UNASSOCIATED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func TestAccCloudFrontFunction_Update_comment(t *testing.T) {
	var conf cloudfront.DescribeFunctionOutput
	resourceName := "aws_cloudfront_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_comment(rName, "test 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "comment", "test 1"),
				),
			},
			{
				Config: testAccFunctionConfig_comment(rName, "test 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "comment", "test 2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func testAccCheckFunctionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_function" {
			continue
		}

		_, err := tfcloudfront.FindFunctionByNameAndStage(conn, rs.Primary.ID, cloudfront.FunctionStageDevelopment)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudFront Function %s still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckFunctionExists(n string, v *cloudfront.DescribeFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("CloudFront Function not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("CloudFront Function ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		output, err := tfcloudfront.FindFunctionByNameAndStage(conn, rs.Primary.ID, cloudfront.FunctionStageDevelopment)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFunctionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT
}
`, rName)
}

func testAccFunctionConfig_publish(rName string, publish bool) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT

  publish = %[2]t
}
`, rName, publish)
}

func testAccFunctionConfig_associated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"

    function_association {
      event_type   = "viewer-request"
      function_arn = aws_cloudfront_function.test.arn
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}

resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT

  publish = true
}
`, rName, testAccDistributionRetainConfig())
}

func testAccFunctionConfig_unassociated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  origin {
    domain_name = "www.example.com"
    origin_id   = "myCustomOrigin"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["SSLv3", "TLSv1"]
    }
  }

  enabled = true

  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "myCustomOrigin"
    smooth_streaming = false

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }

    viewer_protocol_policy = "allow-all"
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["US", "CA", "GB", "DE"]
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}

resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT

  publish = true
}
`, rName, testAccDistributionRetainConfig())
}

func testAccFunctionConfig_codeUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  code    = <<-EOT
function handler(event) {
	// updated code
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT
}
`, rName)
}

func testAccFunctionConfig_comment(rName, comment string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_function" "test" {
  name    = %[1]q
  runtime = "cloudfront-js-1.0"
  comment = %[2]q
  code    = <<-EOT
function handler(event) {
	var response = {
		statusCode: 302,
		statusDescription: 'Found',
		headers: {
			'cloudfront-functions': { value: 'generated-by-CloudFront-Functions' },
			'location': { value: 'https://aws.amazon.com/cloudfront/' }
		}
	};
	return response;
}
EOT
}
`, rName, comment)
}
