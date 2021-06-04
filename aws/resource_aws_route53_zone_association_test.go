package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/atest"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func TestAccAWSRoute53ZoneAssociation_basic(t *testing.T) {
	resourceName := "aws_route53_zone_association.foobar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t) },
		ErrorCheck:   atest.ErrorCheck(t, route53.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
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

func TestAccAWSRoute53ZoneAssociation_disappears(t *testing.T) {
	resourceName := "aws_route53_zone_association.foobar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t) },
		ErrorCheck:   atest.ErrorCheck(t, route53.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
					atest.CheckDisappears(atest.Provider, resourceAwsRoute53ZoneAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ZoneAssociation_disappears_VPC(t *testing.T) {
	resourceName := "aws_route53_zone_association.foobar"
	vpcResourceName := "aws_vpc.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t) },
		ErrorCheck:   atest.ErrorCheck(t, route53.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
					atest.CheckDisappears(atest.Provider, resourceAwsVpc(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ZoneAssociation_disappears_Zone(t *testing.T) {
	resourceName := "aws_route53_zone_association.foobar"
	route53ZoneResourceName := "aws_route53_zone.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t) },
		ErrorCheck:   atest.ErrorCheck(t, route53.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
					atest.CheckDisappears(atest.Provider, resourceAwsRoute53Zone(), route53ZoneResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ZoneAssociation_CrossAccount(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			atest.PreCheck(t)
			atest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        atest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: atest.ProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationCrossAccountConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
				),
			},
			{
				Config:            testAccRoute53ZoneAssociationCrossAccountConfig(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute53ZoneAssociation_CrossRegion(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			atest.PreCheck(t)
			atest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        atest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: atest.ProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationRegionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
				),
			},
			{
				Config:            testAccRoute53ZoneAssociationRegionConfig(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRoute53ZoneAssociationDestroy(s *terraform.State) error {
	conn := atest.Provider.Meta().(*awsprovider.AWSClient).Route53Conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_zone_association" {
			continue
		}

		zoneID, vpcID, vpcRegion, err := resourceAwsRoute53ZoneAssociationParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		hostedZoneSummary, err := route53GetZoneAssociation(conn, zoneID, vpcID, vpcRegion)

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

func testAccCheckRoute53ZoneAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No zone association ID is set")
		}

		zoneID, vpcID, vpcRegion, err := resourceAwsRoute53ZoneAssociationParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := atest.Provider.Meta().(*awsprovider.AWSClient).Route53Conn

		hostedZoneSummary, err := route53GetZoneAssociation(conn, zoneID, vpcID, vpcRegion)

		if err != nil {
			return err
		}

		if hostedZoneSummary == nil {
			return fmt.Errorf("Route 53 Hosted Zone (%s) Association (%s) not found", zoneID, vpcID)
		}

		return nil
	}
}

const testAccRoute53ZoneAssociationConfig = `
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
  name = "foo.com"

  vpc {
    vpc_id = aws_vpc.foo.id
  }

  lifecycle {
    ignore_changes = ["vpc"]
  }
}

resource "aws_route53_zone_association" "foobar" {
  zone_id = aws_route53_zone.foo.id
  vpc_id  = aws_vpc.bar.id
}
`

func testAccRoute53ZoneAssociationCrossAccountConfig() string {
	return atest.ComposeConfig(
		atest.ConfigProviderAlternateAccount(),
		`
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

  name = "foo.com"

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
`)
}

func testAccRoute53ZoneAssociationRegionConfig() string {
	return atest.ComposeConfig(
		atest.ConfigProviderMultipleRegion(2),
		`
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
  name = "foo.com"

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
`)
}
