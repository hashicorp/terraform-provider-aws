// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// lintignore:AWSAT003,AWSAT005
package arcregionswitch_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccARCRegionSwitchPlanDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_arcregionswitch_plan.test"
	resourceName := "aws_arcregionswitch_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "arcregionswitch"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtName, resourceName, acctest.CtName),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_role", resourceName, "execution_role"),
					resource.TestCheckResourceAttrPair(dataSourceName, "recovery_approach", resourceName, "recovery_approach"),
					resource.TestCheckResourceAttrPair(dataSourceName, "regions", resourceName, "regions"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workflow", resourceName, "workflow"),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlanDataSource_route53HealthChecks(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_arcregionswitch_plan.test"
	resourceName := "aws_arcregionswitch_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "arcregionswitch"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanDataSourceConfig_route53HealthChecks(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtName, resourceName, acctest.CtName),
					// Verify Route53 health checks API integration works
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.#", "2"),
					// Verify health check metadata is immediately available
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.0.hosted_zone_id", "Z123456789012345678"),
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.0.record_name", "test.example.com"),
					resource.TestCheckResourceAttrSet(dataSourceName, "route53_health_checks.0.region"),
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.1.hosted_zone_id", "Z123456789012345678"),
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.1.record_name", "test.example.com"),
					resource.TestCheckResourceAttrSet(dataSourceName, "route53_health_checks.1.region"),
					// Note: health_check_id fields exist but are empty initially due to AWS 4+ minute delay
				),
			},
		},
	})
}

func testAccPlanDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_basic(rName),
		`
data "aws_arcregionswitch_plan" "test" {
  arn = aws_arcregionswitch_plan.test.arn
}
`)
}

func TestAccARCRegionSwitchPlanDataSource_route53HealthChecksWithWait(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test requiring 4+ minute wait for health check IDs")
	}

	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_arcregionswitch_plan.test"
	resourceName := "aws_arcregionswitch_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "arcregionswitch"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanDataSourceConfig_route53HealthChecksWithWait(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtName, resourceName, acctest.CtName),
					// Verify health check IDs are populated after waiting
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.#", "2"),
					resource.TestCheckResourceAttrSet(dataSourceName, "route53_health_checks.0.health_check_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "route53_health_checks.1.health_check_id"),
					// Verify metadata is still correct
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.0.hosted_zone_id", "Z123456789012345678"),
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.0.record_name", "test.example.com"),
					resource.TestCheckResourceAttrSet(dataSourceName, "route53_health_checks.0.region"),
				),
			},
		},
	})
}

func testAccPlanDataSourceConfig_route53HealthChecksWithWait(rName string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_route53HealthChecks(rName),
		`
data "aws_arcregionswitch_plan" "test" {
  arn = aws_arcregionswitch_plan.test.arn
}
`)
}

func TestAccARCRegionSwitchPlanDataSource_withoutWaitFlags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_arcregionswitch_plan.test"
	resourceName := "aws_arcregionswitch_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "arcregionswitch"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanDataSourceConfig_withoutWaitFlags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtName, resourceName, acctest.CtName),
					// Verify health checks exist but IDs may be empty without wait
					resource.TestCheckResourceAttr(dataSourceName, "route53_health_checks.#", "2"),
				),
			},
		},
	})
}

func testAccPlanDataSourceConfig_withoutWaitFlags(rName string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_route53HealthChecks(rName),
		`
data "aws_arcregionswitch_plan" "test" {
  arn = aws_arcregionswitch_plan.test.arn
}
`)
}

func testAccPlanDataSourceConfig_route53HealthChecks(rName string) string {
	return acctest.ConfigCompose(
		testAccPlanConfig_route53HealthChecks(rName),
		`
data "aws_arcregionswitch_plan" "test" {
  arn = aws_arcregionswitch_plan.test.arn
}
`)
}

func testAccPlanConfig_route53HealthChecks(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "arc-region-switch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_arcregionswitch_plan" "test" {
  name              = %[1]q
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [%[2]q, %[3]q]
  primary_region    = %[2]q

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "route53-health-check-step"
      execution_block_type = "Route53HealthCheck"

      execution_block_configuration {
        route53_health_check_config {
          hosted_zone_id  = "Z123456789012345678"
          record_name     = "test.example.com"
          timeout_minutes = 10

          record_sets {
            record_set_identifier = "primary"
            region                = %[2]q
          }
          record_sets {
            record_set_identifier = "secondary"
            region                = %[3]q
          }
        }
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "route53-health-check-step-primary"
      execution_block_type = "Route53HealthCheck"

      execution_block_configuration {
        route53_health_check_config {
          hosted_zone_id  = "Z123456789012345678"
          record_name     = "test.example.com"
          timeout_minutes = 10

          record_sets {
            record_set_identifier = "primary"
            region                = %[2]q
          }
          record_sets {
            record_set_identifier = "secondary"
            region                = %[3]q
          }
        }
      }
    }
  }
}
`, rName, acctest.Region(), acctest.AlternateRegion())
}
