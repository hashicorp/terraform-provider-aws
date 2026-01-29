// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccARCRegionSwitchRoute53HealthChecksDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_arcregionswitch_route53_health_checks.test"
	resourceName := "aws_arcregionswitch_plan.test"

	zoneName := acctest.RandomDomain()
	recordName := zoneName.RandomSubdomain()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCRegionSwitch),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthChecksDataSourceConfig_basic(rName, zoneName.String(), recordName.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "plan_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "health_checks.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_checks.0.hosted_zone_id", "aws_route53_zone.test", "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_checks.0.record_name", "aws_route53_record.test", names.AttrName),
					resource.TestCheckResourceAttrSet(dataSourceName, "health_checks.0.region"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_checks.1.hosted_zone_id", "aws_route53_zone.test", "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "health_checks.1.record_name", "aws_route53_record.test", names.AttrName),
					resource.TestCheckResourceAttrSet(dataSourceName, "health_checks.1.region"),
				),
			},
		},
	})
}

func testAccRoute53HealthChecksDataSourceConfig_basic(rName, zoneName, recordName string) string {
	return fmt.Sprintf(`
data "aws_arcregionswitch_route53_health_checks" "test" {
  plan_arn = aws_arcregionswitch_plan.test.arn
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
      route53_health_check_config {
        hosted_zone_id  = aws_route53_zone.test.zone_id
        record_name     = aws_route53_record.test.name
        timeout_minutes = 10

        record_set {
          record_set_identifier = "primary"
          region                = %[2]q
        }
        record_set {
          record_set_identifier = "secondary"
          region                = %[3]q
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
      route53_health_check_config {
        hosted_zone_id  = aws_route53_zone.test.zone_id
        record_name     = aws_route53_record.test.name
        timeout_minutes = 10

        record_set {
          record_set_identifier = "primary"
          region                = %[2]q
        }
        record_set {
          record_set_identifier = "secondary"
          region                = %[3]q
        }
      }
    }
  }
}

resource "aws_route53_record" "test" {
  zone_id = aws_route53_zone.test.zone_id
  name    = %[5]q
  type    = "A"
  ttl     = "30"
  records = ["127.0.0.1", "127.0.0.27"]
}

resource "aws_route53_zone" "test" {
  name = %[4]q
}

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
`, rName, acctest.AlternateRegion(), acctest.Region(), zoneName, recordName)
}
