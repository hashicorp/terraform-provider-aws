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

func TestAccRoute53ResolverRule_basic(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basicNoTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccRoute53ResolverRule_justDotDomainName(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic("."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "."),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccRoute53ResolverRule_trailingDotDomainName(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic("example.com."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccRoute53ResolverRule_tags(t *testing.T) {
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basicTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
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
				Config: testAccRuleConfig_basicTagsChanged,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccRuleConfig_basicNoTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRule_updateName(t *testing.T) {
	var rule1, rule2 route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"
	name1 := fmt.Sprintf("terraform-testacc-r53-resolver-%d", sdkacctest.RandInt())
	name2 := fmt.Sprintf("terraform-testacc-r53-resolver-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basicName(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
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
				Config: testAccRuleConfig_basicName(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule2),
					testAccCheckRulesSame(&rule2, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRule_forward(t *testing.T) {
	var rule1, rule2, rule3 route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"
	resourceNameEp1 := "aws_route53_resolver_endpoint.foo"
	resourceNameEp2 := "aws_route53_resolver_endpoint.bar"
	name := fmt.Sprintf("terraform-testacc-r53-resolver-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_forward(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp1, "id"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.6",
						"port": "53",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleConfig_forwardTargetIPChanged(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule2),
					testAccCheckRulesSame(&rule2, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp1, "id"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.7",
						"port": "53",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.17",
						"port": "54",
					}),
				),
			},
			{
				Config: testAccRuleConfig_forwardEndpointChanged(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule3),
					testAccCheckRulesSame(&rule3, &rule2),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp2, "id"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.7",
						"port": "53",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.17",
						"port": "54",
					}),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRule_forwardEndpointRecreate(t *testing.T) {
	var rule1, rule2 route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.example"
	resourceNameEp := "aws_route53_resolver_endpoint.foo"
	name := fmt.Sprintf("terraform-testacc-r53-resolver-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_forward(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp, "id"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.6",
						"port": "53",
					}),
				),
			},
			{
				Config: testAccRuleConfig_forwardEndpointRecreate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule2),
					testAccCheckRulesDifferent(&rule2, &rule1),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", resourceNameEp, "id"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":   "192.0.2.6",
						"port": "53",
					}),
				),
			},
		},
	})
}

func testAccCheckRulesSame(before, after *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Arn != *after.Arn {
			return fmt.Errorf("Expected Route 53 Resolver rule ARNs to be the same. But they were: %v, %v", *before.Arn, *after.Arn)
		}
		return nil
	}
}

func testAccCheckRulesDifferent(before, after *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Arn == *after.Arn {
			return fmt.Errorf("Expected Route 53 Resolver rule ARNs to be different. But they were both: %v", *before.Arn)
		}
		return nil
	}
}

func testAccCheckRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_rule" {
			continue
		}

		// Try to find the resource
		_, err := conn.GetResolverRule(&route53resolver.GetResolverRuleInput{
			ResolverRuleId: aws.String(rs.Primary.ID),
		})
		// Verify the error is what we want
		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver rule still exists: %s", rs.Primary.ID)
	}
	return nil
}

func testAccCheckRuleExists(n string, rule *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
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

func testAccRuleConfig_basic(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = %[1]q
  rule_type   = "SYSTEM"
}
`, domainName)
}

const testAccRuleConfig_basicNoTags = `
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"
}
`

const testAccRuleConfig_basicTags = `
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`

const testAccRuleConfig_basicTagsChanged = `
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"

  tags = {
    Usage = "changed"
  }
}
`

func testAccRuleConfig_basicName(name string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "SYSTEM"
  name        = %q
}
`, name)
}

func testAccRuleConfig_forward(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = aws_route53_resolver_endpoint.foo.id

  target_ip {
    ip = "192.0.2.6"
  }
}
`, testAccRuleConfig_resolverEndpoint(name), name)
}

func testAccRuleConfig_forwardTargetIPChanged(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = aws_route53_resolver_endpoint.foo.id

  target_ip {
    ip = "192.0.2.7"
  }

  target_ip {
    ip   = "192.0.2.17"
    port = 54
  }
}
`, testAccRuleConfig_resolverEndpoint(name), name)
}

func testAccRuleConfig_forwardEndpointChanged(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = aws_route53_resolver_endpoint.bar.id

  target_ip {
    ip = "192.0.2.7"
  }

  target_ip {
    ip   = "192.0.2.17"
    port = 54
  }
}
`, testAccRuleConfig_resolverEndpoint(name), name)
}

func testAccRuleConfig_forwardEndpointRecreate(name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_rule" "example" {
  domain_name = "example.com"
  rule_type   = "FORWARD"
  name        = %q

  resolver_endpoint_id = aws_route53_resolver_endpoint.foo.id

  target_ip {
    ip = "192.0.2.6"
  }
}
`, testAccRuleConfig_resolverEndpointRecreate(name), name)
}

func testAccRuleConfig_resolverVPC(name string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "sn1" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 0)
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s_1"
  }
}

resource "aws_subnet" "sn2" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 1)
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "%[1]s_2"
  }
}

resource "aws_subnet" "sn3" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 2)
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = "%[1]s_3"
  }
}

resource "aws_security_group" "sg1" {
  vpc_id = aws_vpc.foo.id
  name   = "%[1]s_1"

  tags = {
    Name = "%[1]s_1"
  }
}

resource "aws_security_group" "sg2" {
  vpc_id = aws_vpc.foo.id
  name   = "%[1]s_2"

  tags = {
    Name = "%[1]s_2"
  }
}
`, name)
}

func testAccRuleConfig_resolverEndpoint(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_endpoint" "foo" {
  direction = "OUTBOUND"
  name      = "%[2]s_1"

  security_group_ids = [
    aws_security_group.sg1.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn2.id
  }
}

resource "aws_route53_resolver_endpoint" "bar" {
  direction = "OUTBOUND"
  name      = "%[2]s_2"

  security_group_ids = [
    aws_security_group.sg1.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn3.id
  }
}
`, testAccRuleConfig_resolverVPC(name), name)
}

func testAccRuleConfig_resolverEndpointRecreate(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_endpoint" "foo" {
  direction = "OUTBOUND"
  name      = "%[2]s_1"

  security_group_ids = [
    aws_security_group.sg2.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn2.id
  }
}

resource "aws_route53_resolver_endpoint" "bar" {
  direction = "OUTBOUND"
  name      = "%[2]s_2"

  security_group_ids = [
    aws_security_group.sg1.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn3.id
  }
}
`, testAccRuleConfig_resolverVPC(name), name)
}
