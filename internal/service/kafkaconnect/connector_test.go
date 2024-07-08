// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafkaconnect "github.com/hashicorp/terraform-provider-aws/internal/service/kafkaconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaConnectConnector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "capacity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.max_worker_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.mcu_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.min_worker_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_in_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "capacity.0.autoscaling.0.scale_in_policy.0.cpu_utilization_percentage"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_out_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "capacity.0.autoscaling.0.scale_out_policy.0.cpu_utilization_percentage"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.connector.class", "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.tasks.max", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.topics", "t1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.bootstrap_servers"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.subnets.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.0.authentication_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.0.encryption_type", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "kafkaconnect_version", "2.7.1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "plugin.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "plugin.*", map[string]string{
						"custom_plugin.#": acctest.Ct1,
					}),
					resource.TestCheckResourceAttrSet(resourceName, "service_execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, "worker_configuration.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafkaconnect.ResourceConnector(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaConnectConnector_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_allAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "capacity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.max_worker_count", "6"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.mcu_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.min_worker_count", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_in_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_in_policy.0.cpu_utilization_percentage", "25"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_out_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_out_policy.0.cpu_utilization_percentage", "75"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.connector.class", "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.tasks.max", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.topics", "t1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.bootstrap_servers"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.subnets.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.0.authentication_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.0.encryption_type", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "kafkaconnect_version", "2.7.1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.0.log_group"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.0.delivery_stream", ""),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "plugin.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "plugin.*", map[string]string{
						"custom_plugin.#": acctest.Ct1,
					}),
					resource.TestCheckResourceAttrSet(resourceName, "service_execution_role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, "worker_configuration.#", acctest.Ct1),
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
					testAccCheckConnectorExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "capacity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.0.mcu_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.0.worker_count", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.connector.class", "com.github.jcustenborder.kafka.connect.simulator.SimulatorSinkConnector"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.tasks.max", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.topics", "t1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.bootstrap_servers"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.subnets.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.0.authentication_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.0.encryption_type", "TLS"),
					resource.TestCheckResourceAttr(resourceName, "kafkaconnect_version", "2.7.1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "log_delivery.0.worker_log_delivery.0.cloudwatch_logs.0.log_group"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.0.delivery_stream", ""),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.firehose.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.worker_log_delivery.0.s3.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "plugin.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "plugin.*", map[string]string{
						"custom_plugin.#": acctest.Ct1,
					}),
					resource.TestCheckResourceAttrSet(resourceName, "service_execution_role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, "worker_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "worker_configuration.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "worker_configuration.0.revision"),
				),
			},
		},
	})
}

func TestAccKafkaConnectConnector_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.KafkaConnectEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaConnectServiceID),
		CheckDestroy:             testAccCheckConnectorDestroy(ctx),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConnectorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConnectorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckConnectorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectClient(ctx)

		_, err := tfkafkaconnect.FindConnectorByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckConnectorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mskconnect_connector" {
				continue
			}

			_, err := tfkafkaconnect.FindConnectorByARN(ctx, conn, rs.Primary.ID)

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
}

func testAccConnectorConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 3), fmt.Sprintf(`
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
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.test.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }
}
`, rName))
}

func testAccConnectorConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfig_basic(rName),
		testAccConnectorConfig_base(rName),
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
        subnets         = aws_subnet.test[*].id
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

  tags = {
    key1 = "value1"
  }

  depends_on = [aws_iam_role_policy.test, aws_vpc_endpoint.test]
}
`, rName))
}

func testAccConnectorConfig_allAttributes(rName string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfig_basic(rName),
		testAccWorkerConfigurationConfig_basic(rName),
		testAccConnectorConfig_base(rName),
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
        subnets         = aws_subnet.test[*].id
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
		testAccConnectorConfig_base(rName),
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
        subnets         = aws_subnet.test[*].id
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

func testAccConnectorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfig_basic(rName),
		testAccConnectorConfig_base(rName),
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
        subnets         = aws_subnet.test[*].id
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

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy.test, aws_vpc_endpoint.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccConnectorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfig_basic(rName),
		testAccConnectorConfig_base(rName),
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
        subnets         = aws_subnet.test[*].id
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy.test, aws_vpc_endpoint.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
