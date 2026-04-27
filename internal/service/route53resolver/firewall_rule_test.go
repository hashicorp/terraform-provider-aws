// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverFirewallRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.FirewallRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_rule_group_id", "aws_route53_resolver_firewall_rule_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_domain_list_id", "aws_route53_resolver_firewall_domain_list.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "firewall_domain_redirection_action", "INSPECT_REDIRECTION_DOMAIN"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "100"),
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

func TestAccRoute53ResolverFirewallRule_update_firewallDomainRedirectionAction(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.FirewallRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "firewall_domain_redirection_action", "INSPECT_REDIRECTION_DOMAIN"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallRuleConfig_firewallDomainRedirectionAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "firewall_domain_redirection_action", "TRUST_REDIRECTION_DOMAIN"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRule_block(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.FirewallRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_block(rName, "NODATA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "BLOCK"),
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
	ctx := acctest.Context(t)
	var v awstypes.FirewallRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_blockOverride(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "BLOCK"),
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

func TestAccRoute53ResolverFirewallRule_qType(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.FirewallRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_qType(rName, "A"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "ALLOW"),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_rule_group_id", "aws_route53_resolver_firewall_rule_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_domain_list_id", "aws_route53_resolver_firewall_domain_list.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "100"),
					resource.TestCheckResourceAttr(resourceName, "q_type", "A"),
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
	ctx := acctest.Context(t)
	var v awstypes.FirewallRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53resolver.ResourceFirewallRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFirewallRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53ResolverClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_resolver_firewall_rule" {
				continue
			}

			firewallRuleGroupID, firewallDomainListID, err := tfroute53resolver.FirewallRuleParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfroute53resolver.FindFirewallRuleByTwoPartKey(ctx, conn, firewallRuleGroupID, firewallDomainListID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Resolver Firewall Rule still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFirewallRuleExists(ctx context.Context, t *testing.T, n string, v *awstypes.FirewallRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Resolver Firewall Rule ID is set")
		}

		firewallRuleGroupID, firewallDomainListID, err := tfroute53resolver.FirewallRuleParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).Route53ResolverClient(ctx)

		output, err := tfroute53resolver.FindFirewallRuleByTwoPartKey(ctx, conn, firewallRuleGroupID, firewallDomainListID)

		if err != nil {
			return err
		}

		*v = *output

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

func testAccFirewallRuleConfig_firewallDomainRedirectionAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_domain_list" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_rule" "test" {
  name                               = %[1]q
  action                             = "ALLOW"
  firewall_rule_group_id             = aws_route53_resolver_firewall_rule_group.test.id
  firewall_domain_list_id            = aws_route53_resolver_firewall_domain_list.test.id
  firewall_domain_redirection_action = "TRUST_REDIRECTION_DOMAIN"
  priority                           = 100
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

func testAccFirewallRuleConfig_qType(rName, qType string) string {
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
  q_type                  = %[2]q
}
`, rName, qType)
}

func TestAccRoute53ResolverFirewallRule_dnsThreatProtectionDGA(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.FirewallRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_dnsThreatProtection(rName, "DGA", "HIGH"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "dns_threat_protection", "DGA"),
					resource.TestCheckResourceAttr(resourceName, "confidence_threshold", "HIGH"),
					resource.TestCheckResourceAttr(resourceName, "block_response", "NODATA"),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_rule_group_id", "aws_route53_resolver_firewall_rule_group.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "100"),
					resource.TestCheckResourceAttrSet(resourceName, "firewall_threat_protection_id"),
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

func TestAccRoute53ResolverFirewallRule_dnsThreatProtectionDNSTunneling(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.FirewallRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_dnsThreatProtection(rName, "DNS_TUNNELING", "MEDIUM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "dns_threat_protection", "DNS_TUNNELING"),
					resource.TestCheckResourceAttr(resourceName, "confidence_threshold", "MEDIUM"),
					resource.TestCheckResourceAttr(resourceName, "block_response", "NODATA"),
					resource.TestCheckResourceAttrSet(resourceName, "firewall_threat_protection_id"),
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

func TestAccRoute53ResolverFirewallRule_dnsThreatProtectionAlert(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.FirewallRule
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleConfig_dnsThreatProtectionAlert(rName, "DGA", "LOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "ALERT"),
					resource.TestCheckResourceAttr(resourceName, "dns_threat_protection", "DGA"),
					resource.TestCheckResourceAttr(resourceName, "confidence_threshold", "LOW"),
					resource.TestCheckResourceAttrSet(resourceName, "firewall_threat_protection_id"),
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

func testAccFirewallRuleConfig_dnsThreatProtection(rName, dnsThreatProtection, confidenceThreshold string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_rule" "test" {
  name                   = %[1]q
  action                 = "BLOCK"
  block_response         = "NODATA"
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  dns_threat_protection  = %[2]q
  confidence_threshold   = %[3]q
  priority               = 100
}
`, rName, dnsThreatProtection, confidenceThreshold)
}

func testAccFirewallRuleConfig_dnsThreatProtectionAlert(rName, dnsThreatProtection, confidenceThreshold string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_rule" "test" {
  name                   = %[1]q
  action                 = "ALERT"
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  dns_threat_protection  = %[2]q
  confidence_threshold   = %[3]q
  priority               = 100
}
`, rName, dnsThreatProtection, confidenceThreshold)
}
