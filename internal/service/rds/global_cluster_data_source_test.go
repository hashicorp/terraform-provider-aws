// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSGlobalClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_global_cluster.test"
	resourceName := "aws_rds_global_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDatabaseName, resourceName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDeletionProtection, resourceName, names.AttrDeletionProtection),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEndpoint, resourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_lifecycle_support", resourceName, "engine_lifecycle_support"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngineVersion, resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version_actual", resourceName, "engine_version_actual"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_identifier", resourceName, "global_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_members.#", resourceName, "global_cluster_members.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_resource_id", resourceName, "global_cluster_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStorageEncrypted, resourceName, names.AttrStorageEncrypted),
				),
			},
		},
	})
}

func TestAccRDSGlobalClusterDataSource_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_global_cluster.test"
	resourceName := "aws_rds_global_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterDataSourceConfig_tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_cluster_identifier", resourceName, "global_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
				),
			},
		},
	})
}

func testAccGlobalClusterDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "aurora-postgresql"
}

data "aws_rds_global_cluster" "test" {
  global_cluster_identifier = aws_rds_global_cluster.test.global_cluster_identifier
}
`, rName)
}

func testAccGlobalClusterDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "aurora-postgresql"

  tags = {
    Name  = %[1]q
    Key1  = "Value1"
    Key2  = "Value2"
  }
}

data "aws_rds_global_cluster" "test" {
  global_cluster_identifier = aws_rds_global_cluster.test.global_cluster_identifier
}
`, rName)
}
