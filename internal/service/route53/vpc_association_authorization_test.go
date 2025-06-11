// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53VPCAssociationAuthorization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_vpc_association_authorization.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckVPCAssociationAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAssociationAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAssociationAuthorizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_region", acctest.Region()),
				),
			},
			{
				Config:            testAccVPCAssociationAuthorizationConfig_basic(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53VPCAssociationAuthorization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_vpc_association_authorization.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckVPCAssociationAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAssociationAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAssociationAuthorizationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceVPCAssociationAuthorization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53VPCAssociationAuthorization_concurrent(t *testing.T) {
	ctx := acctest.Context(t)
	resourceNameAlternate := "aws_route53_vpc_association_authorization.alternate"
	resourceNameThird := "aws_route53_vpc_association_authorization.third"
	providers := make(map[string]*schema.Provider)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckThirdAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamed(ctx, t, providers, acctest.ProviderName, acctest.ProviderNameAlternate, acctest.ProviderNameThird),
		CheckDestroy:             testAccCheckVPCAssociationAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAssociationAuthorizationConfig_concurrent(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAssociationAuthorizationExists(ctx, resourceNameAlternate),
					testAccCheckVPCAssociationAuthorizationExists(ctx, resourceNameThird),
				),
			},
		},
	})
}

func TestAccRoute53VPCAssociationAuthorization_crossRegion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_vpc_association_authorization.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckVPCAssociationAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAssociationAuthorizationConfig_crossRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAssociationAuthorizationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_region", acctest.AlternateRegion()),
				),
			},
			{
				Config:            testAccVPCAssociationAuthorizationConfig_crossRegion(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCAssociationAuthorizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_vpc_association_authorization" {
				continue
			}

			_, err := tfroute53.FindVPCAssociationAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes["zone_id"], rs.Primary.Attributes[names.AttrVPCID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 VPC Association Authorization %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCAssociationAuthorizationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		_, err := tfroute53.FindVPCAssociationAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes["zone_id"], rs.Primary.Attributes[names.AttrVPCID])

		return err
	}
}

func testAccVPCAssociationAuthorizationConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_route53_vpc_association_authorization" "test" {
  zone_id = aws_route53_zone.test.id
  vpc_id  = aws_vpc.alternate.id
}

resource "aws_vpc" "alternate" {
  provider             = "awsalternate"
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 0)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "test" {
  name = "example.com"

  vpc {
    vpc_id = aws_vpc.test.id
  }
}

resource "aws_vpc" "test" {
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 1)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCAssociationAuthorizationConfig_concurrent(t *testing.T, rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleAccountProvider(t, 3), fmt.Sprintf(`
resource "aws_route53_vpc_association_authorization" "alternate" {
  zone_id = aws_route53_zone.test.id
  vpc_id  = aws_vpc.alternate.id

  # Try to encourage concurrency
  depends_on = [
    aws_vpc.alternate,
    aws_vpc.third
  ]
}

resource "aws_route53_vpc_association_authorization" "third" {
  zone_id = aws_route53_zone.test.id
  vpc_id  = aws_vpc.third.id

  # Try to encourage concurrency
  depends_on = [
    aws_vpc.alternate,
    aws_vpc.third
  ]
}

resource "aws_route53_zone" "test" {
  name = "example.com"

  vpc {
    vpc_id = aws_vpc.test.id
  }
}

resource "aws_vpc" "test" {
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 0)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "alternate" {
  provider             = "awsalternate"
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 1)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "third" {
  provider             = "awsthird"
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 2)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCAssociationAuthorizationConfig_crossRegion(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_route53_vpc_association_authorization" "test" {
  zone_id    = aws_route53_zone.test.id
  vpc_id     = aws_vpc.alternate.id
  vpc_region = data.aws_region.alternate.name
}

resource "aws_vpc" "alternate" {
  provider             = "awsalternate"
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 1)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_route53_zone" "test" {
  name = "example.com"

  vpc {
    vpc_id = aws_vpc.test.id
  }
}

resource "aws_vpc" "test" {
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 0)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
