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

func TestAccAwsRoute53ResolverRule_basic(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_basicNoTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com."),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAwsRoute53ResolverRule_tags(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_basicTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com."),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverRuleConfig_basicTagsChanged,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com."),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccRoute53ResolverRuleConfig_basicNoTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com."),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAwsRoute53ResolverRule_updateName(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"
	name1 := fmt.Sprintf("terraform-testacc-r53-resolver-%d", acctest.RandInt())
	name2 := fmt.Sprintf("terraform-testacc-r53-resolver-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_basicName(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com."),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverRuleConfig_basicName(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com."),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
				),
			},
		},
	})
}

func TestAccAwsRoute53ResolverRule_forward(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"
	name := fmt.Sprintf("terraform-testacc-r53-resolver-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverRuleConfig_forward(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com."),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.1379138419.ip", "192.0.2.6"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.1379138419.port", "53"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverRuleConfig_forwardChanged(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com."),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.1867764419.ip", "192.0.2.7"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.1867764419.port", "53"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.1677112772.ip", "192.0.2.17"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.1677112772.port", "54"),
				),
			},
		},
	})
}

func testAccCheckRoute53ResolverRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_rule" {
			continue
		}

		// Try to find the resource
		_, err := conn.GetResolverRule(&route53resolver.GetResolverRuleInput{
			ResolverRuleId: aws.String(rs.Primary.ID),
		})
		// Verify the error is what we want
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver rule still exists: %s", rs.Primary.ID)
	}
	return nil
}

func testAccCheckRoute53ResolverRuleExists(n string, rule *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver rule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
		res, err := conn.GetResolverRule(&route53resolver.GetResolverRuleInput{
			ResolverRuleId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*rule = *res.ResolverRule

		return nil
	}
}

const testAccRoute53ResolverRuleConfig_basicNoTags = `
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"
}
`

const testAccRoute53ResolverRuleConfig_basicTags = `
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"

  tags = {
    Environment = "production"
    Usage = "original"
  }
}
`

const testAccRoute53ResolverRuleConfig_basicTagsChanged = `
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"

  tags = {
    Usage = "changed"
  }
}
`

func testAccRoute53ResolverRuleConfig_basicName(name string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"
  name        = %q
}
`, name)
}

func testAccRoute53ResolverRuleConfig_forward(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = "${aws_route53_resolver_endpoint.foo.id}"

  target_ip {
    ip = "192.0.2.6"
  }
}
`, testAccRoute53ResolverRuleConfig_resolverEndpoint(name), name)
}

func testAccRoute53ResolverRuleConfig_forwardChanged(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = "${aws_route53_resolver_endpoint.foo.id}"

  target_ip {
    ip = "192.0.2.7"
  }
  target_ip {
    ip   = "192.0.2.17"
    port = 54
  }
}
`, testAccRoute53ResolverRuleConfig_resolverEndpoint(name), name)
}

func testAccRoute53ResolverRuleConfig_resolverEndpoint(name string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %q
  }
}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "sn1" {
  vpc_id            = "${aws_vpc.foo.id}"
  cidr_block        = "${cidrsubnet(aws_vpc.foo.cidr_block, 2, 0)}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  tags = {
    Name = %q
  }
}

resource "aws_subnet" "sn2" {
  vpc_id            = "${aws_vpc.foo.id}"
  cidr_block        = "${cidrsubnet(aws_vpc.foo.cidr_block, 2, 1)}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"

  tags = {
    Name = %q
  }
}

resource "aws_security_group" "sg1" {
  vpc_id = "${aws_vpc.foo.id}"
  name   = %q

  tags = {
    Name = %q
  }
}

resource "aws_route53_resolver_endpoint" "foo" {
  direction = "OUTBOUND"
  name      = %q

  security_group_ids = [
    "${aws_security_group.sg1.id}",
  ]

  ip_address {
    subnet_id = "${aws_subnet.sn1.id}"
  }

  ip_address {
    subnet_id = "${aws_subnet.sn2.id}"
  }
}
`, name, name, name, name, name, name)
}
