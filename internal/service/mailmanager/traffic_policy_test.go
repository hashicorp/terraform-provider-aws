// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmailmanager "github.com/hashicorp/terraform-provider-aws/internal/service/mailmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMailManagerTrafficPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var trafficPolicy mailmanager.GetTrafficPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_traffic_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficPolicyExists(ctx, t, resourceName, &trafficPolicy),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", regexache.MustCompile(`.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDefaultAction, "ALLOW"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_timestamp"),
					resource.TestCheckNoResourceAttr(resourceName, "max_message_size_bytes"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.action", "DENY"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.0.ip_expression.0.evaluate.0.attribute", "SENDER_IP"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.0.ip_expression.0.operator", "CIDR_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.0.ip_expression.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.0.ip_expression.0.values.0", "192.0.2.0/24"),
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

func TestAccMailManagerTrafficPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)

	var before, after mailmanager.GetTrafficPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_traffic_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_update(rName, "ALLOW", 100000, "CIDR_MATCHES", "192.0.2.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficPolicyExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrDefaultAction, "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "max_message_size_bytes", "100000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccTrafficPolicyConfig_update(rName+"-updated", "DENY", 200000, "NOT_CIDR_MATCHES", "198.51.100.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficPolicyExists(ctx, t, resourceName, &after),
					testAccCheckTrafficPolicyNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrDefaultAction, "DENY"),
					resource.TestCheckResourceAttr(resourceName, "max_message_size_bytes", "200000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.0.ip_expression.0.operator", "NOT_CIDR_MATCHES"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.0.ip_expression.0.values.0", "198.51.100.0/24"),
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

func TestAccMailManagerTrafficPolicy_conditionTypes(t *testing.T) {
	ctx := acctest.Context(t)

	var trafficPolicy mailmanager.GetTrafficPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_traffic_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_conditionTypes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficPolicyExists(ctx, t, resourceName, &trafficPolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.0.ipv6_expression.0.evaluate.0.attribute", "SENDER_IPV6"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.1.string_expression.0.evaluate.0.attribute", "RECIPIENT"),
					resource.TestCheckResourceAttr(resourceName, "policy_statement.0.condition.2.tls_expression.0.evaluate.0.attribute", "TLS_PROTOCOL"),
				),
			},
		},
	})
}

func TestAccMailManagerTrafficPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var trafficPolicy mailmanager.GetTrafficPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_traffic_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficPolicyExists(ctx, t, resourceName, &trafficPolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmailmanager.ResourceTrafficPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckTrafficPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mailmanager_traffic_policy" {
				continue
			}
			_, err := tfmailmanager.FindTrafficPolicyByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.MailManager, create.ErrActionCheckingDestroyed, tfmailmanager.ResNameTrafficPolicy, rs.Primary.ID, err)
			}
			return create.Error(names.MailManager, create.ErrActionCheckingDestroyed, tfmailmanager.ResNameTrafficPolicy, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccCheckTrafficPolicyExists(ctx context.Context, t *testing.T, name string, trafficPolicy *mailmanager.GetTrafficPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameTrafficPolicy, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameTrafficPolicy, name, errors.New("not set"))
		}
		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)
		out, err := tfmailmanager.FindTrafficPolicyByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameTrafficPolicy, rs.Primary.ID, err)
		}
		*trafficPolicy = *out
		return nil
	}
}

func testAccCheckTrafficPolicyNotRecreated(before, after *mailmanager.GetTrafficPolicyOutput) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if before, after := *before.TrafficPolicyId, *after.TrafficPolicyId; before != after {
			return create.Error(names.MailManager, create.ErrActionCheckingNotRecreated, tfmailmanager.ResNameTrafficPolicy, before, errors.New("recreated"))
		}
		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)
	var input mailmanager.ListTrafficPoliciesInput
	_, err := conn.ListTrafficPolicies(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTrafficPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_traffic_policy" "test" {
  default_action = "ALLOW"
  name           = %[1]q

  policy_statement {
    action = "DENY"

    condition {
      ip_expression {
        operator = "CIDR_MATCHES"
        values   = ["192.0.2.0/24"]

        evaluate {
          attribute = "SENDER_IP"
        }
      }
    }
  }
}
`, rName)
}

func testAccTrafficPolicyConfig_update(rName, defaultAction string, maxMessageSize int, operator, cidr string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_traffic_policy" "test" {
  default_action         = %[2]q
  max_message_size_bytes = %[3]d
  name                   = %[1]q

  policy_statement {
    action = "DENY"

    condition {
      ip_expression {
        operator = %[4]q
        values   = [%[5]q]

        evaluate {
          attribute = "SENDER_IP"
        }
      }
    }
  }
}
`, rName, defaultAction, maxMessageSize, operator, cidr)
}

func testAccTrafficPolicyConfig_conditionTypes(rName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_traffic_policy" "test" {
  default_action = "ALLOW"
  name           = %[1]q

  policy_statement {
    action = "DENY"

    condition {
      ipv6_expression {
        operator = "CIDR_MATCHES"
        values   = ["2001:db8::/32"]

        evaluate {
          attribute = "SENDER_IPV6"
        }
      }
    }

    condition {
      string_expression {
        operator = "ENDS_WITH"
        values   = ["@example.com"]

        evaluate {
          attribute = "RECIPIENT"
        }
      }
    }

    condition {
      tls_expression {
        operator = "MINIMUM_TLS_VERSION"
        value    = "TLS1_2"

        evaluate {
          attribute = "TLS_PROTOCOL"
        }
      }
    }
  }
}
`, rName)
}
