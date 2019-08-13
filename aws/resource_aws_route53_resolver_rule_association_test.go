package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
)

func TestAccAwsRoute53ResolverRuleAssociation_basic(t *testing.T) {
	var assn route53resolver.ResolverRuleAssociation
	resourceNameVpc := "aws_vpc.example"
	resourceNameRule := "aws_route53_resolver_rule.example"
	resourceNameAssoc := "aws_route53_resolver_rule_association.example"
	name := fmt.Sprintf("terraform-testacc-r53-resolver-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleAssociationConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleAssociationExists(resourceNameAssoc, &assn),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "vpc_id", resourceNameVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "resolver_rule_id", resourceNameRule, "id"),
					resource.TestCheckResourceAttr(resourceNameAssoc, "name", name),
				),
			},
			{
				ResourceName:      resourceNameAssoc,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRoute53ResolverRuleAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_rule_association" {
			continue
		}

		// Try to find the resource
		_, err := conn.GetResolverRuleAssociation(&route53resolver.GetResolverRuleAssociationInput{
			ResolverRuleAssociationId: aws.String(rs.Primary.ID),
		})
		// Verify the error is what we want
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver rule association still exists: %s", rs.Primary.ID)
	}
	return nil
}

func testAccCheckRoute53ResolverRuleAssociationExists(n string, assn *route53resolver.ResolverRuleAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver rule association ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
		resp, err := conn.GetResolverRuleAssociation(&route53resolver.GetResolverRuleAssociationInput{
			ResolverRuleAssociationId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*assn = *resp.ResolverRuleAssociation

		return nil
	}
}

func testAccRoute53ResolverRuleAssociationConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "example" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  name        = %[1]q
  rule_type   = "SYSTEM"
}

resource "aws_route53_resolver_rule_association" "example" {
  name             = %[1]q
  resolver_rule_id = "${aws_route53_resolver_rule.example.id}"
  vpc_id           = "${aws_vpc.example.id}"
}
`, name)
}
