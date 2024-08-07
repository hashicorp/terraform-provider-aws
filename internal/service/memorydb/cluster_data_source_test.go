// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMemoryDBClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"
	dataSourceName := "data.aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "acl_name", resourceName, "acl_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAutoMinorVersionUpgrade, resourceName, names.AttrAutoMinorVersionUpgrade),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_endpoint.0.address", resourceName, "cluster_endpoint.0.address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_endpoint.0.port", resourceName, "cluster_endpoint.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_tiering", resourceName, "data_tiering"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_patch_version", resourceName, "engine_patch_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngineVersion, resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyARN, resourceName, names.AttrKMSKeyARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "maintenance_window", resourceName, "maintenance_window"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "num_replicas_per_shard", resourceName, "num_replicas_per_shard"),
					resource.TestCheckResourceAttrPair(dataSourceName, "num_shards", resourceName, "num_shards"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrParameterGroupName, resourceName, names.AttrParameterGroupName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttr(dataSourceName, "security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "security_group_ids.*", resourceName, "security_group_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "shards.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.name", resourceName, "shards.0.name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.num_nodes", resourceName, "shards.0.num_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.slots", resourceName, "shards.0.slots"),
					resource.TestCheckResourceAttr(dataSourceName, "shards.0.nodes.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.availability_zone", resourceName, "shards.0.nodes.0.availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.create_time", resourceName, "shards.0.nodes.0.create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.name", resourceName, "shards.0.nodes.0.name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.endpoint.0.address", resourceName, "shards.0.nodes.0.endpoint.0.address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.endpoint.0.port", resourceName, "shards.0.nodes.0.endpoint.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_retention_limit", resourceName, "snapshot_retention_limit"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_window", resourceName, "snapshot_window"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSNSTopicARN, resourceName, names.AttrSNSTopicARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_group_name", resourceName, "subnet_group_name"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Test", "test"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tls_enabled", resourceName, "tls_enabled"),
				),
			},
		},
	})
}

func testAccClusterDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
		testAccClusterConfigBaseUserAndACL(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id
}

resource "aws_kms_key" "test" {}

resource "aws_memorydb_cluster" "test" {
  acl_name                   = aws_memorydb_acl.test.id
  auto_minor_version_upgrade = false
  kms_key_arn                = aws_kms_key.test.arn
  name                       = %[1]q
  node_type                  = "db.t4g.small"
  num_shards                 = 2
  security_group_ids         = [aws_security_group.test.id]
  snapshot_retention_limit   = 7
  subnet_group_name          = aws_memorydb_subnet_group.test.id
  tls_enabled                = true

  tags = {
    Test = "test"
  }
}

data "aws_memorydb_cluster" "test" {
  name = aws_memorydb_cluster.test.name
}
`, rName),
	)
}
