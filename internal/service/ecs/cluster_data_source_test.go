// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSClusterDataSource_ecsCluster(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "pending_tasks_count", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "registered_container_instances_count", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "running_tasks_count", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, "service_connect_defaults.#", resourceName, "service_connect_defaults.#"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccECSClusterDataSource_ecsClusterContainerInsights(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_containerInsights(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "pending_tasks_count", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "registered_container_instances_count", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "running_tasks_count", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttrPair(dataSourceName, "setting.#", resourceName, "setting.#"),
				),
			},
		},
	})
}

func TestAccECSClusterDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "pending_tasks_count", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "registered_container_instances_count", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "running_tasks_count", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, "service_connect_defaults.#", resourceName, "service_connect_defaults.#"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccClusterDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

data "aws_ecs_cluster" "test" {
  cluster_name = aws_ecs_cluster.test.name
}
`, rName)
}

func testAccClusterDataSourceConfig_containerInsights(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

data "aws_ecs_cluster" "test" {
  cluster_name = aws_ecs_cluster.test.name
}
`, rName)
}

func testAccClusterDataSourceConfig_tags(rName, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}

data "aws_ecs_cluster" "test" {
  cluster_name = aws_ecs_cluster.test.name
}
`, rName, tagKey, tagValue)
}
