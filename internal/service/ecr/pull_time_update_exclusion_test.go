// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRPullTimeUpdateExclusion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_pull_time_update_exclusion.test"
	iamUserResourceName := "aws_iam_user.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECREndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullTimeUpdateExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullTimeUpdateExclusionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPullTimeUpdateExclusionExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal_arn", iamUserResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, iamUserResourceName, names.AttrARN),
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

func TestAccECRPullTimeUpdateExclusion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_pull_time_update_exclusion.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECREndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullTimeUpdateExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullTimeUpdateExclusionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPullTimeUpdateExclusionExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfecr.ResourcePullTimeUpdateExclusion, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRPullTimeUpdateExclusion_role(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_pull_time_update_exclusion.test"
	iamRoleResourceName := "aws_iam_role.test"
	iamUserResourceName := "aws_iam_user.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECREndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPullTimeUpdateExclusionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPullTimeUpdateExclusionConfig_role(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPullTimeUpdateExclusionExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, iamRoleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPullTimeUpdateExclusionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPullTimeUpdateExclusionExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal_arn", iamUserResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, iamUserResourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckPullTimeUpdateExclusionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_pull_time_update_exclusion" {
				continue
			}

			found, err := tfecr.FindPullTimeUpdateExclusionByPrincipalARN(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ECR, create.ErrActionCheckingDestroyed, tfecr.ResNamePullTimeUpdateExclusion, rs.Primary.ID, err)
			}

			if found {
				return create.Error(names.ECR, create.ErrActionCheckingDestroyed, tfecr.ResNamePullTimeUpdateExclusion, rs.Primary.ID, errors.New("not destroyed"))
			}
		}

		return nil
	}
}

func testAccCheckPullTimeUpdateExclusionExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ECR, create.ErrActionCheckingExistence, tfecr.ResNamePullTimeUpdateExclusion, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ECR, create.ErrActionCheckingExistence, tfecr.ResNamePullTimeUpdateExclusion, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

		found, err := tfecr.FindPullTimeUpdateExclusionByPrincipalARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ECR, create.ErrActionCheckingExistence, tfecr.ResNamePullTimeUpdateExclusion, rs.Primary.ID, err)
		}

		if !found {
			return create.Error(names.ECR, create.ErrActionCheckingExistence, tfecr.ResNamePullTimeUpdateExclusion, rs.Primary.ID, errors.New("not found"))
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECRClient(ctx)

	input := &ecr.ListPullTimeUpdateExclusionsInput{}

	_, err := conn.ListPullTimeUpdateExclusions(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPullTimeUpdateExclusionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccPullTimeUpdateExclusionConfig_iamUser(rName),
		testAccPullTimeUpdateExclusionConfig_exclusion("aws_iam_user.test.arn"),
	)
}

func testAccPullTimeUpdateExclusionConfig_role(rName string) string {
	return acctest.ConfigCompose(
		testAccPullTimeUpdateExclusionConfig_iamRole(rName),
		testAccPullTimeUpdateExclusionConfig_exclusion("aws_iam_role.test.arn"),
	)
}

func testAccPullTimeUpdateExclusionConfig_iamUser(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_user_policy" "test" {
  name = "ecr-pull-policy"
  user = aws_iam_user.test.name

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
`, rName)
}

func testAccPullTimeUpdateExclusionConfig_iamRole(rName string) string {
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
  name = "ecr-pull-policy"
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
`, rName)
}

func testAccPullTimeUpdateExclusionConfig_exclusion(principalARNRef string) string {
	return fmt.Sprintf(`
resource "aws_ecr_pull_time_update_exclusion" "test" {
  principal_arn = %[1]s
}
`, principalARNRef)
}
