package memorydb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMemoryDBClusterDataSource_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"
	dataSourceName := "data.aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "acl_name", resourceName, "acl_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_minor_version_upgrade", resourceName, "auto_minor_version_upgrade"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_endpoint.0.address", resourceName, "cluster_endpoint.0.address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_endpoint.0.port", resourceName, "cluster_endpoint.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_patch_version", resourceName, "engine_patch_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_arn", resourceName, "kms_key_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "maintenance_window", resourceName, "maintenance_window"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "num_replicas_per_shard", resourceName, "num_replicas_per_shard"),
					resource.TestCheckResourceAttrPair(dataSourceName, "num_shards", resourceName, "num_shards"),
					resource.TestCheckResourceAttrPair(dataSourceName, "parameter_group_name", resourceName, "parameter_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttr(dataSourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "security_group_ids.*", resourceName, "security_group_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "shards.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.name", resourceName, "shards.0.name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.num_nodes", resourceName, "shards.0.num_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.slots", resourceName, "shards.0.slots"),
					resource.TestCheckResourceAttr(dataSourceName, "shards.0.nodes.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.availability_zone", resourceName, "shards.0.nodes.0.availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.create_time", resourceName, "shards.0.nodes.0.create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.name", resourceName, "shards.0.nodes.0.name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.endpoint.0.address", resourceName, "shards.0.nodes.0.endpoint.0.address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shards.0.nodes.0.endpoint.0.port", resourceName, "shards.0.nodes.0.endpoint.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_retention_limit", resourceName, "snapshot_retention_limit"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_window", resourceName, "snapshot_window"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sns_topic_arn", resourceName, "sns_topic_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_group_name", resourceName, "subnet_group_name"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
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
