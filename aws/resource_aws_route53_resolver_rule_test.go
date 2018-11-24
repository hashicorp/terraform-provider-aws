package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
)

func TestAccAWSRoute53ResolverRule_basic(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
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

func TestAccAWSRoute53ResolverRule_updateRule(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
				),
			},
			{
				Config: testAccRoute53ResolverRuleConfig_updateRule,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
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

func TestAccAWSRoute53ResolverRule_updateTags(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
				),
			},
			{
				Config: testAccRoute53ResolverRuleConfig_updateTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
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

func testAccCheckRoute53ResolverRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).r53resolverconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_rule" {
			continue
		}

		req := &route53resolver.GetResolverRuleInput{
			ResolverRuleId: aws.String(rs.Primary.ID),
		}

		exists := true
		_, err := conn.GetResolverRule(req)
		if err != nil {
			if !isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
				return fmt.Errorf("Error reading Route 53 Resolver rule: %s", err)
			}
			exists = false
		}

		if exists {
			return fmt.Errorf("Route 53 Resolver rule still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckRoute53ResolverRuleExists(n string, rule *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).r53resolverconn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resolver rule ID is set")
		}

		req := &route53resolver.GetResolverRuleInput{
			ResolverRuleId: aws.String(rs.Primary.ID),
		}

		res, err := conn.GetResolverRule(req)
		if err != nil {
			if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
				return fmt.Errorf("Route 53 Resolver rule not found")
			}
			return fmt.Errorf("Error reading Route 53 Resolver rule: %s", err)
		}

		*rule = *res.ResolverRule
		return nil
	}
}

const testAccRoute53ResolverRuleConfig_basic = `
resource "aws_route53_resolver_rule" "example" {
	domain_name          = "example.com"
	name                 = "example"
	rule_type            = "SYSTEM"

	tags {
	  Foo = "Barr"
	}
  }
`

const testAccRoute53ResolverRuleConfig_updateRule = `
resource "aws_route53_resolver_rule" "example" {
	domain_name          = "example.com"
	name                 = "test"
	rule_type            = "SYSTEM"

	tags {
	  Foo = "Barr"
	}
  }
`

const testAccRoute53ResolverRuleConfig_updateTags = `
resource "aws_route53_resolver_rule" "example" {
	domain_name          = "example.com"
	name                 = "test"
	rule_type            = "SYSTEM"

	tags {
	  Bar = "Foo"
	}
  }
`
