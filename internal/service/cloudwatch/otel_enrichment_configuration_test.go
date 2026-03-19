// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
)

func TestAccCloudWatchOtelEnrichmentConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_otel_enrichment_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOtelEnrichmentConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOtelEnrichmentConfigurationConfig_basic(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOtelEnrichmentConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
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

func TestAccCloudWatchOtelEnrichmentConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_otel_enrichment_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOtelEnrichmentConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOtelEnrichmentConfigurationConfig_basic(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOtelEnrichmentConfigurationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatch.ResourceOtelEnrichmentConfiguration, resourceName),
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

func testAccCheckOtelEnrichmentConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_otel_enrichment_configuration" {
				continue
			}

			_, err := tfcloudwatch.FindOtelEnrichmentConfiguration(ctx, conn)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.CloudWatch, create.ErrActionCheckingDestroyed, tfcloudwatch.ResNameOtelEnrichmentConfiguration, rs.Primary.ID, err)
			}

			return create.Error(names.CloudWatch, create.ErrActionCheckingDestroyed, tfcloudwatch.ResNameOtelEnrichmentConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOtelEnrichmentConfigurationExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameOtelEnrichmentConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameOtelEnrichmentConfiguration, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

		_, err := tfcloudwatch.FindOtelEnrichmentConfiguration(ctx, conn)
		if err != nil {
			return create.Error(names.CloudWatch, create.ErrActionCheckingExistence, tfcloudwatch.ResNameOtelEnrichmentConfiguration, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).CloudWatchClient(ctx)

	input := cloudwatch.GetOTelEnrichmentConfigurationInput{}
_, err := conn.GetOTelEnrichmentConfiguration(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOtelEnrichmentConfigurationConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_otel_enrichment_configuration" "test" {
  enabled = %[1]t
}
`, enabled)
}
