package route53_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAWSRoute53VpcAssociationAuthorization_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_route53_vpc_association_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckRoute53VPCAssociationAuthorizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53VPCAssociationAuthorizationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53VPCAssociationAuthorizationExists(resourceName),
				),
			},
			{
				Config:            testAccRoute53VPCAssociationAuthorizationConfig(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute53VpcAssociationAuthorization_disappears(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_route53_vpc_association_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckRoute53VPCAssociationAuthorizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53VPCAssociationAuthorizationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53VPCAssociationAuthorizationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceVPCAssociationAuthorization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRoute53VPCAssociationAuthorizationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_vpc_association_authorization" {
			continue
		}

		zone_id, vpc_id, err := resourceAwsRoute53VPCAssociationAuthorizationParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		req := route53.ListVPCAssociationAuthorizationsInput{
			HostedZoneId: aws.String(zone_id),
		}

		res, err := conn.ListVPCAssociationAuthorizations(&req)
		if tfawserr.ErrMessageContains(err, route53.ErrCodeNoSuchHostedZone, "") {
			return nil
		}
		if err != nil {
			return err
		}

		for _, vpc := range res.VPCs {
			if vpc_id == *vpc.VPCId {
				return fmt.Errorf("VPC association authorization for zone %v with %v still exists", zone_id, vpc_id)
			}
		}
	}
	return nil
}

func testAccCheckRoute53VPCAssociationAuthorizationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC association authorization ID is set")
		}

		zone_id, vpc_id, err := resourceAwsRoute53VPCAssociationAuthorizationParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Conn

		req := route53.ListVPCAssociationAuthorizationsInput{
			HostedZoneId: aws.String(zone_id),
		}

		res, err := conn.ListVPCAssociationAuthorizations(&req)
		if err != nil {
			return err
		}

		for _, vpc := range res.VPCs {
			if vpc_id == *vpc.VPCId {
				return nil
			}
		}

		return fmt.Errorf("VPC association authorization not found")
	}
}

func testAccRoute53VPCAssociationAuthorizationConfig() string {
	return acctest.ConfigAlternateAccountProvider() + `
resource "aws_vpc" "test" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_route53_zone" "test" {
  name = "example.com"

  vpc {
    vpc_id = aws_vpc.test.id
  }
}

resource "aws_vpc" "alternate" {
  provider             = "awsalternate"
  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_route53_vpc_association_authorization" "test" {
  zone_id = aws_route53_zone.test.id
  vpc_id  = aws_vpc.alternate.id
}
`
}
