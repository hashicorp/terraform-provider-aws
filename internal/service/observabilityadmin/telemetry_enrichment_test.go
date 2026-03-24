// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfobservabilityadmin "github.com/hashicorp/terraform-provider-aws/internal/service/observabilityadmin"
)

func TestAccObservabilityAdminTelemetryEnrichment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_enrichment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccTelemetryEnrichmentPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryEnrichmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryEnrichmentConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryEnrichmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "aws_resource_explorer_managed_view_arn"),
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

func TestAccObservabilityAdminTelemetryEnrichment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_observabilityadmin_telemetry_enrichment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccTelemetryEnrichmentPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryEnrichmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryEnrichmentConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryEnrichmentExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfobservabilityadmin.ResourceTelemetryEnrichment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTelemetryEnrichmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_observabilityadmin_telemetry_enrichment" {
				continue
			}

			_, err := tfobservabilityadmin.FindTelemetryEnrichmentStatus(ctx, conn)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ObservabilityAdmin, create.ErrActionCheckingDestroyed, tfobservabilityadmin.ResNameTelemetryEnrichment, rs.Primary.ID, err)
			}

			return create.Error(names.ObservabilityAdmin, create.ErrActionCheckingDestroyed, tfobservabilityadmin.ResNameTelemetryEnrichment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTelemetryEnrichmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.ObservabilityAdmin, create.ErrActionCheckingExistence, tfobservabilityadmin.ResNameTelemetryEnrichment, n, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		_, err := tfobservabilityadmin.FindTelemetryEnrichmentStatus(ctx, conn)
		if err != nil {
			return create.Error(names.ObservabilityAdmin, create.ErrActionCheckingExistence, tfobservabilityadmin.ResNameTelemetryEnrichment, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccTelemetryEnrichmentPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

	_, err := conn.GetTelemetryEnrichmentStatus(ctx, &observabilityadmin.GetTelemetryEnrichmentStatusInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	// ResourceNotFoundException is expected when enrichment is not yet enabled — that's fine.
	if err != nil && !retry.NotFound(err) {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTelemetryEnrichmentConfig_basic() string {
	return `
resource "aws_observabilityadmin_telemetry_enrichment" "test" {}
`
}
