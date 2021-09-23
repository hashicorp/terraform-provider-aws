package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataElasticacheCluster_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"
	dataSourceName := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElastiCacheClusterConfigWithDataSource(rName),
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

func TestAccAWSDataElasticacheCluster_Engine_Redis_LogDeliveryConfigurations_Cloudwatch(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elasticache.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElastiCacheCluster_Engine_Redis_LogDeliveryConfigurations_Cloudwatch(rName, true),
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
func testAccAWSElastiCacheClusterConfigWithDataSource(rName string) string {
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

func testAccAWSElastiCacheCluster_Engine_Redis_LogDeliveryConfigurations_Cloudwatch(rName string, enableLogDelivery bool) string {
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

resource "aws_elasticache_cluster" "test" {
  cluster_id        = "%[1]s"
  engine            = "redis"
  node_type         = "cache.t3.micro"
  num_cache_nodes   = 1
  port              = 6379
  apply_immediately = true

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

data "aws_elasticache_cluster" "test" {
  cluster_id = aws_elasticache_cluster.test.cluster_id
}
`, rName, enableLogDelivery)
}
