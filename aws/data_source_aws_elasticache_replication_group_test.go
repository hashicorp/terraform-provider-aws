package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsElasticacheReplicationGroup_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "auth_token_enabled", "false"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "automatic_failover_enabled", resourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az_enabled", resourceName, "multi_az_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "member_clusters.#", resourceName, "member_clusters.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_cache_clusters", resourceName, "number_cache_clusters"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "primary_endpoint_address", resourceName, "primary_endpoint_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "reader_endpoint_address", resourceName, "reader_endpoint_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_description", resourceName, "replication_group_description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_id", resourceName, "replication_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_window", resourceName, "snapshot_window"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsElasticacheReplicationGroup_ClusterMode(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_ClusterMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "auth_token_enabled", "false"),
					resource.TestCheckResourceAttrPair(dataSourceName, "automatic_failover_enabled", resourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az_enabled", resourceName, "multi_az_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration_endpoint_address", resourceName, "configuration_endpoint_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_description", resourceName, "replication_group_description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_id", resourceName, "replication_group_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsElasticacheReplicationGroup_MultiAZ(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_MultiAZ(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "automatic_failover_enabled", resourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az_enabled", resourceName, "multi_az_enabled"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsElasticacheReplicationGroup_NonExistent(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsElasticacheReplicationGroupConfig_NonExistent,
				ExpectError: regexp.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func TestAccDataSourceAwsElasticacheReplicationGroup_Engine_Redis_LogDeliveryConfigurations_CloudWatch(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_Engine_Redis_LogDeliveryConfigurations_Cloudwatch(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configurations.0.destination_details.0.cloudwatch_logs.0.log_group", rName),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configurations.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configurations.0.log_format", "text"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configurations.0.log_type", "slow-log"),
				),
			},
		},
	})
}
func testAccDataSourceAwsElasticacheReplicationGroupConfig_basic(rName string) string {
	return testAccAvailableAZsNoOptInConfig() + fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  availability_zones            = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
  automatic_failover_enabled    = true
  snapshot_window               = "01:00-02:00"
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName)
}

func testAccDataSourceAwsElasticacheReplicationGroupConfig_ClusterMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  port                          = 6379
  automatic_failover_enabled    = true

  cluster_mode {
    replicas_per_node_group = 1
    num_node_groups         = 2
  }
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName)
}

func testAccDataSourceAwsElasticacheReplicationGroupConfig_MultiAZ(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  automatic_failover_enabled    = true
  multi_az_enabled              = true
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName)
}

func testAccDataSourceAwsElasticacheReplicationGroupConfig_Engine_Redis_LogDeliveryConfigurations_Cloudwatch(rName string, enableLogDelivery bool, enableClusterMode bool) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "p" {
  count = tobool("%[2]t") ? 1 : 0
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["${aws_cloudwatch_log_group.lg[0].arn}:log-stream:*"]
    principals {
      identifiers = ["delivery.logs.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "rp" {
  count           = tobool("%[2]t") ? 1 : 0
  policy_document = data.aws_iam_policy_document.p[0].json
  policy_name     = "%[1]s"
  depends_on = [
    aws_cloudwatch_log_group.lg[0]
  ]
}

resource "aws_cloudwatch_log_group" "lg" {
  count             = tobool("%[2]t") ? 1 : 0
  retention_in_days = 1
  name              = "%[1]s"
}
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "%[1]s"
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"
  parameter_group_name          = tobool("%[3]t") ? "default.redis6.x.cluster.on" : "default.redis6.x"
  automatic_failover_enabled    = tobool("%[3]t")

  dynamic "cluster_mode" {
    for_each = tobool("%[3]t") ? [""] : []
    content {
      num_node_groups         = 1
      replicas_per_node_group = 0
    }
  }

  dynamic "log_delivery_configurations" {
    for_each = tobool("%[2]t") ? [""] : []
    content {
      destination_details {
        cloudwatch_logs {
          log_group = aws_cloudwatch_log_group.lg[0].name
        }
      }
      destination_type = "cloudwatch-logs"
      log_format       = "text"
      log_type         = "slow-log"
    }
  }
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName, enableLogDelivery, enableClusterMode)
}
const testAccDataSourceAwsElasticacheReplicationGroupConfig_NonExistent = `
data "aws_elasticache_replication_group" "test" {
  replication_group_id = "tf-acc-test-nonexistent"
}
`
