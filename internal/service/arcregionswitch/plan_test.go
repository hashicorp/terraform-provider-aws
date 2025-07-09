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
				Config: testAccPlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlanExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "recovery_approach", "activePassive"),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "primary_region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.step.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.step.0.execution_block_type", "ManualApproval"),
					resource.TestCheckResourceAttr(resourceName, "workflow.1.step.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workflow.1.step.0.execution_block_type", "ManualApproval"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
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

	input := &arcregionswitch.ListPlansInput{}
	_, err := conn.ListPlans(ctx, input)

	if err != nil {
		t.Skipf("skipping acceptance testing: %s", err)
	}
}

func testAccPlanConfig_basic(rName string) string {
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
  regions           = ["us-east-1", "us-west-2"]
  primary_region    = "us-east-1"

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = "us-west-2"

    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = "us-east-1"

    step {
      name                 = "basic-step-east"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }
}
`, rName)
}
