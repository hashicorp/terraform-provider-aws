// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSDaemonDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecs_daemon.test"
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_arn", resourceName, "cluster_arn"),
				),
			},
		},
	})
}

func TestAccECSDaemonDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecs_daemon.test"
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_arn", resourceName, "cluster_arn"),
				),
			},
		},
	})
}

func TestAccECSDaemonDataSource_optionalFields(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecs_daemon.test"
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonDataSourceConfig_optionalFields(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_arn", resourceName, "cluster_arn"),
				),
			},
		},
	})
}

func testAccDaemonDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDaemonConfig_basic(rName),
		`
data "aws_ecs_daemon" "test" {
  arn = aws_ecs_daemon.test.arn
}
`)
}

func testAccDaemonDataSourceConfig_tags(rName string) string {
	return acctest.ConfigCompose(
		testAccDaemonConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
		`
data "aws_ecs_daemon" "test" {
  arn = aws_ecs_daemon.test.arn
}
`)
}

func testAccDaemonDataSourceConfig_optionalFields(rName string) string {
	return acctest.ConfigCompose(
		testAccDaemonConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ecs_daemon" "test" {
  name                    = %[1]q
  cluster_arn            = aws_ecs_cluster.test.arn
  daemon_task_definition  = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns  = [aws_ecs_capacity_provider.test.arn]
  enable_ecs_managed_tags = true
  enable_execute_command  = true
  propagate_tags          = "DAEMON"
}

data "aws_ecs_daemon" "test" {
  arn = aws_ecs_daemon.test.arn
}
`, rName))
}
