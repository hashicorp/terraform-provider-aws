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

func TestAccRoute53ResolverFirewallRuleGroupAssociation_basic(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_rule_group_id", "aws_route53_resolver_firewall_rule_group.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "mutation_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "priority", "101"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
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

func TestAccRoute53ResolverFirewallRuleGroupAssociation_name(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNewName := sdkacctest.RandomWithPrefix("tf-acc-test2")
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_basic(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rNewName),
				),
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroupAssociation_mutationProtection(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_mutationProtection(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mutation_protection", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_mutationProtection(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mutation_protection", "DISABLED"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroupAssociation_priority(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_priority(rName, 101),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "priority", "101"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_priority(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "priority", "200"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroupAssociation_disappears(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53resolver.ResourceFirewallRuleGroupAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroupAssociation_tags(t *testing.T) {
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallRuleGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckFirewallRuleGroupAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_firewall_rule_group_association" {
			continue
		}

		// Try to find the resource
		_, err := tfroute53resolver.FindFirewallRuleGroupAssociationByID(conn, rs.Primary.ID)
		// Verify the error is what we want
		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver DNS Firewall rule group association still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckFirewallRuleGroupAssociationExists(n string, v *route53resolver.FirewallRuleGroupAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver DNS Firewall rule group association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
		out, err := tfroute53resolver.FindFirewallRuleGroupAssociationByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccFirewallRuleGroupAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccFirewallRuleGroupAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  mutation_protection    = "DISABLED"
  priority               = 101
  vpc_id                 = aws_vpc.test.id
}
`, testAccFirewallRuleGroupAssociationConfig_base(rName), rName)
}

func testAccFirewallRuleGroupAssociationConfig_mutationProtection(rName, mutationProtection string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  mutation_protection    = %[3]q
  priority               = 101
  vpc_id                 = aws_vpc.test.id
}
`, testAccFirewallRuleGroupAssociationConfig_base(rName), rName, mutationProtection)
}

func testAccFirewallRuleGroupAssociationConfig_priority(rName string, priority int) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = %[3]d
  vpc_id                 = aws_vpc.test.id
}
`, testAccFirewallRuleGroupAssociationConfig_base(rName), rName, priority)
}

func testAccFirewallRuleGroupAssociationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = 101
  vpc_id                 = aws_vpc.test.id

  tags = {
    %[3]q = %[4]q
  }
}
`, testAccFirewallRuleGroupAssociationConfig_base(rName), rName, tagKey1, tagValue1)
}

func testAccFirewallRuleGroupAssociationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[2]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = 101
  vpc_id                 = aws_vpc.test.id

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, testAccFirewallRuleGroupAssociationConfig_base(rName), rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
