package route53resolver_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRoute53ResolverRuleAssociation_basic(t *testing.T) {
	var assn route53resolver.ResolverRuleAssociation
	resourceNameVpc := "aws_vpc.example"
	resourceNameRule := "aws_route53_resolver_rule.example"
	resourceNameAssoc := "aws_route53_resolver_rule_association.example"
	name := fmt.Sprintf("terraform-testacc-r53-resolver-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAssociationConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleAssociationExists(resourceNameAssoc, &assn),
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

func testAccCheckRuleAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_rule_association" {
			continue
		}

		// Try to find the resource
		_, err := conn.GetResolverRuleAssociation(&route53resolver.GetResolverRuleAssociationInput{
			ResolverRuleAssociationId: aws.String(rs.Primary.ID),
		})
		// Verify the error is what we want
		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver rule association still exists: %s", rs.Primary.ID)
	}
	return nil
}

func testAccCheckRuleAssociationExists(n string, assn *route53resolver.ResolverRuleAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver rule association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
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

func testAccRuleAssociationConfig_basic(name string) string {
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
  resolver_rule_id = aws_route53_resolver_rule.example.id
  vpc_id           = aws_vpc.example.id
}
`, name)
}
