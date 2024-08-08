// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"fmt"
	"os"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccShieldProtectionDataSource_route53HostedZone(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(10)

	ds1ResourceName := "data.aws_shield_protection.test"
	protectionResourceName := "aws_shield_protection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID, "route53"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionDataSource_route53HostedZoneByARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
			{
				Config: testAccProtectionDataSource_route53HostedZoneById(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccShieldProtectionDataSource_cloudfront(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(10)

	ds1ResourceName := "data.aws_shield_protection.test"
	protectionResourceName := "aws_shield_protection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			acctest.PreCheckPartitionHasService(t, "cloudfront")
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionDataSource_cloudfrontByARN(rName, testAccProtectionDataSourceCloudFrontRetainConfig()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
			{
				Config: testAccProtectionDataSource_cloudfrontById(rName, testAccProtectionDataSourceCloudFrontRetainConfig()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccShieldProtectionDataSource_alb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(10)

	ds1ResourceName := "data.aws_shield_protection.test"
	protectionResourceName := "aws_shield_protection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID, "route53"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionDataSource_albByARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
			{
				Config: testAccProtectionDataSource_albById(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccShieldProtectionDataSource_elasticIPAddress(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(10)

	ds1ResourceName := "data.aws_shield_protection.test"
	protectionResourceName := "aws_shield_protection.test"

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
				Config: testAccProtectionDataSource_elasticIPAddressByARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
			{
				Config: testAccProtectionDataSource_elasticIPAddressById(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccShieldProtectionDataSource_globalAccelerator(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(10)

	ds1ResourceName := "data.aws_shield_protection.test"
	protectionResourceName := "aws_shield_protection.test"

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
				Config: testAccProtectionDataSource_globalAcceleratorByARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
			{
				Config: testAccProtectionDataSource_globalAcceleratorById(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccProtectionDataSourceConfig_route53HostedZone(hostedZoneName string) string {
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
  name         = "%[1]s"
  resource_arn = "arn:${data.aws_partition.current.partition}:route53:::hostedzone/${aws_route53_zone.test.zone_id}"
}
`, hostedZoneName)
}

func testAccProtectionDataSource_route53HostedZoneByARN(hostedZoneName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_route53HostedZone(hostedZoneName),
		`
data "aws_shield_protection" "test" {
  resource_arn = "arn:${data.aws_partition.current.partition}:route53:::hostedzone/${aws_route53_zone.test.zone_id}"

  depends_on = [
    aws_shield_protection.test
  ]
}
`)
}

func testAccProtectionDataSource_route53HostedZoneById(hostedZoneName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_route53HostedZone(hostedZoneName),
		`
data "aws_shield_protection" "test" {
  protection_id = aws_shield_protection.test.id
}
`)
}

func TestAccShieldProtectionDataSource_elb(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(10)

	ds1ResourceName := "data.aws_shield_protection.test"
	protectionResourceName := "aws_shield_protection.test"

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
				Config: testAccProtectionDataSource_elbByARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
			{
				Config: testAccProtectionDataSource_elbById(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "protection_id", protectionResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccProtectionDataSourceConfig_elb(rName string) string {
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

func testAccProtectionDataSource_elbByARN(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_elb(rName),
		`
data "aws_shield_protection" "test" {
  resource_arn = aws_elb.test.arn

  depends_on = [
    aws_shield_protection.test
  ]
}
`)
}

func testAccProtectionDataSource_elbById(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_elb(rName),
		`		
data "aws_shield_protection" "test" {
  protection_id = aws_shield_protection.test.id
}
`)
}

func testAccProtectionDataSourceConfig_alb(rName string) string {
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

func testAccProtectionDataSource_albByARN(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_alb(rName),
		`
data "aws_shield_protection" "test" {
  resource_arn = aws_lb.test.arn

  depends_on = [
    aws_shield_protection.test
  ]
}
`)
}

func testAccProtectionDataSource_albById(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_alb(rName),
		`
data "aws_shield_protection" "test" {
  protection_id = aws_shield_protection.test.id
}
`)
}

// Set the environment variable TF_TEST_CLOUDFRONT_RETAIN
// when doing manual tests so that the test is not waiting for
// the distribution to be removed completely.
func testAccProtectionDataSourceCloudFrontRetainConfig() string {
	if _, ok := os.LookupEnv("TF_TEST_CLOUDFRONT_RETAIN"); ok {
		return "retain_on_delete = true"
	}
	return ""
}

func testAccProtectionDataSourceConfig_cloudfront(rName string, retainOnDelete string) string {
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

func testAccProtectionDataSource_cloudfrontByARN(rName string, retainOnDelete string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_cloudfront(rName, retainOnDelete),
		`
data "aws_shield_protection" "test" {
  resource_arn = aws_cloudfront_distribution.test.arn

  depends_on = [
    aws_shield_protection.test
  ]
}
`)
}

func testAccProtectionDataSource_cloudfrontById(rName string, retainOnDelete string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_cloudfront(rName, retainOnDelete),
		`
data "aws_shield_protection" "test" {
  protection_id = aws_shield_protection.test.id
}
`)
}

func testAccProtectionDataSourceConfig_elasticIPAddress(rName string) string {
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

func testAccProtectionDataSource_elasticIPAddressByARN(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_elasticIPAddress(rName),
		`
data "aws_shield_protection" "test" {
  resource_arn = "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.test.id}"

  depends_on = [
    aws_shield_protection.test
  ]
}
`)
}

func testAccProtectionDataSource_elasticIPAddressById(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_elasticIPAddress(rName),
		`
data "aws_shield_protection" "test" {
  protection_id = aws_shield_protection.test.id
}
`)
}

func testAccProtectionDataSourceConfig_globalAccelerator(rName string) string {
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

func testAccProtectionDataSource_globalAcceleratorByARN(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_globalAccelerator(rName),
		`
data "aws_shield_protection" "test" {
  resource_arn = aws_globalaccelerator_accelerator.test.id

  depends_on = [
    aws_shield_protection.test
  ]
}
`)
}

func testAccProtectionDataSource_globalAcceleratorById(rName string) string {
	return acctest.ConfigCompose(
		testAccProtectionDataSourceConfig_globalAccelerator(rName),
		`
data "aws_shield_protection" "test" {
  protection_id = aws_shield_protection.test.id
}
`)
}
