// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobDefinitionDataSource_basicName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_batch_job_definition.test"
	resourceName := "aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionDataSourceConfig_basicName(rName, acctest.Ct1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "retry_strategy.0.attempts", acctest.Ct10),
					resource.TestCheckResourceAttr(dataSourceName, "revision", acctest.Ct1),
				),
			},
			{
				Config: testAccJobDefinitionDataSourceConfig_basicNameRevision(rName, acctest.Ct2, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "revision", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccBatchJobDefinitionDataSource_basicARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionDataSourceConfig_basicARN(rName, acctest.Ct1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "retry_strategy.0.attempts", acctest.Ct10),
					resource.TestCheckResourceAttr(dataSourceName, "revision", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "revision", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "batch", regexache.MustCompile(fmt.Sprintf(`job-definition/%s:\d+`, rName))),
				),
			},
			{
				Config: testAccJobDefinitionDataSourceConfig_basicARN(rName, acctest.Ct2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "revision", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccBatchJobDefinitionDataSource_basicARN_NodeProperties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionDataSourceConfig_basicARNNode(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "node_properties.0.main_node", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "node_properties.0.node_range_properties.#", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, "node_properties.0.node_range_properties.0.container.0.image", "busybox"),
				),
			},
		},
	})
}

func TestAccBatchJobDefinitionDataSource_basicARN_EKSProperties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_batch_job_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BatchEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobDefinitionDataSourceConfig_basicARNEKS(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "eks_properties.0.pod_properties.0.containers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "eks_properties.0.pod_properties.0.containers.0.image", "public.ecr.aws/amazonlinux/amazonlinux:1"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrType, "container"),
				),
			},
		},
	})
}

func testAccJobDefinitionDataSourceConfig_basicARN(rName string, increment string) string {
	return acctest.ConfigCompose(
		testAccJobDefinitionDataSourceConfig_container(rName, increment),
		`
data "aws_batch_job_definition" "test" {
  arn = aws_batch_job_definition.test.arn
}
`)
}

func testAccJobDefinitionDataSourceConfig_basicName(rName string, increment string) string {
	return acctest.ConfigCompose(
		testAccJobDefinitionDataSourceConfig_container(rName, increment),
		fmt.Sprintf(`
data "aws_batch_job_definition" "test" {
  name = %[1]q

  depends_on = [aws_batch_job_definition.test]
}
`, rName, increment))
}

func testAccJobDefinitionDataSourceConfig_basicNameRevision(rName string, increment string, revision int) string {
	return acctest.ConfigCompose(
		testAccJobDefinitionDataSourceConfig_container(rName, increment),
		fmt.Sprintf(`
data "aws_batch_job_definition" "test" {
  name     = %[1]q
  revision = %[2]d

  depends_on = [aws_batch_job_definition.test]
}
`, rName, revision))
}

func testAccJobDefinitionDataSourceConfig_container(rName string, increment string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_definition" "test" {
  container_properties = jsonencode({
    command = ["echo", "test%[2]s"]
    image   = "busybox"
    memory  = 128
    vcpus   = 1
  })
  name = %[1]q
  type = "container"
  retry_strategy {
    attempts = 10
  }
}
`, rName, increment)
}

func testAccJobDefinitionDataSourceConfig_basicARNNode(rName string) string {
	return acctest.ConfigCompose(
		testAccJobDefinitionConfig_NodeProperties(rName), `
data "aws_batch_job_definition" "test" {
  arn = aws_batch_job_definition.test.arn
}`)
}

func testAccJobDefinitionDataSourceConfig_basicARNEKS(rName string) string {
	return acctest.ConfigCompose(
		testAccJobDefinitionConfig_EKSProperties_basic(rName), `
data "aws_batch_job_definition" "test" {
  arn = aws_batch_job_definition.test.arn
}
`)
}
