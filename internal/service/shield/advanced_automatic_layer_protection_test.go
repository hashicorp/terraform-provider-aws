package shield_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccShieldProtection_automaticLayerResponse(t *testing.T) {

	resourceName := "aws_shield_advanced_automatic_layer_protection.test"
	rName := sdkacctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, shield.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutomaticLayerResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionConfig_automaticLayerResponse(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutomaticLayerResponseExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckAutomaticLayerResponseDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_shield_advanced_automatic_layer_protection" {
			continue
		}

		input := &shield.DescribeProtectionInput{
			ResourceArn: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeProtection(input)

		if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Protection.ApplicationLayerAutomaticResponseConfiguration != nil &&
			aws.StringValue(resp.Protection.ResourceArn) == rs.Primary.ID {
			return fmt.Errorf("The Shield protection with ARN %v still has protection enabled", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAutomaticLayerResponseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn

		input := &shield.DescribeProtectionInput{
			ResourceArn: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeProtection(input)

		if resp.Protection.ApplicationLayerAutomaticResponseConfiguration == nil {
			return fmt.Errorf(
				"The Shield protection with ARN %v has no Application Layer Automatic Response Configuration",
				rs.Primary.ID)
		}

		return err
	}
}

func testAccProtectionConfig_automaticLayerResponse(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "CLOUDFRONT"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
  lifecycle {
    ignore_changes = [
      rule,
    ]
  }
}

resource "aws_cloudfront_distribution" "test" {
  origin {
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"

      origin_ssl_protocols = [
        "TLSv1",
        "TLSv1.1",
        "TLSv1.2",
      ]
    }

    # This is a fake origin and it's set to this name to indicate that.
    domain_name = "%[1]s.com"
    origin_id   = %[1]q
  }

  enabled             = false
  wait_for_deployment = false
  web_acl_id 		  =  aws_wafv2_web_acl.test.arn

  default_cache_behavior {
    allowed_methods  = ["HEAD", "DELETE", "POST", "GET", "OPTIONS", "PUT", "PATCH"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = %[1]q

    forwarded_values {
      query_string = false
      headers      = ["*"]

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  tags = {
    Name = %[1]q
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = aws_cloudfront_distribution.test.arn
}

resource "aws_shield_advanced_automatic_layer_protection" "test" {
  resource_arn = aws_cloudfront_distribution.test.arn
  action = "COUNT"
}
`, rName)
}
