// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/config"
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

func testAccTelemetryRuleForOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v observabilityadmin.GetTelemetryRuleForOrganizationOutput
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TelemetryRuleForOrganization/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_arn"), tfknownvalue.RegionalARNExact("observabilityadmin", "organization-telemetry-rule/"+rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TelemetryRuleForOrganization/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_name",
			},
		},
	})
}

func testAccTelemetryRuleForOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v observabilityadmin.GetTelemetryRuleForOrganizationOutput
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TelemetryRuleForOrganization/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfobservabilityadmin.ResourceTelemetryRuleForOrganization, resourceName),
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

func testAccTelemetryRuleForOrganization_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v observabilityadmin.GetTelemetryRuleForOrganizationOutput
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryRuleForOrganizationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName, &v),
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
				Config: testAccTelemetryRuleForOrganizationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName, &v),
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
				Config: testAccTelemetryRuleForOrganizationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName, &v),
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

func testAccCheckTelemetryRuleForOrganizationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_observabilityadmin_telemetry_rule_for_organization" {
				continue
			}

			_, err := tfobservabilityadmin.FindTelemetryRuleForOrganizationByName(ctx, conn, rs.Primary.Attributes["rule_name"])

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ObservabilityAdmin Telemetry Rule For Organization %s still exists", rs.Primary.Attributes["rule_name"])
		}

		return nil
	}
}

func testAccCheckTelemetryRuleForOrganizationExists(ctx context.Context, t *testing.T, n string, v *observabilityadmin.GetTelemetryRuleForOrganizationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		output, err := tfobservabilityadmin.FindTelemetryRuleForOrganizationByName(ctx, conn, rs.Primary.Attributes["rule_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTelemetryRuleForOrganizationPreCheck(ctx context.Context, t *testing.T) {
	// https://docs.aws.amazon.com/organizations/latest/userguide/services-that-can-integrate-cloudwatch.html
	acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/observabilityadmin.amazonaws.com")
	acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
	acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
}

func testAccTelemetryRuleForOrganizationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    resource_type  = "AWS::SecurityHub::HubV2"
    telemetry_type = "Logs"
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation_for_organization.test]
}

resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "test" {}
`, rName, tagKey1, tagValue1)
}

func testAccTelemetryRuleForOrganizationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    resource_type  = "AWS::SecurityHub::HubV2"
    telemetry_type = "Logs"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_observabilityadmin_telemetry_evaluation_for_organization.test]
}

resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "test" {}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
