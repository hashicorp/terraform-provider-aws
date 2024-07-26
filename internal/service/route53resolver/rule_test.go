// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rule route53resolver.ResolverRule
	domainName := acctest.RandomDomainName()
	resourceName := "aws_route53_resolver_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "resolver_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", acctest.Ct0),
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

func TestAccRoute53ResolverRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rule route53resolver.ResolverRule
	domainName := acctest.RandomDomainName()
	resourceName := "aws_route53_resolver_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53resolver.ResourceRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ResolverRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var rule route53resolver.ResolverRule
	domainName := acctest.RandomDomainName()
	resourceName := "aws_route53_resolver_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_tags1(domainName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleConfig_tags2(domainName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRuleConfig_tags1(domainName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRule_justDotDomainName(t *testing.T) {
	ctx := acctest.Context(t)
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic("."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, "."),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	ctx := acctest.Context(t)
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic("example.com."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, "example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "SYSTEM"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccRoute53ResolverRule_updateName(t *testing.T) {
	ctx := acctest.Context(t)
	var rule1, rule2 route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.test"
	domainName := acctest.RandomDomainName()
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_name(rName1, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleConfig_name(rName2, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule2),
					testAccCheckRulesSame(&rule2, &rule1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRule_forward(t *testing.T) {
	ctx := acctest.Context(t)
	var rule1, rule2, rule3 route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.test"
	ep1ResourceName := "aws_route53_resolver_endpoint.test.0"
	ep2ResourceName := "aws_route53_resolver_endpoint.test.1"
	domainName := acctest.RandomDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_forward(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", ep1ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":           "192.0.2.6",
						names.AttrPort: "53",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleConfig_forwardTargetIPChanged(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule2),
					testAccCheckRulesSame(&rule2, &rule1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", ep1ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":           "192.0.2.7",
						names.AttrPort: "53",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":           "192.0.2.17",
						names.AttrPort: "54",
					}),
				),
			},
			{
				Config: testAccRuleConfig_forwardEndpointChanged(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule3),
					testAccCheckRulesSame(&rule3, &rule2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", ep2ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":           "192.0.2.7",
						names.AttrPort: "53",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":           "192.0.2.17",
						names.AttrPort: "54",
					}),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRule_forwardMultiProtocol(t *testing.T) {
	ctx := acctest.Context(t)
	var rule route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.test"
	epResourceName := "aws_route53_resolver_endpoint.test.0"
	domainName := acctest.RandomDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_forward(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", epResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":               "192.0.2.6",
						names.AttrPort:     "53",
						names.AttrProtocol: "Do53",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleConfig_forwardMultiProtocol(rName, domainName, "DoH"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", epResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":               "192.0.2.6",
						names.AttrPort:     "53",
						names.AttrProtocol: "DoH",
					}),
				),
			},
			{
				Config: testAccRuleConfig_forwardMultiProtocol(rName, domainName, "Do53"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", epResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":               "192.0.2.6",
						names.AttrPort:     "53",
						names.AttrProtocol: "Do53",
					}),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRule_forwardEndpointRecreate(t *testing.T) {
	ctx := acctest.Context(t)
	var rule1, rule2 route53resolver.ResolverRule
	resourceName := "aws_route53_resolver_rule.test"
	epResourceName := "aws_route53_resolver_endpoint.test.0"
	domainName := acctest.RandomDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_forward(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", epResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":           "192.0.2.6",
						names.AttrPort: "53",
					}),
				),
			},
			{
				Config: testAccRuleConfig_forwardEndpointRecreate(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule2),
					testAccCheckRulesDifferent(&rule2, &rule1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "rule_type", "FORWARD"),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_endpoint_id", epResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "target_ip.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_ip.*", map[string]string{
						"ip":           "192.0.2.6",
						names.AttrPort: "53",
					}),
				),
			},
		},
	})
}

func testAccCheckRulesSame(before, after *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.Arn), aws.StringValue(after.Arn); before != after {
			return fmt.Errorf("Expected Route53 Resolver Rule ARNs to be the same. But they were: %s, %s", before, after)
		}

		return nil
	}
}

func testAccCheckRulesDifferent(before, after *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.Arn), aws.StringValue(after.Arn); before == after {
			return fmt.Errorf("Expected Route53 Resolver rule ARNs to be different. But they were both: %s", before)
		}

		return nil
	}
}

func testAccCheckRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_resolver_rule" {
				continue
			}

			_, err := tfroute53resolver.FindResolverRuleByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Resolver Rule still exists: %s", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckRuleExists(ctx context.Context, n string, v *route53resolver.ResolverRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Resolver Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		output, err := tfroute53resolver.FindResolverRuleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRuleConfig_basic(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[1]q
  rule_type   = "SYSTEM"
}
`, domainName)
}

func testAccRuleConfig_tags1(domainName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[1]q
  rule_type   = "SYSTEM"

  tags = {
    %[2]q = %[3]q
  }
}
`, domainName, tagKey1, tagValue1)
}

func testAccRuleConfig_tags2(domainName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[1]q
  rule_type   = "SYSTEM"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, domainName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRuleConfig_name(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "SYSTEM"
  name        = %[1]q
}
`, rName, domainName)
}

func testAccRuleConfig_forward(rName, domainName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_resolverEndpointBase(rName), fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.test[0].id

  target_ip {
    ip = "192.0.2.6"
  }
}
`, rName, domainName))
}

func testAccRuleConfig_forwardMultiProtocol(rName, domainName, protocol string) string {
	return acctest.ConfigCompose(testAccRuleConfig_resolverEndpointMultiProtocolBase(rName), fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.test[0].id

  target_ip {
    ip       = "192.0.2.6"
    protocol = %[3]q
  }
}
`, rName, domainName, protocol))
}

func testAccRuleConfig_forwardTargetIPChanged(rName, domainName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_resolverEndpointBase(rName), fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.test[0].id

  target_ip {
    ip = "192.0.2.7"
  }

  target_ip {
    ip   = "192.0.2.17"
    port = 54
  }
}
`, rName, domainName))
}

func testAccRuleConfig_forwardEndpointChanged(rName, domainName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_resolverEndpointBase(rName), fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.test[1].id

  target_ip {
    ip = "192.0.2.7"
  }

  target_ip {
    ip   = "192.0.2.17"
    port = 54
  }
}
`, rName, domainName))
}

func testAccRuleConfig_forwardEndpointRecreate(rName, domainName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_resolverEndpointRecreateBase(rName), fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.test[0].id

  target_ip {
    ip = "192.0.2.6"
  }
}
`, rName, domainName))
}

func testAccRuleConfig_vpcBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id
  name   = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccRuleConfig_resolverEndpointBase(rName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_route53_resolver_endpoint" "test" {
  count = 2

  direction = "OUTBOUND"
  name      = "%[1]s-${count.index}"

  security_group_ids = [aws_security_group.test[0].id]

  ip_address {
    subnet_id = aws_subnet.test[2].id
  }

  ip_address {
    subnet_id = aws_subnet.test[count.index].id
  }
}
`, rName))
}

func testAccRuleConfig_resolverEndpointRecreateBase(rName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_route53_resolver_endpoint" "test" {
  count = 2

  direction = "OUTBOUND"
  name      = "%[1]s-${count.index}"

  security_group_ids = [aws_security_group.test[1].id]

  ip_address {
    subnet_id = aws_subnet.test[2].id
  }

  ip_address {
    subnet_id = aws_subnet.test[count.index].id
  }
}
`, rName))
}

func testAccRuleConfig_resolverEndpointMultiProtocolBase(rName string) string {
	return acctest.ConfigCompose(testAccRuleConfig_vpcBase(rName), fmt.Sprintf(`
resource "aws_route53_resolver_endpoint" "test" {
  count = 2

  direction = "OUTBOUND"
  name      = "%[1]s-${count.index}"

  security_group_ids = [aws_security_group.test[0].id]

  ip_address {
    subnet_id = aws_subnet.test[2].id
  }

  ip_address {
    subnet_id = aws_subnet.test[count.index].id
  }

  protocols = ["Do53", "DoH"]
}
`, rName))
}
