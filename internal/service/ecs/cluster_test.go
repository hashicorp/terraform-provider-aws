// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccECSCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ecs", fmt.Sprintf("cluster/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "service_connect_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "setting.*", map[string]string{
						"name":  "containerInsights",
						"value": "disabled",
					}),
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
	ctx := acctest.Context(t)
	var v ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecs.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
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
				Config: testAccClusterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccECSCluster_serviceConnectDefaults(t *testing.T) {
	ctx := acctest.Context(t)
	var v ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ns := fmt.Sprintf("%s-%s", acctest.ResourcePrefix, sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha))
	resourceName := "aws_ecs_cluster.test"
	namespace1ResourceName := "aws_service_discovery_http_namespace.test.0"
	namespace2ResourceName := "aws_service_discovery_http_namespace.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_serviceConnectDefaults(rName, ns, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "service_connect_defaults.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "service_connect_defaults.0.namespace", namespace1ResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_serviceConnectDefaults(rName, ns, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "service_connect_defaults.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "service_connect_defaults.0.namespace", namespace2ResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccECSCluster_containerInsights(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ecs", fmt.Sprintf("cluster/%s", rName)),
				),
			},
			{
				Config: testAccClusterConfig_containerInsights(rName, "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
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
				Config: testAccClusterConfig_containerInsights(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
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
	ctx := acctest.Context(t)
	var cluster1 ecs.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_configuration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
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
				Config: testAccClusterConfig_configuration(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
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

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_cluster" {
				continue
			}

			_, err := tfecs.FindClusterByNameOrARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECS Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, n string, v *ecs.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECS Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn(ctx)

		output, err := tfecs.FindClusterByNameOrARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}
`, rName)
}

func testAccClusterConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccClusterConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccClusterConfig_serviceConnectDefaults(rName, ns string, idx int) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  count = 2

  name = "%[2]s-${count.index}"
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q

  service_connect_defaults {
    namespace = aws_service_discovery_http_namespace.test[%[3]d].arn
  }
}
`, rName, ns, idx)
}

func testAccClusterConfig_containerInsights(rName, value string) string {
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

func testAccClusterConfig_configuration(rName string, enable bool) string {
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
