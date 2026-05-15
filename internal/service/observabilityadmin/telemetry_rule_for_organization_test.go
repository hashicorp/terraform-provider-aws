// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfobservabilityadmin "github.com/hashicorp/terraform-provider-aws/internal/service/observabilityadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTelemetryRuleForOrganizationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

	// Ensure telemetry evaluation for organization is running (prerequisite for rules).
	var evalInput observabilityadmin.GetTelemetryEvaluationStatusForOrganizationInput
	evalOutput, err := conn.GetTelemetryEvaluationStatusForOrganization(ctx, &evalInput)

	if err != nil {
		if acctest.PreCheckSkipError(err) {
			t.Skipf("skipping acceptance testing: %s", err)
		}
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if evalOutput.Status != "RUNNING" {
		var startInput observabilityadmin.StartTelemetryEvaluationForOrganizationInput
		_, err := conn.StartTelemetryEvaluationForOrganization(ctx, &startInput)
		if err != nil {
			t.Fatalf("failed to start telemetry evaluation for organization: %s", err)
		}
	}
}

func TestAccObservabilityAdminTelemetryRuleForOrganization_serial(t *testing.T) {
	t.Parallel()
	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:        testAccObservabilityAdminTelemetryRuleForOrganization_basic,
		acctest.CtDisappears:   testAccObservabilityAdminTelemetryRuleForOrganization_disappears,
		"tags":                 testAccObservabilityAdminTelemetryRuleForOrganization_tags,
		"List_basic":           testAccTelemetryRuleForOrganization_List_basic,
		"List_includeResource": testAccTelemetryRuleForOrganization_List_includeResource,
	}
	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccObservabilityAdminTelemetryRuleForOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v observabilityadmin.GetTelemetryRuleForOrganizationOutput
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			// https://docs.aws.amazon.com/organizations/latest/userguide/services-that-can-integrate-cloudwatch.html
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/observabilityadmin.amazonaws.com")
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryRuleForOrganizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_arn"),
				),
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

func testAccObservabilityAdminTelemetryRuleForOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v observabilityadmin.GetTelemetryRuleForOrganizationOutput
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			// https://docs.aws.amazon.com/organizations/latest/userguide/services-that-can-integrate-cloudwatch.html
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/observabilityadmin.amazonaws.com")
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryRuleForOrganizationConfig_disappears(rName),
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

func testAccObservabilityAdminTelemetryRuleForOrganization_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v observabilityadmin.GetTelemetryRuleForOrganizationOutput
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			// https://docs.aws.amazon.com/organizations/latest/userguide/services-that-can-integrate-cloudwatch.html
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/observabilityadmin.amazonaws.com")
			acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
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
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_name",
			},
			{
				Config: testAccTelemetryRuleForOrganizationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTelemetryRuleForOrganizationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
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

			_, err := tfobservabilityadmin.FindTelemetryRuleForOrganization(ctx, conn, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("ObservabilityAdmin Telemetry Rule For Organization %s still exists", rs.Primary.ID)
			}
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

		output, err := tfobservabilityadmin.FindTelemetryRuleForOrganization(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTelemetryRuleForOrganizationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    resource_type  = "AWS::SecurityHub::Hub"
    telemetry_type = "Logs"
  }
}
`, rName)
}

func testAccTelemetryRuleForOrganizationConfig_disappears(rName string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    resource_type  = "AWS::EC2::Instance"
    telemetry_type = "Metrics"
  }
}
`, rName)
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
}
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
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
