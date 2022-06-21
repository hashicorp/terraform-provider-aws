package route53_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
)

func TestAccRoute53ZoneAssociation_basic(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(resourceName),
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
	resourceName := "aws_route53_zone_association.test"

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53.ResourceZoneAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_Disappears_vpc(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"
	vpcResourceName := "aws_vpc.bar"

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ec2.ResourceVPC(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_Disappears_zone(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"
	route53ZoneResourceName := "aws_route53_zone.foo"

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53.ResourceZone(), route53ZoneResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_crossAccount(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_crossAccount(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(resourceName),
				),
			},
			{
				Config:            testAccZoneAssociationConfig_crossAccount(domainName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53ZoneAssociation_crossRegion(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZoneAssociationConfig_region(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckZoneAssociationExists(resourceName),
				),
			},
			{
				Config:            testAccZoneAssociationConfig_region(domainName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckZoneAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_zone_association" {
			continue
		}

		zoneID, vpcID, vpcRegion, err := tfroute53.ZoneAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		hostedZoneSummary, err := tfroute53.GetZoneAssociation(conn, zoneID, vpcID, vpcRegion)

		if tfawserr.ErrMessageContains(err, "AccessDenied", "is not owned by you") {
			continue
		}

		if err != nil {
			return err
		}

		if hostedZoneSummary != nil {
			return fmt.Errorf("Route 53 Zone Association (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckZoneAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No zone association ID is set")
		}

		zoneID, vpcID, vpcRegion, err := tfroute53.ZoneAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

		hostedZoneSummary, err := tfroute53.GetZoneAssociation(conn, zoneID, vpcID, vpcRegion)

		if err != nil {
			return err
		}

		if hostedZoneSummary == nil {
			return fmt.Errorf("Route 53 Hosted Zone (%s) Association (%s) not found", zoneID, vpcID)
		}

		return nil
	}
}

func testAccZoneAssociationConfig_basic(domainName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "terraform-testacc-route53-zone-association-foo"
  }
}

resource "aws_vpc" "bar" {
  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "terraform-testacc-route53-zone-association-bar"
  }
}

resource "aws_route53_zone" "foo" {
  name = %[1]q

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
`, domainName)
}

func testAccZoneAssociationConfig_crossAccount(domainName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_vpc" "test" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_route53_zone" "test" {
  provider = "awsalternate"

  name = %[1]q

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
`, domainName))
}

func testAccZoneAssociationConfig_region(domainName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "terraform-testacc-route53-zone-association-region-foo"
  }
}

resource "aws_vpc" "alternate" {
  provider = "awsalternate"

  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "terraform-testacc-route53-zone-association-region-bar"
  }
}

resource "aws_route53_zone" "test" {
  name = %[1]q

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
`, domainName))
}
