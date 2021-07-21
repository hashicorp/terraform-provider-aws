package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSShieldProtectionHealthCheck_disappears(t *testing.T) {
	healthCheckResourceName := "aws_shield_protection_health_check_association.acctest"
	protectionResourceName := "aws_shield_protection.acctest"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(shield.EndpointsID, t)
			testAccPreCheckAWSShield(t)
		},
		ErrorCheck:   testAccErrorCheck(t, shield.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionHealthCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShieldProtectionaHealthCheckAssociationConfig(rName, testAccShieldProtectionCloudfrontRetainConfig()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionHealthCheckAssociationExists(protectionResourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsShieldProtection(), healthCheckResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSShieldProtectionHealthCheckAssociation_basic(t *testing.T) {
	resourceName := "aws_shield_protection_health_check_assocation.acctest"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(shield.EndpointsID, t)
			testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t)
			testAccPreCheckAWSShield(t)
		},
		ErrorCheck:   testAccErrorCheck(t, shield.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionHealthCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccShieldProtectionaHealthCheckAssociationConfig(rName, testAccShieldProtectionCloudfrontRetainConfig()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionHealthCheckAssociationExists(resourceName),
				),
			},
			//{
			//	ResourceName:      resourceName,
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//},
		},
	})
}

func testAccCheckAWSShieldProtectionHealthCheckAssociationDestroy(s *terraform.State) error {
	shieldconn := testAccProvider.Meta().(*AWSClient).shieldconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_shield_protection" {
			continue
		}

		input := &shield.DescribeProtectionInput{
			ProtectionId: aws.String(rs.Primary.ID),
		}

		resp, err := shieldconn.DescribeProtection(input)

		if isAWSErr(err, shield.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.Protection != nil && aws.StringValue(resp.Protection.Id) == rs.Primary.ID {
			return fmt.Errorf("The Shield protection with ID %v still exists", rs.Primary.ID)
		}

		if resp != nil && resp.Protection != nil && len(aws.StringValueSlice(resp.Protection.HealthCheckIds)) == 0 {
			return fmt.Errorf("The Shield protection HealthCheck with IDs %v still exists", aws.StringValueSlice(resp.Protection.HealthCheckIds))
		}
	}

	return nil
}

func testAccCheckAWSShieldProtectionHealthCheckAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).shieldconn

		input := &shield.DescribeProtectionInput{
			ProtectionId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeProtection(input)

		if err != nil {
			return err
		}

		if resp == nil || resp.Protection == nil {
			return fmt.Errorf("The Shield protection does not exist")
		}

		if resp.Protection.HealthCheckIds == nil || len(aws.StringValueSlice(resp.Protection.HealthCheckIds)) != 1 {
			return fmt.Errorf("The Shield protection HealthCheck does not exist")
		}

		return nil
	}
}

func testAccShieldProtectionaHealthCheckAssociationConfig(rName, retainOnDelete string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

resource "aws_cloudfront_distribution" "acctest" {
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
    domain_name = "${var.name}.com"
    origin_id   = "acctest"
  }

  enabled             = false
  wait_for_deployment = false

  default_cache_behavior {
    allowed_methods  = ["HEAD", "DELETE", "POST", "GET", "OPTIONS", "PUT", "PATCH"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "acctest"

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
    foo  = "bar"
    Name = var.name
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}

resource "aws_shield_protection" "acctest" {
  name         = var.name
  resource_arn = aws_cloudfront_distribution.acctest.arn
}

resource "aws_route53_health_check" "acctest" {
  fqdn              = "example.com"
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "5"
  request_interval  = "30"

  tags = {
    Name = "tf-test-health-check"
  }
}

resource "aws_shield_protection_health_check_association" "acctest" {
  shield_protection_id = aws_shield_protection.acctest.id
  health_check_id      = aws_route53_health_check.acctest.id
}
`, rName, retainOnDelete)
}
