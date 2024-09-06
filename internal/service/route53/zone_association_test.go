// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ZoneAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone_association.test"
	domainName := acctest.RandomFQDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(ctx, resourceName),
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

func TestAccRoute53ZoneAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone_association.test"
	domainName := acctest.RandomFQDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceZoneAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_Disappears_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone_association.test"
	vpcResourceName := "aws_vpc.bar"
	domainName := acctest.RandomFQDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, ec2.ResourceVPC(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_Disappears_zone(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone_association.test"
	route53ZoneResourceName := "aws_route53_zone.foo"
	domainName := acctest.RandomFQDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckZoneAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceZone(), route53ZoneResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_crossAccount(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone_association.test"
	domainName := acctest.RandomFQDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckZoneAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_crossAccount(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(ctx, resourceName),
				),
			},
			{
				Config:            testAccZoneAssociationConfig_crossAccount(rName, domainName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_crossRegion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone_association.test"
	domainName := acctest.RandomFQDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckZoneAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_crossRegion(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(ctx, resourceName),
				),
			},
			{
				Config:            testAccZoneAssociationConfig_crossRegion(rName, domainName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_crossAccountAndRegion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_zone_association.test"
	domainName := acctest.RandomFQDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckZoneAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_crossAccountAndRegion(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(ctx, resourceName),
				),
			},
			{
				Config:            testAccZoneAssociationConfig_crossAccountAndRegion(rName, domainName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckZoneAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_zone_association" {
				continue
			}

			_, err := tfroute53.FindZoneAssociationByThreePartKey(ctx, conn, rs.Primary.Attributes["zone_id"], rs.Primary.Attributes[names.AttrVPCID], rs.Primary.Attributes["vpc_region"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Zone Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckZoneAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		_, err := tfroute53.FindZoneAssociationByThreePartKey(ctx, conn, rs.Primary.Attributes["zone_id"], rs.Primary.Attributes[names.AttrVPCID], rs.Primary.Attributes["vpc_region"])

		return err
	}
}

func testAccZoneAssociationConfig_basic(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "bar" {
  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "foo" {
  name = %[2]q

  vpc {
    vpc_id = aws_vpc.foo.id
  }

  lifecycle {
    ignore_changes = ["vpc"]
  }
}

resource "aws_route53_zone_association" "test" {
  zone_id = aws_route53_zone.foo.id
  vpc_id  = aws_vpc.bar.id
}
`, rName, domainName)
}

func testAccZoneAssociationConfig_crossAccount(rName, domainName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "test" {
  provider = "awsalternate"

  name = %[2]q

  vpc {
    vpc_id = aws_vpc.alternate.id
  }

  lifecycle {
    ignore_changes = [vpc]
  }
}

resource "aws_route53_vpc_association_authorization" "test" {
  provider = "awsalternate"

  vpc_id  = aws_vpc.test.id
  zone_id = aws_route53_zone.test.id
}

resource "aws_route53_zone_association" "test" {
  vpc_id  = aws_route53_vpc_association_authorization.test.vpc_id
  zone_id = aws_route53_vpc_association_authorization.test.zone_id
}
`, rName, domainName))
}

func testAccZoneAssociationConfig_crossRegion(rName, domainName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "test" {
  name = %[2]q

  vpc {
    vpc_id     = aws_vpc.test.id
    vpc_region = data.aws_region.current.name
  }

  lifecycle {
    ignore_changes = [vpc]
  }
}

resource "aws_route53_zone_association" "test" {
  vpc_id     = aws_vpc.alternate.id
  vpc_region = data.aws_region.alternate.name
  zone_id    = aws_route53_zone.test.id
}
`, rName, domainName))
}

func testAccZoneAssociationConfig_crossAccountAndRegion(rName, domainName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_route53_zone_association" "test" {
  vpc_id  = aws_route53_vpc_association_authorization.test.vpc_id
  zone_id = aws_route53_vpc_association_authorization.test.zone_id
}

resource "aws_route53_vpc_association_authorization" "test" {
  provider = "awsalternate"

  vpc_id     = aws_vpc.test.id
  zone_id    = aws_route53_zone.test.id
  vpc_region = data.aws_region.current.name
}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_zone" "test" {
  provider = "awsalternate"

  name = %[2]q

  vpc {
    vpc_id = aws_vpc.alternate.id
  }

  lifecycle {
    ignore_changes = [vpc]
  }
}
`, rName, domainName))
}
