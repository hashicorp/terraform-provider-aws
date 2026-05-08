// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerModelPackageGroupPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.GetModelPackageGroupPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelPackageGroupPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccModelPackageGroupPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelPackageGroupPolicyExists(ctx, t, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "model_package_group_name", rName),
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

func TestAccSageMakerModelPackageGroupPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.GetModelPackageGroupPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelPackageGroupPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccModelPackageGroupPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelPackageGroupPolicyExists(ctx, t, resourceName, &mpg),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceModelPackageGroupPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSageMakerModelPackageGroupPolicy_Disappears_modelPackageGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.GetModelPackageGroupPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckModelPackageGroupPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccModelPackageGroupPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelPackageGroupPolicyExists(ctx, t, resourceName, &mpg),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceModelPackageGroup(), "aws_sagemaker_model_package_group.test"),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceModelPackageGroupPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckModelPackageGroupPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_model_package_group_policy" {
				continue
			}

			_, err := tfsagemaker.FindModelPackageGroupPolicyByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("reading SageMaker AI Model Package Group Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckModelPackageGroupPolicyExists(ctx context.Context, t *testing.T, n string, mpg *sagemaker.GetModelPackageGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Model Package Group ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)
		resp, err := tfsagemaker.FindModelPackageGroupPolicyByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*mpg = *resp

		return nil
	}
}

func testAccModelPackageGroupPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid       = "AddPermModelPackageGroup"
    actions   = ["sagemaker:DescribeModelPackage", "sagemaker:ListModelPackages"]
    resources = [aws_sagemaker_model_package_group.test.arn]
    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
      type        = "AWS"
    }
  }
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q
}

resource "aws_sagemaker_model_package_group_policy" "test" {
  model_package_group_name = aws_sagemaker_model_package_group.test.model_package_group_name
  resource_policy          = jsonencode(jsondecode(data.aws_iam_policy_document.test.json))
}
`, rName)
}
