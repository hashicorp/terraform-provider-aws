// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codebuild_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeBuildFleetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_fleet.test"
	datasourceName := "data.aws_codebuild_fleet.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "base_capacity", resourceName, "base_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "compute_type", resourceName, "compute_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "created", resourceName, "created"),
					resource.TestCheckResourceAttrPair(datasourceName, "environment_type", resourceName, "environment_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "last_modified", resourceName, "last_modified"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "overflow_behavior", resourceName, "overflow_behavior"),
					resource.TestCheckResourceAttrPair(datasourceName, "scaling_configuration.0.max_capacity", resourceName, "scaling_configuration.0.max_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "scaling_configuration.0.scaling_type", resourceName, "scaling_configuration.0.scaling_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.metric_type", resourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.metric_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.target_value", resourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.target_value"),
				),
			},
		},
	})
}

func TestAccCodeBuildFleetDataSource_customInstanceType(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codebuild_fleet.test"
	datasourceName := "data.aws_codebuild_fleet.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetDataSourceConfig_customInstanceType(rName, "t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "base_capacity", resourceName, "base_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "compute_configuration.0.disk", resourceName, "compute_configuration.0.disk"),
					resource.TestCheckResourceAttrPair(datasourceName, "compute_configuration.0.instance_type", resourceName, "compute_configuration.0.instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "compute_type", resourceName, "compute_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "created", resourceName, "created"),
					resource.TestCheckResourceAttrPair(datasourceName, "environment_type", resourceName, "environment_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "last_modified", resourceName, "last_modified"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "overflow_behavior", resourceName, "overflow_behavior"),
				),
			},
		},
	})
}

func testAccFleetDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "ARM_CONTAINER"
  name              = %q
  overflow_behavior = "QUEUE"
  scaling_configuration {
    max_capacity = 2
    scaling_type = "TARGET_TRACKING_SCALING"
    target_tracking_scaling_configs {
      metric_type  = "FLEET_UTILIZATION_RATE"
      target_value = 97.5
    }
  }
}

data "aws_codebuild_fleet" "test" {
  name = aws_codebuild_fleet.test.name
}
`, rName)
}

func testAccFleetDataSourceConfig_customInstanceType(rName, instanceType string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity = 1
  compute_type  = "CUSTOM_INSTANCE_TYPE"
  compute_configuration {
    instance_type = %[2]q
  }
  environment_type  = "LINUX_CONTAINER"
  name              = %[1]q
  overflow_behavior = "QUEUE"
}

data "aws_codebuild_fleet" "test" {
  name = aws_codebuild_fleet.test.name
}
`, rName, instanceType)
}
