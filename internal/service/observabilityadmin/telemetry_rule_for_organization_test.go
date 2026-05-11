// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
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

func TestAccObservabilityAdminTelemetryRuleForOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryRuleForOrganizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_name",
			},
		},
	})
}

func TestAccObservabilityAdminTelemetryRuleForOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryRuleForOrganizationConfig_disappears(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfobservabilityadmin.ResourceTelemetryRuleForOrganization, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccObservabilityAdminTelemetryRuleForOrganization_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			testAccTelemetryRuleForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryRuleForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryRuleForOrganizationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName),
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
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTelemetryRuleForOrganizationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName),
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

func testAccCheckTelemetryRuleForOrganizationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		_, err := tfobservabilityadmin.FindTelemetryRuleForOrganization(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccTelemetryRuleForOrganizationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    resource_type  = "AWS::EC2::VPC"
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
    resource_type  = "AWS::EC2::VPC"
    telemetry_type = "Logs"
  }
}
`, rName)
}

func testAccTelemetryRuleForOrganizationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    resource_type  = "AWS::EC2::VPC"
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
    resource_type  = "AWS::EC2::VPC"
    telemetry_type = "Logs"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}