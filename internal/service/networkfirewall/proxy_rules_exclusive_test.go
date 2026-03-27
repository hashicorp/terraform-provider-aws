// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccNetworkFirewallProxyRulesExclusive_basic(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v networkfirewall.DescribeProxyRuleGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_rules_exclusive.test"
	ruleGroupResourceName := "aws_networkfirewall_proxy_rule_group.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyRulesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyRulesExclusiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyRulesExclusiveExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "proxy_rule_group_arn", ruleGroupResourceName, names.AttrARN),
					// Pre-DNS phase
					resource.TestCheckResourceAttr(resourceName, "pre_dns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.proxy_rule_name", fmt.Sprintf("%s-predns", rName)),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.action", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_key", "request:DestinationDomain"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_operator", "StringEquals"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_values.0", "amazonaws.com"),
					// Pre-REQUEST phase
					resource.TestCheckResourceAttr(resourceName, "pre_request.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.proxy_rule_name", fmt.Sprintf("%s-prerequest", rName)),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.action", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.conditions.0.condition_key", "request:Http:Method"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.conditions.0.condition_operator", "StringEquals"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.conditions.0.condition_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.conditions.0.condition_values.0", "DELETE"),
					// Post-RESPONSE phase
					resource.TestCheckResourceAttr(resourceName, "post_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.proxy_rule_name", fmt.Sprintf("%s-postresponse", rName)),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.action", "ALERT"),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.conditions.0.condition_key", "response:Http:StatusCode"),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.conditions.0.condition_operator", "NumericGreaterThanEquals"),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.conditions.0.condition_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.conditions.0.condition_values.0", "500"),
				),
			},
			{
				Config:       testAccProxyRulesExclusiveConfig_single(rName),
				ResourceName: resourceName,
				ImportState:  true,
			},
		},
	})
}

func testAccNetworkFirewallProxyRulesExclusive_disappears(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v networkfirewall.DescribeProxyRuleGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_rules_exclusive.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyRulesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyRulesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyRulesExclusiveExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkfirewall.ResourceProxyRulesExclusive, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccNetworkFirewallProxyRulesExclusive_updateAdd(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v1, v2 networkfirewall.DescribeProxyRuleGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_rules_exclusive.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyRulesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyRulesExclusiveConfig_single(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyRulesExclusiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "post_response.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyRulesExclusiveConfig_add(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyRulesExclusiveExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "post_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.proxy_rule_name", fmt.Sprintf("%s-prerequest-new", rName)),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.proxy_rule_name", fmt.Sprintf("%s-postresponse-new", rName)),
				),
			},
		},
	})
}

func testAccNetworkFirewallProxyRulesExclusive_updateModify(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v1, v2 networkfirewall.DescribeProxyRuleGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_rules_exclusive.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyRulesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyRulesExclusiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyRulesExclusiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.action", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_values.0", "amazonaws.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyRulesExclusiveConfig_modified(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyRulesExclusiveExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.action", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_values.0", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_values.1", "test.com"),
				),
			},
		},
	})
}

func testAccNetworkFirewallProxyRulesExclusive_updateRemove(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v1, v2 networkfirewall.DescribeProxyRuleGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_rules_exclusive.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyRulesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyRulesExclusiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyRulesExclusiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "post_response.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyRulesExclusiveConfig_single(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyRulesExclusiveExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "post_response.#", "0"),
				),
			},
		},
	})
}

func testAccNetworkFirewallProxyRulesExclusive_multipleRulesPerPhase(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v networkfirewall.DescribeProxyRuleGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_rules_exclusive.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyRulesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyRulesExclusiveConfig_multiplePerPhase(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyRulesExclusiveExists(ctx, t, resourceName, &v),
					// Pre-DNS phase - 2 rules
					resource.TestCheckResourceAttr(resourceName, "pre_dns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.proxy_rule_name", fmt.Sprintf("%s-predns-1", rName)),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.action", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.0.conditions.0.condition_values.0", "amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.1.proxy_rule_name", fmt.Sprintf("%s-predns-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.1.action", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "pre_dns.1.conditions.0.condition_values.0", "malicious.com"),
					// Pre-REQUEST phase - 2 rules
					resource.TestCheckResourceAttr(resourceName, "pre_request.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.proxy_rule_name", fmt.Sprintf("%s-prerequest-1", rName)),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.action", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.0.conditions.0.condition_values.0", "DELETE"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.1.proxy_rule_name", fmt.Sprintf("%s-prerequest-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "pre_request.1.action", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "pre_request.1.conditions.0.condition_values.0", "PUT"),
					// Post-RESPONSE phase - 2 rules
					resource.TestCheckResourceAttr(resourceName, "post_response.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.proxy_rule_name", fmt.Sprintf("%s-postresponse-1", rName)),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.action", "ALERT"),
					resource.TestCheckResourceAttr(resourceName, "post_response.0.conditions.0.condition_values.0", "500"),
					resource.TestCheckResourceAttr(resourceName, "post_response.1.proxy_rule_name", fmt.Sprintf("%s-postresponse-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "post_response.1.action", "ALERT"),
					resource.TestCheckResourceAttr(resourceName, "post_response.1.conditions.0.condition_values.0", "404"),
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

func testAccCheckProxyRulesExclusiveDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_proxy_rules_exclusive" {
				continue
			}

			// The resource ID is the proxy rule group ARN
			out, err := tfnetworkfirewall.FindProxyRuleGroupByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if there are any rules in the group
			if out != nil && out.ProxyRuleGroup != nil && out.ProxyRuleGroup.Rules != nil {
				rules := out.ProxyRuleGroup.Rules
				if len(rules.PreDNS) > 0 || len(rules.PreREQUEST) > 0 || len(rules.PostRESPONSE) > 0 {
					return fmt.Errorf("NetworkFirewall Proxy Rules still exist in group %s", rs.Primary.ID)
				}
			}
		}

		return nil
	}
}

func testAccCheckProxyRulesExclusiveExists(ctx context.Context, t *testing.T, n string, v ...*networkfirewall.DescribeProxyRuleGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		output, err := tfnetworkfirewall.FindProxyRuleGroupByARN(ctx, conn, rs.Primary.Attributes["proxy_rule_group_arn"])

		if err != nil {
			return err
		}

		if len(v) > 0 {
			*v[0] = *output
		}

		return nil
	}
}

func testAccProxyRulesExclusiveConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_proxy_rule_group" "test" {
  name = %[1]q
}

resource "aws_networkfirewall_proxy_rules_exclusive" "test" {
  proxy_rule_group_arn = aws_networkfirewall_proxy_rule_group.test.arn

  pre_dns {
    proxy_rule_name = "%[1]s-predns"
    action          = "ALLOW"
    description     = "%[1]s-predns-description"

    conditions {
      condition_key      = "request:DestinationDomain"
      condition_operator = "StringEquals"
      condition_values   = ["amazonaws.com"]
    }
  }

  pre_request {
    proxy_rule_name = "%[1]s-prerequest"
    action          = "DENY"
    description     = "%[1]s-prerequest-description"

    conditions {
      condition_key      = "request:Http:Method"
      condition_operator = "StringEquals"
      condition_values   = ["DELETE"]
    }
  }

  post_response {
    proxy_rule_name = "%[1]s-postresponse"
    action          = "ALERT"
    description     = "%[1]s-postresponse-description"

    conditions {
      condition_key      = "response:Http:StatusCode"
      condition_operator = "NumericGreaterThanEquals"
      condition_values   = ["500"]
    }
  }
}
`, rName)
}

func testAccProxyRulesExclusiveConfig_single(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_proxy_rule_group" "test" {
  name = %[1]q
}

resource "aws_networkfirewall_proxy_rules_exclusive" "test" {
  proxy_rule_group_arn = aws_networkfirewall_proxy_rule_group.test.arn

  pre_dns {
    proxy_rule_name = "%[1]s-predns"
    action          = "ALLOW"

    conditions {
      condition_key      = "request:DestinationDomain"
      condition_operator = "StringEquals"
      condition_values   = ["amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccProxyRulesExclusiveConfig_modified(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_proxy_rule_group" "test" {
  name = %[1]q
}

resource "aws_networkfirewall_proxy_rules_exclusive" "test" {
  proxy_rule_group_arn = aws_networkfirewall_proxy_rule_group.test.arn

  pre_dns {
    proxy_rule_name = "%[1]s-predns"
    action          = "ALLOW"
    description     = "%[1]s-predns-description"

    conditions {
      condition_key      = "request:DestinationDomain"
      condition_operator = "StringEquals"
      condition_values   = ["example.com", "test.com"]
    }
  }

  pre_request {
    proxy_rule_name = "%[1]s-prerequest"
    action          = "DENY"

    conditions {
      condition_key      = "request:Http:Method"
      condition_operator = "StringEquals"
      condition_values   = ["DELETE"]
    }
  }

  post_response {
    proxy_rule_name = "%[1]s-postresponse"
    action          = "ALERT"
    description     = "%[1]s-postresponse-description"

    conditions {
      condition_key      = "response:Http:StatusCode"
      condition_operator = "NumericGreaterThanEquals"
      condition_values   = ["500"]
    }
  }
}
`, rName)
}

func testAccProxyRulesExclusiveConfig_add(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_proxy_rule_group" "test" {
  name = %[1]q
}

resource "aws_networkfirewall_proxy_rules_exclusive" "test" {
  proxy_rule_group_arn = aws_networkfirewall_proxy_rule_group.test.arn

  pre_dns {
    proxy_rule_name = "%[1]s-predns"
    action          = "ALLOW"

    conditions {
      condition_key      = "request:DestinationDomain"
      condition_operator = "StringEquals"
      condition_values   = ["amazonaws.com"]
    }
  }

  pre_request {
    proxy_rule_name = "%[1]s-prerequest-new"
    action          = "DENY"

    conditions {
      condition_key      = "request:Http:Method"
      condition_operator = "StringEquals"
      condition_values   = ["POST"]
    }
  }

  post_response {
    proxy_rule_name = "%[1]s-postresponse-new"
    action          = "ALERT"

    conditions {
      condition_key      = "response:Http:StatusCode"
      condition_operator = "NumericGreaterThanEquals"
      condition_values   = ["400"]
    }
  }
}
`, rName)
}

func testAccProxyRulesExclusiveConfig_multiplePerPhase(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_proxy_rule_group" "test" {
  name = %[1]q
}

resource "aws_networkfirewall_proxy_rules_exclusive" "test" {
  proxy_rule_group_arn = aws_networkfirewall_proxy_rule_group.test.arn

  pre_dns {
    proxy_rule_name = "%[1]s-predns-1"
    action          = "ALLOW"

    conditions {
      condition_key      = "request:DestinationDomain"
      condition_operator = "StringEquals"
      condition_values   = ["amazonaws.com"]
    }
  }

  pre_dns {
    proxy_rule_name = "%[1]s-predns-2"
    action          = "DENY"

    conditions {
      condition_key      = "request:DestinationDomain"
      condition_operator = "StringEquals"
      condition_values   = ["malicious.com"]
    }
  }

  pre_request {
    proxy_rule_name = "%[1]s-prerequest-1"
    action          = "DENY"

    conditions {
      condition_key      = "request:Http:Method"
      condition_operator = "StringEquals"
      condition_values   = ["DELETE"]
    }
  }

  pre_request {
    proxy_rule_name = "%[1]s-prerequest-2"
    action          = "DENY"

    conditions {
      condition_key      = "request:Http:Method"
      condition_operator = "StringEquals"
      condition_values   = ["PUT"]
    }
  }

  post_response {
    proxy_rule_name = "%[1]s-postresponse-1"
    action          = "ALERT"

    conditions {
      condition_key      = "response:Http:StatusCode"
      condition_operator = "NumericGreaterThanEquals"
      condition_values   = ["500"]
    }
  }

  post_response {
    proxy_rule_name = "%[1]s-postresponse-2"
    action          = "ALERT"

    conditions {
      condition_key      = "response:Http:StatusCode"
      condition_operator = "NumericEquals"
      condition_values   = ["404"]
    }
  }
}
`, rName)
}
