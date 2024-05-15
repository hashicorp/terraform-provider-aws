// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfshield "github.com/hashicorp/terraform-provider-aws/internal/service/shield"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccShieldProtection_globalAccelerator(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionConfig_globalAccelerator(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccShieldProtection_elasticIPAddress(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionConfig_elasticIPAddress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccShieldProtection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionConfig_elasticIPAddress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfshield.ResourceProtection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccShieldProtection_alb(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionConfig_alb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccShieldProtection_elb(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionConfig_elb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccShieldProtection_cloudFront(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionConfig_cloudFront(rName, testAccProtectionCloudFrontRetainConfig()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccShieldProtection_CloudFront_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionConfig_cloudFrontTags1(rName, testAccProtectionCloudFrontRetainConfig(), "Key1", acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProtectionConfig_cloudFrontTags2(rName, testAccProtectionCloudFrontRetainConfig(), "Key1", acctest.CtValue1Updated, "Key2", acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccProtectionConfig_cloudFrontTags1(rName, testAccProtectionCloudFrontRetainConfig(), "Key2", acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccShieldProtection_route53(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_shield_protection.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldEndpointID, "route53"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionConfig_route53HostedZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckProtectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_shield_protection" {
				continue
			}

			_, err := tfshield.FindProtectionByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Shield Protection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProtectionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		_, err := tfshield.FindProtectionByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

	input := &shield.ListProtectionsInput{}

	errResourceNotFoundException := &types.ResourceNotFoundException{}

	_, err := conn.ListProtections(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, errResourceNotFoundException.ErrorCode(), "subscription does not exist") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// Set the environment variable TF_TEST_CLOUDFRONT_RETAIN
// when doing manual tests so that the test is not waiting for
// the distribution to be removed completely.
func testAccProtectionCloudFrontRetainConfig() string {
	if _, ok := os.LookupEnv("TF_TEST_CLOUDFRONT_RETAIN"); ok {
		return "retain_on_delete = true"
	}
	return ""
}

func testAccProtectionConfig_route53HostedZone(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name    = "%[1]s.com."
  comment = "Terraform Acceptance Tests"

  tags = {
    foo  = "bar"
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = "arn:${data.aws_partition.current.partition}:route53:::hostedzone/${aws_route53_zone.test.zone_id}"
}
`, rName)
}

func testAccProtectionConfig_elb(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    foo  = "bar"
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count                   = 2
  vpc_id                  = aws_vpc.test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    foo  = "bar"
    Name = %[1]q
  }
}

resource "aws_elb" "test" {
  name = %[1]q

  subnets  = aws_subnet.test[*].id
  internal = true

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    foo  = "bar"
    Name = %[1]q
  }

  cross_zone_load_balancing = true
}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = aws_elb.test.arn
}
`, rName)
}

func testAccProtectionConfig_alb(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    foo  = "bar"
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    foo  = "bar"
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count                   = 2
  vpc_id                  = aws_vpc.test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    foo  = "bar"
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "test"
  vpc_id      = aws_vpc.test.id

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

  tags = {
    foo  = "bar"
    Name = %[1]q
  }
}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = aws_lb.test.arn
}
`, rName)
}

func testAccProtectionConfig_cloudFront(rName, retainOnDelete string) string {
	return fmt.Sprintf(`
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
    foo  = "bar"
    Name = %[1]q
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = aws_cloudfront_distribution.test.arn

}
`, rName, retainOnDelete)
}

func testAccProtectionConfig_cloudFrontTags1(rName, retainOnDelete, tagKey string, tagValue string) string {
	return fmt.Sprintf(`
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
    foo  = "bar"
    Name = %[1]q
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = aws_cloudfront_distribution.test.arn

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, retainOnDelete, tagKey, tagValue)
}

func testAccProtectionConfig_cloudFrontTags2(rName, retainOnDelete, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
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
    foo  = "bar"
    Name = %[1]q
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  %[2]s
}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = aws_cloudfront_distribution.test.arn

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, retainOnDelete, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccProtectionConfig_elasticIPAddress(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    foo  = "bar"
    Name = %[1]q
  }
}

resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.test.id}"
}
`, rName)
}

func testAccProtectionConfig_globalAccelerator(rName string) string {
	return fmt.Sprintf(`
resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = aws_globalaccelerator_accelerator.test.id
}

resource "aws_globalaccelerator_accelerator" "test" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = true
}
`, rName)
}
