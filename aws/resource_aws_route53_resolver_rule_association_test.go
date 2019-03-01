package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
)

func TestAccAWSRoute53ResolverRuleAssociation_basic(t *testing.T) {
	var assn route53resolver.ResolverRuleAssociation
	resourceName := "aws_route53_resolver_rule_association.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleAssociationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleAssociationExists(resourceName, &assn),
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

func testAccCheckRoute53ResolverRuleAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_rule_association" {
			continue
		}

		req := &route53resolver.GetResolverRuleAssociationInput{
			ResolverRuleAssociationId: aws.String(rs.Primary.ID),
		}

		exists := true
		_, err := conn.GetResolverRuleAssociation(req)
		if err != nil {
			if !isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
				return fmt.Errorf("Error reading Route 53 Resolver rule association: %s", err)
			}
			exists = false
		}

		if exists {
			return fmt.Errorf("Route 53 Resolver rule association still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckRoute53ResolverRuleAssociationExists(n string, assn *route53resolver.ResolverRuleAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).route53resolverconn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resolver rule association ID is set")
		}

		req := &route53resolver.GetResolverRuleAssociationInput{
			ResolverRuleAssociationId: aws.String(rs.Primary.ID),
		}

		res, err := conn.GetResolverRuleAssociation(req)
		if err != nil {
			if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
				return fmt.Errorf("Route 53 Resolver rule association not found")
			}
			return fmt.Errorf("Error reading Route 53 Resolver rule association: %s", err)
		}

		*assn = *res.ResolverRuleAssociation
		return nil
	}
}

const testAccRoute53ResolverRuleAssociationConfig_basic = `
resource "aws_vpc" "example" {
	cidr_block = "10.6.0.0/16"
	enable_dns_hostnames = true
	enable_dns_support = true
	tags {
		Name = "terraform-testacc-route53-zone-association-foo"
	}
}

resource "aws_route53_resolver_rule" "example" {
	domain_name          = "example.com"
	name                 = "example"
	rule_type            = "SYSTEM"
}

resource "aws_route53_resolver_rule_association" "example" {
	resolver_rule_id = "${aws_route53_resolver_rule.example.id}"
	vpc_id  = "${aws_vpc.example.id}"
}
`
