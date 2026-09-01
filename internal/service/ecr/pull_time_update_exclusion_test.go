// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRPullTimeUpdateExclusion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_pull_time_update_exclusion.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECREndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullTimeUpdateExclusionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPullTimeUpdateExclusionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPullTimeUpdateExclusionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "principal_arn"),
				ImportStateVerifyIdentifierAttribute: "principal_arn",
			},
		},
	})
}

func TestAccECRPullTimeUpdateExclusion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_pull_time_update_exclusion.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECREndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullTimeUpdateExclusionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPullTimeUpdateExclusionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPullTimeUpdateExclusionExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecr.ResourcePullTimeUpdateExclusion, resourceName),
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

func testAccCheckPullTimeUpdateExclusionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_pull_time_update_exclusion" {
				continue
			}

			err := tfecr.FindPullTimeUpdateExclusionByPrincipalARN(ctx, conn, rs.Primary.Attributes["principal_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Pull Time Update Exclusion %s still exists", rs.Primary.Attributes["principal_arn"])
		}

		return nil
	}
}

func testAccCheckPullTimeUpdateExclusionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

		return tfecr.FindPullTimeUpdateExclusionByPrincipalARN(ctx, conn, rs.Primary.Attributes["principal_arn"])
	}
}

func testAccPullTimeUpdateExclusionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%[1]s-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = "%[1]s-policy"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_ecr_pull_time_update_exclusion" "test" {
  principal_arn = aws_iam_role.test.arn
}
`, rName)
}
