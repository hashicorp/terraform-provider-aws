// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudWatchOTelEnrichment_serial(t *testing.T) {
	t.Parallel()
	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccOTelEnrichment_basic,
		acctest.CtDisappears: testAccOTelEnrichment_disappears,
		"filters":            testAccOTelEnrichment_filters,
		"filtersUpdate":      testAccOTelEnrichment_filtersUpdate,
		"Identity":           testAccCloudWatchOTelEnrichment_identitySerial,
	}
	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccOTelEnrichment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_otel_enrichment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheckOTelEnrichment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOTelEnrichmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOtelEnrichmentConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOTelEnrichmentExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func testAccOTelEnrichment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_otel_enrichment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheckOTelEnrichment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOTelEnrichmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOtelEnrichmentConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOTelEnrichmentExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatch.ResourceOtelEnrichment, resourceName),
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

func testAccOTelEnrichment_filters(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_otel_enrichment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheckOTelEnrichment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOTelEnrichmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOtelEnrichmentConfig_filters(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOTelEnrichmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "include_filters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filters.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "include_filters.*", map[string]string{
						names.AttrNamespace: "AWS/EC2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "include_filters.*", map[string]string{
						names.AttrNamespace: "AWS/RDS",
						"metric_names.#":    "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "exclude_filters.*", map[string]string{
						names.AttrNamespace: "AWS/EC2",
						"metric_names.#":    "1",
					}),
				),
			},
		},
	})
}

func testAccOTelEnrichment_filtersUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_otel_enrichment.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheckOTelEnrichment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOTelEnrichmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Start unfiltered, so every supported namespace is enriched.
			{
				Config: testAccOtelEnrichmentConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOTelEnrichmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "include_filters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filters.#", "0"),
				),
			},
			// Add filters in place; this must not restart enrichment.
			{
				Config: testAccOtelEnrichmentConfig_filters(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOTelEnrichmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "include_filters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filters.#", "1"),
				),
			},
			// Narrow to a single include selector and drop the excludes.
			{
				Config: testAccOtelEnrichmentConfig_filtersIncludeOnly(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOTelEnrichmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "include_filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filters.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "include_filters.*", map[string]string{
						names.AttrNamespace: "AWS/Lambda",
					}),
				),
			},
			// Remove every filter, reverting to all supported namespaces.
			{
				Config: testAccOtelEnrichmentConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOTelEnrichmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "include_filters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "exclude_filters.#", "0"),
				),
			},
		},
	})
}

func testAccCheckOTelEnrichmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_otel_enrichment" {
				continue
			}

			_, err := tfcloudwatch.FindOtelEnrichment(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return errors.New("CloudWatch OTel Enrichment still enabled")
		}

		return nil
	}
}

func testAccCheckOTelEnrichmentExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		_, err := tfcloudwatch.FindOtelEnrichment(ctx, conn)

		return err
	}
}

func testAccPreCheckOTelEnrichment(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

	input := cloudwatch.GetOTelEnrichmentInput{}
	_, err := conn.GetOTelEnrichment(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	// Skip SignatureDoesNotMatch errors (region/endpoint mismatch)
	if err != nil && strings.Contains(err.Error(), "SignatureDoesNotMatch") {
		t.Skipf("skipping acceptance testing: region/endpoint mismatch: %s", err)
	}
	// Skip InvalidClientTokenId errors (credential issues with private API)
	if err != nil && strings.Contains(err.Error(), "InvalidClientTokenId") {
		t.Skipf("skipping acceptance testing: credential issue with private API: %s", err)
	}
	// Skip AccessDenied errors (API not yet available in granite environment)
	if err != nil && strings.Contains(err.Error(), "AccessDenied") {
		t.Skipf("skipping acceptance testing: API not yet available: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOtelEnrichmentConfig_basic() string {
	return `
resource "aws_observabilityadmin_telemetry_enrichment" "test" {
}

resource "aws_cloudwatch_otel_enrichment" "test" {
  depends_on = [aws_observabilityadmin_telemetry_enrichment.test]
}
`
}

func testAccOtelEnrichmentConfig_filters() string {
	return `
resource "aws_observabilityadmin_telemetry_enrichment" "test" {
}

resource "aws_cloudwatch_otel_enrichment" "test" {
  include_filters {
    namespace = "AWS/EC2"
  }

  include_filters {
    namespace    = "AWS/RDS"
    metric_names = ["CPUUtilization", "DatabaseConnections"]
  }

  exclude_filters {
    namespace    = "AWS/EC2"
    metric_names = ["NetworkPacketsIn"]
  }

  depends_on = [aws_observabilityadmin_telemetry_enrichment.test]
}
`
}

func testAccOtelEnrichmentConfig_filtersIncludeOnly() string {
	return `
resource "aws_observabilityadmin_telemetry_enrichment" "test" {
}

resource "aws_cloudwatch_otel_enrichment" "test" {
  include_filters {
    namespace = "AWS/Lambda"
  }

  depends_on = [aws_observabilityadmin_telemetry_enrichment.test]
}
`
}
