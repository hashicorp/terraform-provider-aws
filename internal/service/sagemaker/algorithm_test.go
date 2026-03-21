// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerAlgorithm_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var algorithm sagemaker.DescribeAlgorithmOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_algorithm.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlgorithmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlgorithmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName, &algorithm),
					resource.TestCheckResourceAttr(resourceName, "algorithm_name", rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`algorithm/.+`)),
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

func TestAccSageMakerAlgorithm_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_algorithm.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlgorithmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAlgorithmConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlgorithmExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsagemaker.ResourceAlgorithm, resourceName),
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

func testAccCheckAlgorithmDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_algorithm" {
				continue
			}

			_, err := tfsagemaker.FindAlgorithmByName(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameAlgorithm, rs.Primary.ID, err)
			}

			return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameAlgorithm, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAlgorithmExists(ctx context.Context, t *testing.T, name string, outputs ...*sagemaker.DescribeAlgorithmOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameAlgorithm, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameAlgorithm, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		output, err := tfsagemaker.FindAlgorithmByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameAlgorithm, rs.Primary.ID, err)
		}

		if len(outputs) > 0 && outputs[0] != nil {
			*outputs[0] = *output
		}

		return nil
	}
}

func testAccAlgorithmConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "linear-learner"
  image_tag       = "1"
}

resource "aws_sagemaker_algorithm" "test" {
  algorithm_name = %[1]q

  training_specification {
    training_image                     = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    supported_training_instance_types  = ["ml.m5.large"]

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }
}
`, rName)
}
