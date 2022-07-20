package memorydb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMemoryDBSnapshotDataSource_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_snapshot.test"
	dataSourceName := "data.aws_memorydb_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.description", resourceName, "cluster_configuration.0.description"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.engine_version", resourceName, "cluster_configuration.0.engine_version"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.maintenance_window", resourceName, "cluster_configuration.0.maintenance_window"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.name", resourceName, "cluster_configuration.0.name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.node_type", resourceName, "cluster_configuration.0.node_type"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.num_shards", resourceName, "cluster_configuration.0.num_shards"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.parameter_group_name", resourceName, "cluster_configuration.0.parameter_group_name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.port", resourceName, "cluster_configuration.0.port"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.snapshot_retention_limit", resourceName, "cluster_configuration.0.snapshot_retention_limit"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.snapshot_window", resourceName, "cluster_configuration.0.snapshot_window"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.subnet_group_name", resourceName, "cluster_configuration.0.subnet_group_name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_configuration.0.vpc_id", resourceName, "cluster_configuration.0.vpc_id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "kms_key_arn", resourceName, "kms_key_arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "source", resourceName, "source"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Test", "test"),
				),
			},
		},
	})
}

func testAccSnapshotDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSnapshotConfigBase(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_memorydb_snapshot" "test" {
  cluster_name = aws_memorydb_cluster.test.name
  kms_key_arn  = aws_kms_key.test.arn
  name         = %[1]q

  tags = {
    Test = "test"
  }
}

data "aws_memorydb_snapshot" "test" {
  name = aws_memorydb_snapshot.test.name
}
`, rName),
	)
}
