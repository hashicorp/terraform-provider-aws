package elasticache_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticache"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccElastiCacheClusterDataSource_Data_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"
	dataSourceName := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", resourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_address", resourceName, "cluster_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration_endpoint", resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "num_cache_nodes", resourceName, "num_cache_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
				),
			},
		},
	})
}

func TestAccElastiCacheClusterDataSource_Outpost_Data_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"
	dataSourceName := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_Outpost_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "outpost_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "cache_nodes.0.outpost_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", resourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_address", resourceName, "cluster_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration_endpoint", resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "num_cache_nodes", resourceName, "num_cache_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
				),
			},
		},
	})
}

func TestAccElastiCacheClusterDataSource_Engine_Redis_LogDeliveryConfigurations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, elasticache.DestinationTypeKinesisFirehose, elasticache.LogFormatJson, true, elasticache.DestinationTypeCloudwatchLogs, elasticache.LogFormatText),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.log_type", "engine-log"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.log_format", "json"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.log_type", "slow-log"),
				),
			},
		},
	})
}

func testAccClusterDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  port            = 11211
}

data "aws_elasticache_cluster" "test" {
  cluster_id = aws_elasticache_cluster.test.cluster_id
}
`, rName)
}

func testAccClusterDataSourceConfig_Outpost_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_elasticache_subnet_group" "test" {
  name       = "my-cache-subnet"
  subnet_ids = [aws_subnet.test.id]
}

resource "aws_elasticache_cluster" "test" {
  cluster_id      	= %[1]q
  outpost_mode		= "single-outpost"
  preferred_outpost_arn = data.aws_outposts_outpost.test.arn
  engine          	= "memcached"
  node_type       	= "cache.r5.large"
  num_cache_nodes 	= 1
  port            	= 11211
  subnet_group_name	= aws_elasticache_subnet_group.test.name
}

data "aws_elasticache_cluster" "test" {
  cluster_id = aws_elasticache_cluster.test.cluster_id
}
`, rName)
}
