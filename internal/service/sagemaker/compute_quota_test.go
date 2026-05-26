// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	computeQuotaClusterARNEnvVar   = "AWS_SAGEMAKER_HYPERPOD_CLUSTER_ARN"
	computeQuotaInstanceTypeEnvVar = "AWS_SAGEMAKER_HYPERPOD_INSTANCE_TYPE"
)

func TestAccSageMakerComputeQuota_basic(t *testing.T) {
	ctx := acctest.Context(t)
	clusterARN := acctest.SkipIfEnvVarNotSet(t, computeQuotaClusterARNEnvVar)

	var computeQuota sagemaker.DescribeComputeQuotaOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_compute_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeQuotaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeQuotaConfig_basic(rName, clusterARN, testAccComputeQuotaInstanceType()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeQuotaExists(ctx, t, resourceName, &computeQuota),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "activation_state", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "compute_quota_config.0.compute_quota_resources.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_quota_config.0.resource_sharing_config.0.borrow_limit", "50"),
					resource.TestCheckResourceAttr(resourceName, "compute_quota_config.0.resource_sharing_config.0.strategy", "LendAndBorrow"),
					resource.TestCheckResourceAttr(resourceName, "compute_quota_target.0.team_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "compute_quota_version"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Created"),
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

func TestAccSageMakerComputeQuota_update(t *testing.T) {
	ctx := acctest.Context(t)
	clusterARN := acctest.SkipIfEnvVarNotSet(t, computeQuotaClusterARNEnvVar)

	var computeQuota1, computeQuota2 sagemaker.DescribeComputeQuotaOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_compute_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeQuotaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeQuotaConfig_basic(rName, clusterARN, testAccComputeQuotaInstanceType()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeQuotaExists(ctx, t, resourceName, &computeQuota1),
				),
			},
			{
				Config: testAccComputeQuotaConfig_updated(rName, clusterARN, testAccComputeQuotaInstanceType()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeQuotaExists(ctx, t, resourceName, &computeQuota2),
					testAccCheckComputeQuotaNotRecreated(&computeQuota1, &computeQuota2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated"),
					resource.TestCheckResourceAttr(resourceName, "compute_quota_config.0.resource_sharing_config.0.borrow_limit", "100"),
					resource.TestCheckResourceAttr(resourceName, "compute_quota_target.0.fair_share_weight", "1"),
				),
			},
		},
	})
}

func testAccCheckComputeQuotaDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_compute_quota" {
				continue
			}

			_, err := tfsagemaker.FindComputeQuotaByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameComputeQuota, rs.Primary.ID, err)
			}

			return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameComputeQuota, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckComputeQuotaExists(ctx context.Context, t *testing.T, name string, computeQuota *sagemaker.DescribeComputeQuotaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameComputeQuota, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameComputeQuota, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)
		resp, err := tfsagemaker.FindComputeQuotaByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameComputeQuota, rs.Primary.ID, err)
		}

		*computeQuota = *resp

		return nil
	}
}

func testAccCheckComputeQuotaNotRecreated(before, after *sagemaker.DescribeComputeQuotaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.ComputeQuotaId), aws.ToString(after.ComputeQuotaId); before != after {
			return create.Error(names.SageMaker, create.ErrActionCheckingNotRecreated, tfsagemaker.ResNameComputeQuota, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccComputeQuotaInstanceType() string {
	if v := os.Getenv(computeQuotaInstanceTypeEnvVar); v != "" {
		return v
	}

	return "ml.m5.2xlarge"
}

func testAccComputeQuotaConfig_basic(rName, clusterARN, instanceType string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_compute_quota" "test" {
  cluster_arn = %[2]q
  name        = %[1]q

  compute_quota_config {
    compute_quota_resources {
      count         = 1
      instance_type = %[3]q
    }

    resource_sharing_config {
      strategy = "LendAndBorrow"
    }
  }

  compute_quota_target {
    team_name = %[1]q
  }
}
`, rName, clusterARN, instanceType)
}

func testAccComputeQuotaConfig_updated(rName, clusterARN, instanceType string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_compute_quota" "test" {
  cluster_arn = %[2]q
  description = "updated"
  name        = %[1]q

  compute_quota_config {
    compute_quota_resources {
      count         = 1
      instance_type = %[3]q
    }

    resource_sharing_config {
      borrow_limit = 100
      strategy     = "LendAndBorrow"
    }
  }

  compute_quota_target {
    fair_share_weight = 1
    team_name         = %[1]q
  }
}
`, rName, clusterARN, instanceType)
}
