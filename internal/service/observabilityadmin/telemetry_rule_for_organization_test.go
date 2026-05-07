// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfobservabilityadmin "github.com/hashicorp/terraform-provider-aws/internal/service/observabilityadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccObservabilityAdminTelemetryRuleForOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_rule_for_organization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_arn"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "rule_name"),
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

	acctest.Test(ctx, t, resource.TestCase{
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryRuleForOrganizationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfobservabilityadmin.ResourceTelemetryRuleForOrganization, resourceName),
				),
				ExpectNonEmptyPlan: true,
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

			_, err := tfobservabilityadmin.FindTelemetryRuleForOrganization(ctx, conn, rs.Primary.Attributes["rule_name"])

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Observability Admin Telemetry Rule For Organization %s still exists", rs.Primary.Attributes["rule_name"])
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

		_, err := tfobservabilityadmin.FindTelemetryRuleForOrganization(ctx, conn, rs.Primary.Attributes["rule_name"])

		return err
	}
}

func testAccTelemetryRuleForOrganizationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

	input := observabilityadmin.GetTelemetryEvaluationStatusForOrganizationInput{}
	_, err := conn.GetTelemetryEvaluationStatusForOrganization(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTelemetryRuleForOrganizationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_rule_for_organization" "test" {
  rule_name = %[1]q

  rule {
    telemetry_type = "Logs"
  }
}
`, rName)
}
