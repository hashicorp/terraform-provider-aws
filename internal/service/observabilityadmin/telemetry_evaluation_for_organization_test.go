// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"errors"
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

func testAccTelemetryEvaluationForOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_evaluation_for_organization.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccTelemetryEvaluationForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryEvaluationForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryEvaluationForOrganizationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryEvaluationForOrganizationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccTelemetryEvaluationForOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_evaluation_for_organization.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccTelemetryEvaluationForOrganizationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryEvaluationForOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryEvaluationForOrganizationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryEvaluationForOrganizationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfobservabilityadmin.ResourceTelemetryEvaluationForOrganization, resourceName),
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

func testAccCheckTelemetryEvaluationForOrganizationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_observabilityadmin_telemetry_evaluation_for_organization" {
				continue
			}

			_, err := tfobservabilityadmin.FindTelemetryEvaluationForOrganization(ctx, conn)

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return errors.New("Observability Admin Telemetry Evaluation For Organization still exists")
		}

		return nil
	}
}

func testAccCheckTelemetryEvaluationForOrganizationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		_, err := tfobservabilityadmin.FindTelemetryEvaluationForOrganization(ctx, conn)

		return err
	}
}

func testAccTelemetryEvaluationForOrganizationPreCheck(ctx context.Context, t *testing.T) {
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

func testAccTelemetryEvaluationForOrganizationConfig_basic() string {
	return `
resource "aws_observabilityadmin_telemetry_evaluation_for_organization" "test" {}
`
}
