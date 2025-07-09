// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// lintignore:AWSAT003,AWSAT005
package arcregionswitch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	sdktypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfarcregionswitch "github.com/hashicorp/terraform-provider-aws/internal/service/arcregionswitch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccARCRegionSwitchPlan_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var plan sdktypes.Plan
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
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
				Config: testAccPlanConfig_basic(rName, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtName, rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_approach", "activePassive"),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "primary_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.step.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.step.0.execution_block_type", "ManualApproval"),
					resource.TestCheckResourceAttr(resourceName, "workflow.1.step.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow.1.step.0.execution_block_type", "ManualApproval"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role"),
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

func testAccCheckPlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ARCRegionSwitchClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_arcregionswitch_plan" {
				continue
			}

			_, err := tfarcregionswitch.FindPlanByARN(ctx, conn, rs.Primary.ID)

			if err == nil {
				return fmt.Errorf("ARC Region Switch Plan %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckPlanExists(ctx context.Context, n string, v *sdktypes.Plan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Plan not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ARC Region Switch Plan ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ARCRegionSwitchClient(ctx)

		output, err := tfarcregionswitch.FindPlanByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ARCRegionSwitchClient(ctx)

	input := arcregionswitch.ListPlansInput{}
	_, err := conn.ListPlans(ctx, &input)

	if err != nil {
		t.Skipf("skipping acceptance testing: %s", err)
	}
}

func TestAccARCRegionSwitchPlan_update(t *testing.T) {
	ctx := acctest.Context(t)
	var plan sdktypes.Plan
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
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
				Config: testAccPlanConfig_update(rName, "Initial description", 30, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "recovery_time_objective_minutes", "30"),
					resource.TestCheckResourceAttr(resourceName, "associated_alarms.#", "1"),
				),
			},
			{
				Config: testAccPlanConfig_update(rName, "Updated description", 60, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "recovery_time_objective_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "associated_alarms.#", "2"),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlan_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var plan sdktypes.Plan
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
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
				Config: testAccPlanConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccPlanConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPlanConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlan_singleRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var plan sdktypes.Plan
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
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
				Config: testAccPlanConfig_minimalRegions(rName, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "primary_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", "2"),
				),
			},
		},
	})
}

func TestAccARCRegionSwitchPlan_multipleWorkflowsSameAction(t *testing.T) {
	ctx := acctest.Context(t)
	var plan sdktypes.Plan
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
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
				Config: testAccPlanConfig_multipleWorkflowsSameAction(rName, acctest.AlternateRegion(), acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*", map[string]string{
						"workflow_target_action": "activate",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "workflow.*", map[string]string{
						"workflow_target_action": "deactivate",
					}),
				),
			},
		},
	})
}

func testAccPlanConfig_basic(rName, primaryRegion, alternateRegion string) string {
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
    workflow_target_region = %[2]q

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
`, rName, primaryRegion, alternateRegion)
}

func testAccPlanConfig_update(rName, description string, rto int, primaryRegion, alternateRegion string) string {
	alarms := `
  associated_alarms {
    name                = "test-alarm-1"
    alarm_type          = "applicationHealth"
    resource_identifier = "arn:aws:cloudwatch:%[4]s:123456789012:alarm:test-alarm-1"
  }`

	if rto == 60 {
		alarms += `
  associated_alarms {
    name                = "test-alarm-2"
    alarm_type          = "applicationHealth"
    resource_identifier = "arn:aws:cloudwatch:%[4]s:123456789012:alarm:test-alarm-2"
  }`
	}

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
  name                            = %[1]q
  execution_role                  = aws_iam_role.test.arn
  recovery_approach               = "activePassive"
  regions                         = [%[4]q, %[5]q]
  primary_region                  = %[4]q
  description                     = %[2]q
  recovery_time_objective_minutes = %[3]d
%[6]s
  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[5]q

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
    workflow_target_region = %[4]q

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
`, rName, description, rto, primaryRegion, alternateRegion, alarms)
}

func testAccPlanConfig_tags1(rName, tagKey1, tagValue1, primaryRegion, alternateRegion string) string {
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
  regions           = [%[4]q, %[5]q]
  primary_region    = %[4]q

  tags = {
    %[2]q = %[3]q
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[5]q

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
    workflow_target_region = %[4]q

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
`, rName, tagKey1, tagValue1, primaryRegion, alternateRegion)
}

func testAccPlanConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2, primaryRegion, alternateRegion string) string {
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
  regions           = [%[6]q, %[7]q]
  primary_region    = %[6]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[7]q

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
    workflow_target_region = %[6]q

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
`, rName, tagKey1, tagValue1, tagKey2, tagValue2, primaryRegion, alternateRegion)
}

func testAccPlanConfig_minimalRegions(rName, primaryRegion, alternateRegion string) string {
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
      name                 = "minimal-step-secondary"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = %[2]q

    step {
      name                 = "minimal-step-primary"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }
}
`, rName, primaryRegion, alternateRegion)
}

func testAccPlanConfig_multipleWorkflowsSameAction(rName, primaryRegion, alternateRegion string) string {
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
  recovery_approach = "activeActive"
  regions           = [%[2]q, %[3]q]
  primary_region    = %[2]q

  workflow {
    workflow_target_action = "activate"

    step {
      name                 = "activate-step"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

  workflow {
    workflow_target_action = "deactivate"

    step {
      name                 = "deactivate-step"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }
}
`, rName, primaryRegion, alternateRegion)
}
