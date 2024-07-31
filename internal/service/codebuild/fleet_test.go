// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodebuild "github.com/hashicorp/terraform-provider-aws/internal/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeBuildFleet_basic(t *testing.T) {
	ctx := context.Background()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_fleet.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "LINUX_CONTAINER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "overflow_behavior", "ON_DEMAND"),
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

func TestAccCodeBuildFleet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_codebuild_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodebuild.ResourceFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeBuildFleet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccFleetConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFleetConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCodeBuildFleet_updateBasicParameters(t *testing.T) {
	ctx := context.Background()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_fleet.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "LINUX_CONTAINER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "overflow_behavior", "ON_DEMAND"),
				),
			},
			{
				Config: testAccFleetConfig_updateBaseCapacity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "LINUX_CONTAINER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "overflow_behavior", "ON_DEMAND"),
				),
			},
			{
				Config: testAccFleetConfig_updateComputeType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_LARGE"),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "LINUX_CONTAINER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "overflow_behavior", "ON_DEMAND"),
				),
			},
			{
				Config: testAccFleetConfig_updateEnvironmentType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_LARGE"),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "ARM_CONTAINER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "overflow_behavior", "ON_DEMAND"),
				),
			},
		},
	})
}

func TestAccCodeBuildFleet_updateScalingConfiguration(t *testing.T) {
	ctx := context.Background()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codebuild_fleet.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeBuildServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_scalingConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "ARM_CONTAINER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "overflow_behavior", "QUEUE"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.scaling_type", "TARGET_TRACKING_SCALING"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.metric_type", "FLEET_UTILIZATION_RATE"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.target_value", "97.5"),
				),
			},
			{
				Config: testAccFleetConfig_noScalingConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "ARM_CONTAINER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "overflow_behavior", "ON_DEMAND"),
				),
			},
		},
	})
}

func testAccCheckFleetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codebuild_fleet" {
				continue
			}

			_, err := tfcodebuild.FindFleetByARNOrNames(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeBuild Fleet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFleetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeBuildClient(ctx)

		fleet, err := tfcodebuild.FindFleetByARNOrNames(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if len(fleet.Fleets) == 0 {
			return fmt.Errorf("Fleet not found: %s", rs.Primary.ID)
		}

		expectedName := rs.Primary.Attributes["name"]
		if *fleet.Fleets[0].Name != expectedName {
			return fmt.Errorf("Fleet name mismatch, expected: %s, got: %s", expectedName, *fleet.Fleets[0].Name)
		}

		return nil
	}
}

func testAccFleetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = %q
  overflow_behavior = "ON_DEMAND"
}
`, rName)
}

func testAccFleetConfig_updateBaseCapacity(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 2
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = %q
  overflow_behavior = "ON_DEMAND"
}
`, rName)
}

func testAccFleetConfig_updateComputeType(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_LARGE"
  environment_type  = "LINUX_CONTAINER"
  name              = %q
  overflow_behavior = "ON_DEMAND"
}
`, rName)
}

func testAccFleetConfig_updateEnvironmentType(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_LARGE"
  environment_type  = "ARM_CONTAINER"
  name              = %q
  overflow_behavior = "ON_DEMAND"
}
`, rName)
}

func testAccFleetConfig_scalingConfiguration(rName string) string {
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
`, rName)
}

func testAccFleetConfig_noScalingConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "ARM_CONTAINER"
  name              = %q
  overflow_behavior = "ON_DEMAND"
}
`, rName)
}

func testAccFleetConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = %q
  overflow_behavior = "ON_DEMAND"
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccFleetConfig_tags2(rName string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = %q
  overflow_behavior = "ON_DEMAND"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
