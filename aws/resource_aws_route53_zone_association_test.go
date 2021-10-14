package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSRoute53ZoneAssociation_basic(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig(domainName),
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
	resourceName := "aws_route53_zone_association.test"

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsRoute53ZoneAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ZoneAssociation_disappears_VPC(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"
	vpcResourceName := "aws_vpc.bar"

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsVpc(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ZoneAssociation_disappears_Zone(t *testing.T) {
	resourceName := "aws_route53_zone_association.test"
	route53ZoneResourceName := "aws_route53_zone.foo"

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsRoute53Zone(), route53ZoneResourceName),
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

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationCrossAccountConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
				),
			},
			{
				Config:            testAccRoute53ZoneAssociationCrossAccountConfig(domainName),
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

	domainName := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationRegionConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName),
				),
			},
			{
				Config:            testAccRoute53ZoneAssociationRegionConfig(domainName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRoute53ZoneAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).r53conn
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

		conn := testAccProvider.Meta().(*AWSClient).r53conn

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

func testAccRoute53ZoneAssociationConfig(domainName string) string {
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

func testAccRoute53ZoneAssociationCrossAccountConfig(domainName string) string {
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

func testAccRoute53ZoneAssociationRegionConfig(domainName string) string {
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
