// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// lintignore:AWSAT003,AWSAT005
package arcregionswitch_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccARCRegionSwitchPlan_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCRegionSwitch),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPlanConfig_invalidRecoveryApproach(rName),
				ExpectError: regexache.MustCompile(`The provided value does not match any valid values`),
			},
			{
				Config:      testAccPlanConfig_invalidExecutionRole(rName),
				ExpectError: regexache.MustCompile(`value must be a valid ARN`),
			},
			{
				Config:      testAccPlanConfig_singleRegion(rName),
				ExpectError: regexache.MustCompile(`length greater than or equal to 2`),
			},
		},
	})
}

func testAccPlanConfig_invalidRecoveryApproach(rName string) string {
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
  recovery_approach = "invalidApproach"
  regions           = [%[2]q, %[3]q]
  primary_region    = %[2]q

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }
}
`, rName, acctest.Region(), acctest.AlternateRegion())
}

func testAccPlanConfig_invalidExecutionRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_arcregionswitch_plan" "test" {
  name              = %[1]q
  execution_role    = "invalid-arn"
  recovery_approach = "activePassive"
  regions           = [%[2]q, %[3]q]
  primary_region    = %[2]q

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = "arn:aws:iam::123456789012:role/test"
          timeout_minutes = 60
        }
      }
    }
  }
}
`, rName, acctest.Region(), acctest.AlternateRegion())
}

func testAccPlanConfig_singleRegion(rName string) string {
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
  regions           = [%[2]q]
  primary_region    = %[2]q

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "single-region-step"
      execution_block_type = "ManualApproval"

      execution_block_configuration {
        execution_approval_config {
          approval_role   = aws_iam_role.test.arn
          timeout_minutes = 60
        }
      }
    }
  }
}
`, rName, acctest.Region())
}
