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
		ErrorCheck:               acctest.ErrorCheck(t, names.ARCRegionSwitch),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlanDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_role", resourceName, "execution_role"),
					resource.TestCheckResourceAttrPair(dataSourceName, "recovery_approach", resourceName, "recovery_approach"),
					resource.TestCheckResourceAttrPair(dataSourceName, "regions", resourceName, "regions"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workflow", resourceName, "workflow"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Environment", "test"),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlanDataSource_regionOverride(t *testing.T) {
	dataSourceName := "data.aws_arcregionswitch_plan.test"
	resourceName := "aws_arcregionswitch_plan.test"

	t.Parallel()

	testcases := []string{acctest.Region(), acctest.AlternateRegion()}

	for _, resourceRegion := range testcases {
		t.Run(resourceRegion, func(t *testing.T) {
			t.Parallel()

			for _, datasourceRegion := range testcases {
				t.Run(datasourceRegion, func(t *testing.T) {
					ctx := acctest.Context(t)
					rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

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
								// Cross-region test cases will succeed because `aws_arcregionswitch_plan` is global
								Config: testAccPlanDataSourceConfig_regionOverride(rName, resourceRegion, datasourceRegion),
								Check: resource.ComposeTestCheckFunc(
									resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
								),
							},
						},
					})
				})
			}
		})
	}
}

func testAccPlanDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_arcregionswitch_plan" "test" {
  arn = aws_arcregionswitch_plan.test.arn
}

resource "aws_arcregionswitch_plan" "test" {
  name              = %[1]q
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [%[3]q, %[2]q]
  primary_region    = %[3]q

  tags = {
    Name        = %[1]q
    Environment = "test"
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "basic-step-primary"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }
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
`, rName, acctest.AlternateRegion(), acctest.Region())
}

func testAccPlanDataSourceConfig_regionOverride(rName, resourceRegion, dataSourceRegion string) string {
	return fmt.Sprintf(`
data "aws_arcregionswitch_plan" "test" {
  region = %[5]q

  arn = aws_arcregionswitch_plan.test.arn
}

resource "aws_arcregionswitch_plan" "test" {
  region = %[4]q

  name              = %[1]q
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [%[3]q, %[2]q]
  primary_region    = %[3]q

  tags = {
    Name        = %[1]q
    Environment = "test"
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[3]q

    step {
      name                 = "basic-step-primary"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }
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
`, rName, acctest.AlternateRegion(), acctest.Region(), resourceRegion, dataSourceRegion)
}
