// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfobservabilityadmin "github.com/hashicorp/terraform-provider-aws/internal/service/observabilityadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTelemetryRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_rule.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_arn"), tfknownvalue.RegionalARNExact("observabilityadmin", "telemetry-rule/"+rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_name",
			},
		},
	})
}

func testAccTelemetryRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_rule.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfobservabilityadmin.ResourceTelemetryRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccTelemetryRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_rule.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_name",
			},
			{
				Config: testAccTelemetryRuleConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccTelemetryRuleConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckTelemetryRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_observabilityadmin_telemetry_rule" {
				continue
			}

			_, err := tfobservabilityadmin.FindTelemetryRuleByName(ctx, conn, rs.Primary.Attributes["rule_name"])

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ObservabilityAdmin Telemetry Rule %s still exists", rs.Primary.Attributes["rule_name"])
		}

		return nil
	}
}

func testAccCheckTelemetryRuleExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		_, err := tfobservabilityadmin.FindTelemetryRuleByName(ctx, conn, rs.Primary.Attributes["rule_name"])

		return err
	}
}

func testAccTelemetryRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule" "test" {
  rule_name = %[1]q

  rule {
    resource_type  = "AWS::EC2::VPC"
    telemetry_type = "Logs"
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation.test]
}

resource "aws_observabilityadmin_telemetry_evaluation" "test" {}
`, rName)
}

func testAccTelemetryRuleConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule" "test" {
  rule_name = %[1]q

  rule {
    resource_type  = "AWS::EC2::VPC"
    telemetry_type = "Logs"
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation.test]
}

resource "aws_observabilityadmin_telemetry_evaluation" "test" {}
`, rName, tagKey1, tagValue1)
}

func testAccTelemetryRuleConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule" "test" {
  rule_name = %[1]q

  rule {
    resource_type  = "AWS::EC2::VPC"
    telemetry_type = "Logs"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation.test]
}

resource "aws_observabilityadmin_telemetry_evaluation" "test" {}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
