package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
)

func TestAccECSCluster_basic(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ecs", fmt.Sprintf("cluster/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECSCluster_disappears(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					acctest.CheckResourceDisappears(acctest.Provider, tfecs.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSCluster_tags(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccECSCluster_singleCapacityProvider(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	providerName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSingleCapacityProvider(rName, providerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECSCluster_capacityProviders(t *testing.T) {
	var cluster ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProviders(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccClusterCapacityProvidersReOrdered(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccECSCluster_capacityProvidersUpdate(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersFargate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersFargateSpot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
				),
			},
			{
				Config: testAccClusterCapacityProvidersFargateBoth(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
				),
			},
		},
	})
}

func TestAccECSCluster_capacityProvidersNoStrategy(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersFargateNoStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersFargateSpotNoStrategy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
				),
			},
		},
	})
}

func TestAccECSCluster_containerInsights(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ecs", fmt.Sprintf("cluster/%s", rName)),
				),
			},
			{
				Config: testAccClusterContainerInsightsConfig(rName, "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ecs", fmt.Sprintf("cluster/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "setting.*", map[string]string{
						"name":  "containerInsights",
						"value": "enabled",
					}),
				),
			},
			{
				Config: testAccClusterContainerInsightsConfig(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "setting.*", map[string]string{
						"name":  "containerInsights",
						"value": "disabled",
					}),
				),
			},
		},
	})
}

func TestAccECSCluster_configuration(t *testing.T) {
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigurationConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.execute_command_configuration.0.kms_key_id", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.logging", "OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_encryption_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_log_group_name", "aws_cloudwatch_log_group.test", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfigurationConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.execute_command_configuration.0.kms_key_id", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.logging", "OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_encryption_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_log_group_name", "aws_cloudwatch_log_group.test", "name"),
				),
			},
		},
	})
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_cluster" {
			continue
		}

		c, err := tfecs.FindClusterByNameOrARN(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if aws.StringValue(c.ClusterArn) == rs.Primary.ID && aws.StringValue(c.Status) != "INACTIVE" {
			return fmt.Errorf("ECS cluster still exists:\n%s", c)
		}
	}

	return nil
}

func testAccCheckClusterExists(resourceName string, cluster *ecs.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn
		c, err := tfecs.FindClusterByNameOrARN(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading ECS Cluster (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(c.ClusterArn) == rs.Primary.ID && aws.StringValue(c.Status) != "INACTIVE" {
			*cluster = *c
			return nil
		}

		return fmt.Errorf("ECS Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccClusterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q
}
`, rName)
}

func testAccClusterTags1Config(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q

  tags = {
    %q = %q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccClusterCapacityProviderConfig(rName string) string {
	return testAccCapacityProviderBaseConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName)
}

func testAccClusterSingleCapacityProvider(rName, providerName string) string {
	return testAccClusterCapacityProviderConfig(providerName) + fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = [aws_ecs_capacity_provider.test.name]

  default_capacity_provider_strategy {
    base              = 1
    capacity_provider = aws_ecs_capacity_provider.test.name
    weight            = 1
  }
}
`, rName)
}

func testAccClusterCapacityProviders(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE_SPOT", "FARGATE"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
    base              = 1
  }

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
}
`, rName)
}

func testAccClusterCapacityProvidersReOrdered(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
    base              = 1
  }
}
`, rName)
}

func testAccClusterCapacityProvidersFargate(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base              = 1
    capacity_provider = "FARGATE"
    weight            = 1
  }
}
`, rName)
}

func testAccClusterCapacityProvidersFargateSpot(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }
}
`, rName)
}

func testAccClusterCapacityProvidersFargateBoth(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }
}
`, rName)
}

func testAccClusterCapacityProvidersFargateNoStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE"]
}
`, rName)
}

func testAccClusterCapacityProvidersFargateSpotNoStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE_SPOT"]
}
`, rName)
}

func testAccClusterTags2Config(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccClusterContainerInsightsConfig(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
  setting {
    name  = "containerInsights"
    value = %[2]q
  }
}
`, rName, value)
}

func testAccClusterConfigurationConfig(rName string, enable bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q

  configuration {
    execute_command_configuration {
      kms_key_id = aws_kms_key.test.arn
      logging    = "OVERRIDE"

      log_configuration {
        cloud_watch_encryption_enabled = %[2]t
        cloud_watch_log_group_name     = aws_cloudwatch_log_group.test.name
      }
    }
  }
}
`, rName, enable)
}
