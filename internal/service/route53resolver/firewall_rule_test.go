package route53resolver_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
)

func TestAccRoute53ResolverFirewallRule_basic(t *testing.T) {
	var v route53resolver.FirewallRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "action", "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_rule_group_id", "aws_route53_resolver_firewall_rule_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_domain_list_id", "aws_route53_resolver_firewall_domain_list.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "priority", "100"),
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

func TestAccRoute53ResolverFirewallRule_block(t *testing.T) {
	var v route53resolver.FirewallRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_block(rName, "NODATA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "action", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "block_response", "NODATA"),
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

func TestAccRoute53ResolverFirewallRule_blockOverride(t *testing.T) {
	var v route53resolver.FirewallRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_blockOverride(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "action", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "block_override_dns_type", "CNAME"),
					resource.TestCheckResourceAttr(resourceName, "block_override_domain", "example.com."),
					resource.TestCheckResourceAttr(resourceName, "block_override_ttl", "60"),
					resource.TestCheckResourceAttr(resourceName, "block_response", "OVERRIDE"),
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

func TestAccRoute53ResolverFirewallRule_disappears(t *testing.T) {
	var v route53resolver.FirewallRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53resolver.ResourceFirewallRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFirewallRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_firewall_rule" {
			continue
		}

		// Try to find the resource
		_, err := tfroute53resolver.FindFirewallRuleByID(conn, rs.Primary.ID)
		// Verify the error is what we want
		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver DNS Firewall rule still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckFirewallRuleExists(n string, v *route53resolver.FirewallRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver DNS Firewall rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
		out, err := tfroute53resolver.FindFirewallRuleByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccFirewallRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_domain_list" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_rule" "test" {
  name                    = %[1]q
  action                  = "ALLOW"
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.test.id
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.test.id
  priority                = 100
}
`, rName)
}

func testAccFirewallRuleConfig_block(rName, blockResponse string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_domain_list" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_rule" "test" {
  name                    = %[1]q
  action                  = "BLOCK"
  block_response          = %[2]q
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.test.id
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.test.id
  priority                = 100
}
`, rName, blockResponse)
}

func testAccFirewallRuleConfig_blockOverride(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_domain_list" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_rule" "test" {
  name                    = %[1]q
  action                  = "BLOCK"
  block_override_dns_type = "CNAME"
  block_override_domain   = "example.com."
  block_override_ttl      = 60
  block_response          = "OVERRIDE"
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.test.id
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.test.id
  priority                = 100
}
`, rName)
}
