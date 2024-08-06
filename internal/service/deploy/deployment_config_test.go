// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodedeploy "github.com/hashicorp/terraform-provider-aws/internal/service/deploy"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDeployDeploymentConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config types.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfigConfig_fleet(rName, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentConfigExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "deployment_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", acctest.Ct0),
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

func TestAccDeployDeploymentConfig_fleetPercent(t *testing.T) {
	ctx := acctest.Context(t)
	var config1, config2 types.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfigConfig_fleet(rName, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentConfigExists(ctx, resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.type", "FLEET_PERCENT"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.value", "75"),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", acctest.Ct0),
				),
			},
			{
				Config: testAccDeploymentConfigConfig_fleet(rName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentConfigExists(ctx, resourceName, &config2),
					testAccCheckDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.type", "FLEET_PERCENT"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.value", "50"),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", acctest.Ct0),
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

func TestAccDeployDeploymentConfig_hostCount(t *testing.T) {
	ctx := acctest.Context(t)
	var config1, config2 types.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfigConfig_hostCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentConfigExists(ctx, resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.type", "HOST_COUNT"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.value", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", acctest.Ct0),
				),
			},
			{
				Config: testAccDeploymentConfigConfig_hostCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentConfigExists(ctx, resourceName, &config2),
					testAccCheckDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.type", "HOST_COUNT"),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.0.value", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", acctest.Ct0),
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

func TestAccDeployDeploymentConfig_trafficCanary(t *testing.T) {
	ctx := acctest.Context(t)
	var config1, config2 types.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfigConfig_trafficCanary(rName, 10, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentConfigExists(ctx, resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.type", "TimeBasedCanary"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.0.interval", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.0.percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", acctest.Ct0),
				),
			},
			{
				Config: testAccDeploymentConfigConfig_trafficCanary(rName, 3, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentConfigExists(ctx, resourceName, &config2),
					testAccCheckDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.type", "TimeBasedCanary"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.0.interval", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.0.percentage", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", acctest.Ct0),
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

func TestAccDeployDeploymentConfig_trafficLinear(t *testing.T) {
	ctx := acctest.Context(t)
	var config1, config2 types.DeploymentConfigInfo
	resourceName := "aws_codedeploy_deployment_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfigConfig_trafficLinear(rName, 10, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentConfigExists(ctx, resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.type", "TimeBasedLinear"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.0.interval", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.0.percentage", "50"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", acctest.Ct0),
				),
			},
			{
				Config: testAccDeploymentConfigConfig_trafficLinear(rName, 3, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentConfigExists(ctx, resourceName, &config2),
					testAccCheckDeploymentConfigRecreated(&config1, &config2),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.type", "TimeBasedLinear"),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.0.interval", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_linear.0.percentage", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "traffic_routing_config.0.time_based_canary.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "minimum_healthy_hosts.#", acctest.Ct0),
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

func testAccCheckDeploymentConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DeployClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codedeploy_deployment_config" {
				continue
			}

			_, err := tfcodedeploy.FindDeploymentConfigByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeDeploy Deployment Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDeploymentConfigExists(ctx context.Context, n string, v *types.DeploymentConfigInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DeployClient(ctx)

		output, err := tfcodedeploy.FindDeploymentConfigByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDeploymentConfigRecreated(i, j *types.DeploymentConfigInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToTime(i.CreateTime).Equal(aws.ToTime(j.CreateTime)) {
			return errors.New("CodeDeploy Deployment Config was not recreated")
		}

		return nil
	}
}

func testAccDeploymentConfigConfig_fleet(rName string, value int) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_deployment_config" "test" {
  deployment_config_name = %[1]q

  minimum_healthy_hosts {
    type  = "FLEET_PERCENT"
    value = %[2]d
  }
}
`, rName, value)
}

func testAccDeploymentConfigConfig_hostCount(rName string, value int) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_deployment_config" "test" {
  deployment_config_name = %[1]q

  minimum_healthy_hosts {
    type  = "HOST_COUNT"
    value = %[2]d
  }
}
`, rName, value)
}

func testAccDeploymentConfigConfig_trafficCanary(rName string, interval, percentage int) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_deployment_config" "test" {
  deployment_config_name = %[1]q
  compute_platform       = "Lambda"

  traffic_routing_config {
    type = "TimeBasedCanary"

    time_based_canary {
      interval   = %[2]d
      percentage = %[3]d
    }
  }
}
`, rName, interval, percentage)
}

func testAccDeploymentConfigConfig_trafficLinear(rName string, interval, percentage int) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_deployment_config" "test" {
  deployment_config_name = %[1]q
  compute_platform       = "Lambda"

  traffic_routing_config {
    type = "TimeBasedLinear"

    time_based_linear {
      interval   = %[2]d
      percentage = %[3]d
    }
  }
}
`, rName, interval, percentage)
}
