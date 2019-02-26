package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSShieldProtectionGlobalAccelerate(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	globalAcceleratorConfig := fmt.Sprintf(testAccShieldProtectionGlobalAcceleratorConfig, rName)
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSShieldProtectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: globalAcceleratorConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionExists("aws_shield_protection.acctest"),
				),
			},
		},
	})
}

func TestAccAWSShieldProtectionElasticIPAddress(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	elasticIPAddressConfig := fmt.Sprintf(testAccShieldProtectionElasticIPAddressConfig, rName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: elasticIPAddressConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionExists("aws_shield_protection.acctest"),
				),
			},
		},
	})
}

func TestAccAWSShieldProtectionAlb(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	albConfig := fmt.Sprintf(testAccShieldProtectionAlbConfig, rName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: albConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionExists("aws_shield_protection.acctest"),
				),
			},
		},
	})
}

func TestAccAWSShieldProtectionElb(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	elbConfig := fmt.Sprintf(testAccShieldProtectionElbConfig, rName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: elbConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionExists("aws_shield_protection.acctest"),
				),
			},
		},
	})
}

func TestAccAWSShieldProtectionCloudfront(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	cloudfrontConfig := fmt.Sprintf(testAccShieldProtectionCloudfrontConfig, rName, testAccShieldProtectionCloudfrontRetainConfig())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: cloudfrontConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionExists("aws_shield_protection.acctest"),
				),
			},
		},
	})
}

func TestAccAWSShieldProtectionRoute53(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	route53Config := fmt.Sprintf(testAccShieldProtectionRoute53HostedZoneConfig, rName)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: route53Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSShieldProtectionExists("aws_shield_protection.acctest"),
				),
			},
		},
	})
}

func TestAccAWSShieldProtection_importBasic(t *testing.T) {
	resourceName := "aws_shield_protection.acctest"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	elasticIPAddressConfig := fmt.Sprintf(testAccShieldProtectionElasticIPAddressConfig, rName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSShieldProtectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: elasticIPAddressConfig,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSShieldProtectionDestroy(s *terraform.State) error {
	shieldconn := testAccProvider.Meta().(*AWSClient).shieldconn

	for _, rs := range s.RootModule().Resources {

		if rs.Type != "aws_shield_protection" {
			continue
		}

		input := &shield.DescribeProtectionInput{
			ProtectionId: aws.String(rs.Primary.ID),
		}

		resp, err := shieldconn.DescribeProtection(input)
		if err == nil {
			if len(*resp.Protection.Id) > 0 && *resp.Protection.Id == rs.Primary.ID {
				return fmt.Errorf("The Shield protection with ID %v still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAWSShieldProtectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

// Set the environment variable TF_TEST_CLOUDFRONT_RETAIN
// when doing manual tests so that the test is not waiting for
// the distribution to be removed completely.
func testAccShieldProtectionCloudfrontRetainConfig() string {
	if _, ok := os.LookupEnv("TF_TEST_CLOUDFRONT_RETAIN"); ok {
		return "retain_on_delete = true"
	}
	return ""
}

var testAccShieldProtectionRoute53HostedZoneConfig = `
variable "name" {
  default = "%s"
}

resource "aws_route53_zone" "acctest" {
  name    = "${var.name}.com."
  comment = "Terraform Acceptance Tests"

  tags {
    foo  = "bar"
    Name = "${var.name}"
  }
}

resource "aws_shield_protection" "acctest" {
  name         = "${var.name}"
  resource_arn = "arn:aws:route53:::hostedzone/${aws_route53_zone.acctest.zone_id}"
}
`

var testAccShieldProtectionElbConfig = `
data "aws_availability_zones" "available" {}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

variable "name" {
  default = "%s"
}

resource "aws_vpc" "acctest" {
  cidr_block = "10.0.0.0/16"

  tags {
    foo  = "bar"
    Name = "${var.name}"
  }
}

resource "aws_subnet" "acctest" {
  count                   = 2
  vpc_id                  = "${aws_vpc.acctest.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    foo  = "bar"
    Name = "${var.name}"
  }

}
resource "aws_elb" "acctest" {
  name = "${var.name}"
  #availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  subnets            = ["${aws_subnet.acctest.*.id[0]}", "${aws_subnet.acctest.*.id[1]}"]
  internal           = true

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags {
    foo  = "bar"
    Name = "${var.name}"
  }

  cross_zone_load_balancing = true
}

resource "aws_shield_protection" "acctest" {
  name         = "${var.name}"
  resource_arn = "${aws_elb.acctest.arn}"
}
`

var testAccShieldProtectionAlbConfig = `
data "aws_availability_zones" "available" {}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

variable "name" {
  default = "%s"
}

resource "aws_lb" "acctest" {
  name            = "${var.name}"
  internal        = true
  security_groups = ["${aws_security_group.acctest.id}"]
  subnets         = ["${aws_subnet.acctest.*.id[0]}", "${aws_subnet.acctest.*.id[1]}"]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags {
    foo  = "bar"
    Name = "${var.name}"
  }
}

resource "aws_vpc" "acctest" {
  cidr_block = "10.0.0.0/16"

  tags {
    foo  = "bar"
    Name = "${var.name}"
  }
}

resource "aws_subnet" "acctest" {
  count                   = 2
  vpc_id                  = "${aws_vpc.acctest.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    foo  = "bar"
    Name = "${var.name}"
  }

}

resource "aws_security_group" "acctest" {
  name        = "${var.name}"
  description = "acctest"
  vpc_id      = "${aws_vpc.acctest.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    foo  = "bar"
    Name = "${var.name}"
  }
}

resource "aws_shield_protection" "acctest" {
  name         = "${var.name}"
  resource_arn = "${aws_lb.acctest.arn}"
}
`

var testAccShieldProtectionCloudfrontConfig = `
variable "name" {
  default = "%s"
}

resource "aws_cloudfront_distribution" "acctest" {
  origin {
    custom_origin_config = {
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

  enabled         = true
  is_ipv6_enabled = true
  comment         = "acctest"

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

  tags {
    foo  = "bar"
    Name = "${var.name}"
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %s
}

resource "aws_shield_protection" "acctest" {
  name         = "${var.name}"
  resource_arn = "${aws_cloudfront_distribution.acctest.arn}"
}
`

var testAccShieldProtectionElasticIPAddressConfig = `
variable "name" {
  default = "%s"
}

data "aws_availability_zones" "available" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_eip" "acctest" {
  tags {
    foo  = "bar"
    Name = "${var.name}"
  }
}

resource "aws_shield_protection" "acctest" {
  name         = "${var.name}"
  resource_arn = "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.acctest.id}"
}
`

var testAccShieldProtectionGlobalAcceleratorConfig = `
provider "aws" {
  region = "us-west-2"
  alias  = "globalaccelerator"
}

variable "name" {
  default = "%s"
}

resource "aws_shield_protection" "acctest" {
  name         = "${var.name}"
  resource_arn = "${aws_globalaccelerator_accelerator.acctest.id}"
}

resource "aws_globalaccelerator_accelerator" "acctest" {
  provider        = "aws.globalaccelerator"
  name            = "${var.name}"
  ip_address_type = "IPV4"
  enabled         = true
}
`
