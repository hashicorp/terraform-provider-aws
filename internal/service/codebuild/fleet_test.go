// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
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
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "LINUX_CONTAINER"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
				Config: testAccFleetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccFleetConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFleetConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCodeBuildFleet_baseCapacity(t *testing.T) {
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
				Config: testAccFleetConfig_baseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "1"),
				),
			},
			{
				Config: testAccFleetConfig_baseCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "base_capacity", "2"),
				),
			},
		},
	})
}

func TestAccCodeBuildFleet_computeConfiguration(t *testing.T) {
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
				Config: testAccFleetConfig_computeConfiguration(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "compute_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_configuration.0.machine_type", string(types.MachineTypeGeneral)),
					resource.TestCheckResourceAttr(resourceName, "compute_configuration.0.vcpu", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFleetConfig_computeConfiguration(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "compute_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_configuration.0.machine_type", string(types.MachineTypeGeneral)),
					resource.TestCheckResourceAttr(resourceName, "compute_configuration.0.vcpu", "4"),
				),
			},
		},
	})
}

func TestAccCodeBuildFleet_computeType(t *testing.T) {
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
				Config: testAccFleetConfig_computeType(rName, types.ComputeTypeBuildGeneral1Small),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_SMALL"),
				),
			},
			{
				Config: testAccFleetConfig_computeType(rName, types.ComputeTypeBuildGeneral1Medium),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "compute_type", "BUILD_GENERAL1_MEDIUM"),
				),
			},
		},
	})
}

func TestAccCodeBuildFleet_environmentType(t *testing.T) {
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
				Config: testAccFleetConfig_environmentType(rName, types.EnvironmentTypeLinuxContainer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "LINUX_CONTAINER"),
				),
			},
			{
				Config: testAccFleetConfig_environmentType(rName, types.EnvironmentTypeArmContainer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "environment_type", "ARM_CONTAINER"),
				),
			},
		},
	})
}

func TestAccCodeBuildFleet_imageId(t *testing.T) {
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
				Config: testAccFleetConfig_imageId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_id", "aws/codebuild/macos-arm-base:14"),
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

func TestAccCodeBuildFleet_scalingConfiguration(t *testing.T) {
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
				Config: testAccFleetConfig_scalingConfiguration1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.scaling_type", "TARGET_TRACKING_SCALING"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.target_tracking_scaling_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.metric_type", "FLEET_UTILIZATION_RATE"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.target_value", "97.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFleetConfig_scalingConfiguration2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.max_capacity", "3"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.scaling_type", "TARGET_TRACKING_SCALING"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.target_tracking_scaling_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.metric_type", "FLEET_UTILIZATION_RATE"),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.0.target_tracking_scaling_configs.0.target_value", "90.5"),
				),
			},
			{
				Config: testAccFleetConfig_scalingConfiguration3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scaling_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccCodeBuildFleet_vpcConfig(t *testing.T) {
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
				Config: testAccFleetConfig_vpcConfig2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.subnets.0", "aws_subnet.test.1", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexache.MustCompile(`^vpc-`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFleetConfig_vpcConfig1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.subnets.0", "aws_subnet.test.0", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexache.MustCompile(`^vpc-`)),
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

			_, err := tfcodebuild.FindFleetByARN(ctx, conn, rs.Primary.ID)

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

		_, err := tfcodebuild.FindFleetByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccFleetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = %[1]q
  overflow_behavior = "ON_DEMAND"
}
`, rName)
}

func testAccFleetConfig_baseCapacity(rName string, baseCapacity int) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = %[2]d
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = %[1]q
  overflow_behavior = "ON_DEMAND"
}
`, rName, baseCapacity)
}

func testAccFleetConfig_computeConfiguration(rName string, vcpu int) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity    = 1
  compute_type     = "ATTRIBUTE_BASED_COMPUTE"
  environment_type = "LINUX_CONTAINER"
  name             = %[1]q

  compute_configuration {
    machine_type = "GENERAL"
    vcpu         = %[2]d
  }
}
`, rName, vcpu)
}

func testAccFleetConfig_computeType(rName string, computeType types.ComputeType) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = %[2]q
  environment_type  = "LINUX_CONTAINER"
  name              = %[1]q
  overflow_behavior = "ON_DEMAND"
}
`, rName, string(computeType))
}

func testAccFleetConfig_environmentType(rName string, environmentType types.EnvironmentType) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_LARGE"
  environment_type  = %[2]q
  name              = %[1]q
  overflow_behavior = "ON_DEMAND"
}
`, rName, string(environmentType))
}

func testAccFleetConfig_imageId(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_MEDIUM"
  environment_type  = "MAC_ARM"
  name              = %[1]q
  overflow_behavior = "QUEUE"
  image_id          = "aws/codebuild/macos-arm-base:14"
}
`, rName)
}

func testAccFleetConfig_scalingConfiguration1(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "ARM_CONTAINER"
  name              = %[1]q
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

func testAccFleetConfig_scalingConfiguration2(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "ARM_CONTAINER"
  name              = %[1]q
  overflow_behavior = "QUEUE"

  scaling_configuration {
    max_capacity = 3
    scaling_type = "TARGET_TRACKING_SCALING"

    target_tracking_scaling_configs {
      metric_type  = "FLEET_UTILIZATION_RATE"
      target_value = 90.5
    }
  }
}
`, rName)
}

func testAccFleetConfig_scalingConfiguration3(rName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "ARM_CONTAINER"
  name              = %[1]q
  overflow_behavior = "QUEUE"
}
`, rName)
}

func testAccFleetConfig_baseFleetServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "codebuild.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": "*",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:CreateNetworkInterfacePermission",
        "ec2:DescribeDhcpOptions",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeVpcs",
        "ec2:ModifyNetworkInterfaceAttribute",
        "ec2:DeleteNetworkInterface"
      ]
    }
  ]
}
POLICY
}
`, rName)
}

func testAccFleetConfig_baseVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccFleetConfig_vpcConfig1(rName string) string {
	return acctest.ConfigCompose(
		testAccFleetConfig_baseFleetServiceRole(rName),
		testAccFleetConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity      = 1
  compute_type       = "BUILD_GENERAL1_SMALL"
  environment_type   = "LINUX_CONTAINER"
  name               = %[1]q
  overflow_behavior  = "ON_DEMAND"
  fleet_service_role = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test[0].id]
    vpc_id             = aws_vpc.test.id
  }
}
`, rName))
}

func testAccFleetConfig_vpcConfig2(rName string) string {
	return acctest.ConfigCompose(
		testAccFleetConfig_baseFleetServiceRole(rName),
		testAccFleetConfig_baseVPC(rName),
		fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity      = 1
  compute_type       = "BUILD_GENERAL1_SMALL"
  environment_type   = "LINUX_CONTAINER"
  name               = %[1]q
  overflow_behavior  = "ON_DEMAND"
  fleet_service_role = aws_iam_role.test.arn

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test[1].id]
    vpc_id             = aws_vpc.test.id
  }
}
`, rName))
}

func testAccFleetConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = %[1]q
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
  name              = %[1]q
  overflow_behavior = "ON_DEMAND"
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
