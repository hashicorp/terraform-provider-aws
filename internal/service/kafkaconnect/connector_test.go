package kafkaconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafkaconnect "github.com/hashicorp/terraform-provider-aws/internal/service/kafkaconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccKafkaConnectConnector_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      testAccCheckConnectorDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "capacity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.max_worker_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.mcu_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.min_worker_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_in_policy.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "capacity.0.autoscaling.0.scale_in_policy.0.cpu_utilization_percentage"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_out_policy.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "capacity.0.autoscaling.0.scale_out_policy.0.cpu_utilization_percentage"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.connector.class", "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.tasks.max", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.topics", "t1"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.bootstrap_servers"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.subnets.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.0.authentication_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.0.encryption_type", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "kafkaconnect_version", "2.7.1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "plugin.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "plugin.*", map[string]string{
						"custom_plugin.#": "1",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "service_execution_role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "worker_configuration.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKafkaConnectConnector_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      testAccCheckConnectorDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfkafkaconnect.ResourceConnector(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaConnectConnector_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      testAccCheckConnectorDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_allAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "capacity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.max_worker_count", "6"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.mcu_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.min_worker_count", "4"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_in_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_in_policy.0.cpu_utilization_percentage", "25"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_out_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_out_policy.0.cpu_utilization_percentage", "75"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.connector.class", "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.tasks.max", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.topics", "t1"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.bootstrap_servers"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.subnets.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.0.authentication_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.0.encryption_type", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "kafkaconnect_version", "2.7.1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.0.enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.0.log_group"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.0.delivery_stream", ""),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "plugin.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "plugin.*", map[string]string{
						"custom_plugin.#": "1",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "service_execution_role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "worker_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "worker_configuration.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "worker_configuration.0.revision"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectorConfig_allAttributesCapacityUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "capacity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.0.mcu_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.0.worker_count", "4"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.connector.class", "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.tasks.max", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.topics", "t1"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.bootstrap_servers"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.subnets.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.0.authentication_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.0.encryption_type", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "kafkaconnect_version", "2.7.1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.0.enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.0.log_group"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.0.delivery_stream", ""),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "plugin.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "plugin.*", map[string]string{
						"custom_plugin.#": "1",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "service_execution_role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "worker_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "worker_configuration.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "worker_configuration.0.revision"),
				),
			},
		},
	})
}

func testAccCheckConnectorExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Connect Connector ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectConn

		_, err := tfkafkaconnect.FindConnectorByARN(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckConnectorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_mskconnect_connector" {
			continue
		}

		_, err := tfkafkaconnect.FindConnectorByARN(context.TODO(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("MSK Connect Connector %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccConnectorBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.10.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.10.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test3" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.10.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.s3"
  vpc_endpoint_type = "Interface"

  security_group_ids = [
    aws_security_group.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type        = "Service"
      identifiers = ["kafkaconnect.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "s3:*"
    ],
    "Resource": ["*"]
  },{
    "Effect": "Allow",
    "Action": [
      "logs:*"
    ],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.2.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.test1.id, aws_subnet.test2.id, aws_subnet.test3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.test.id]
  }
}
`, rName))
}

func testAccConnectorConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfig_basic(rName),
		testAccConnectorBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_mskconnect_connector" "test" {
  name = %[1]q

  kafkaconnect_version = "2.7.1"

  capacity {
    autoscaling {
      min_worker_count = 1
      max_worker_count = 2
    }
  }

  connector_configuration = {
    "connector.class" = "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"
    "tasks.max"       = "1"
    "topics"          = "t1"
  }

  kafka_cluster {
    apache_kafka_cluster {
      bootstrap_servers = aws_msk_cluster.test.bootstrap_brokers_tls

      vpc {
        security_groups = [aws_security_group.test.id]
        subnets         = [aws_subnet.test1.id, aws_subnet.test2.id, aws_subnet.test3.id]
      }
    }
  }

  kafka_cluster_client_authentication {
    authentication_type = "NONE"
  }

  kafka_cluster_encryption_in_transit {
    encryption_type = "TLS"
  }

  plugin {
    custom_plugin {
      arn      = aws_mskconnect_custom_plugin.test.arn
      revision = aws_mskconnect_custom_plugin.test.latest_revision
    }
  }

  service_execution_role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test, aws_vpc_endpoint.test]
}
`, rName))
}

func testAccConnectorConfig_allAttributes(rName string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfig_basic(rName),
		testAccWorkerConfigurationConfig_basic(rName),
		testAccConnectorBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_mskconnect_connector" "test" {
  name = %[1]q

  kafkaconnect_version = "2.7.1"

  capacity {
    autoscaling {
      mcu_count        = 2
      min_worker_count = 4
      max_worker_count = 6

      scale_in_policy {
        cpu_utilization_percentage = 25
      }

      scale_out_policy {
        cpu_utilization_percentage = 75
      }
    }
  }

  connector_configuration = {
    "connector.class" = "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"
    "tasks.max"       = "1"
    "topics"          = "t1"
  }

  kafka_cluster {
    apache_kafka_cluster {
      bootstrap_servers = aws_msk_cluster.test.bootstrap_brokers_tls

      vpc {
        security_groups = [aws_security_group.test.id]
        subnets         = [aws_subnet.test1.id, aws_subnet.test2.id, aws_subnet.test3.id]
      }
    }
  }

  kafka_cluster_client_authentication {
    authentication_type = "NONE"
  }

  kafka_cluster_encryption_in_transit {
    encryption_type = "TLS"
  }

  log_delivery {
    worker_log_delivery {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.test.name
      }

      firehose {
        enabled = false
      }

      s3 {
        enabled = false
      }
    }
  }

  plugin {
    custom_plugin {
      arn      = aws_mskconnect_custom_plugin.test.arn
      revision = aws_mskconnect_custom_plugin.test.latest_revision
    }
  }

  service_execution_role_arn = aws_iam_role.test.arn

  worker_configuration {
    arn      = aws_mskconnect_worker_configuration.test.arn
    revision = aws_mskconnect_worker_configuration.test.latest_revision
  }

  depends_on = [aws_iam_role_policy.test, aws_vpc_endpoint.test]
}
`, rName))
}

func testAccConnectorConfig_allAttributesCapacityUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfig_basic(rName),
		testAccWorkerConfigurationConfig_basic(rName),
		testAccConnectorBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_mskconnect_connector" "test" {
  name = %[1]q

  kafkaconnect_version = "2.7.1"

  capacity {
    provisioned_capacity {
      worker_count = 4
    }
  }

  connector_configuration = {
    "connector.class" = "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"
    "tasks.max"       = "1"
    "topics"          = "t1"
  }

  kafka_cluster {
    apache_kafka_cluster {
      bootstrap_servers = aws_msk_cluster.test.bootstrap_brokers_tls

      vpc {
        security_groups = [aws_security_group.test.id]
        subnets         = [aws_subnet.test1.id, aws_subnet.test2.id, aws_subnet.test3.id]
      }
    }
  }

  kafka_cluster_client_authentication {
    authentication_type = "NONE"
  }

  kafka_cluster_encryption_in_transit {
    encryption_type = "TLS"
  }

  log_delivery {
    worker_log_delivery {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.test.name
      }

      firehose {
        enabled = false
      }

      s3 {
        enabled = false
      }
    }
  }

  plugin {
    custom_plugin {
      arn      = aws_mskconnect_custom_plugin.test.arn
      revision = aws_mskconnect_custom_plugin.test.latest_revision
    }
  }

  service_execution_role_arn = aws_iam_role.test.arn

  worker_configuration {
    arn      = aws_mskconnect_worker_configuration.test.arn
    revision = aws_mskconnect_worker_configuration.test.latest_revision
  }

  depends_on = [aws_iam_role_policy.test, aws_vpc_endpoint.test]
}
`, rName))
}
