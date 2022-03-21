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
	resourceName := "aws_mskconnect_custom_plugin.test"
	bootstrapServers := fmt.Sprintf("%s:9094,%s:9094", acctest.RandomDomainName(), acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      testAccCheckConnectorDestroy,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfigBasic(rName, bootstrapServers),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectorExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "capacity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.max_worker_count", "10"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.mcu_count", "4"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.min_worker_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_in_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_in_policy.0.cpu_utilization_percentage", "45"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_out_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.autoscaling.0.scale_out_policy.0.cpu_utilization_percentage", "55"),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.provisioned_capacity.#", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "connector_configuration.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "connector_configuration.connector.class", "io.confluent.connect.s3.S3SinkConnector"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.bootstrap_servers", bootstrapServers),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.vpc.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_client_authentication.0.authentication_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster_encryption_in_transit.0.encryption_type", "PLAINTEXT"),
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

func testAccConnectorConfigBase(rName string) string {
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
      "kafka-cluster:Connect",
      "kafka-cluster:DescribeCluster",
      "kafka-cluster:ReadData",
      "kafka-cluster:WriteData",
      "kafka-cluster:CreateTopic",
      "kafka-cluster:DescribeTopic",
      "kafka-cluster:AlterGroup",
      "kafka-cluster:DescribeGroup"
    ],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccConnectorConfigBasic(rName, bootstrapServers string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfigBasic(rName),
		testAccWorkerConfigurationBasic(rName, "key.converter=hello\nvalue.converter=world"),
		testAccConnectorConfigBase(rName),
		fmt.Sprintf(`
resource "aws_mskconnect_connector" "test" {
  name = %[1]q

  kafkaconnect_version = "2.7.1"

  capacity {
    autoscaling {
      mcu_count        = 4
      min_worker_count = 1
      max_worker_count = 10

      scale_in_policy {
        cpu_utilization_percentage = 45
      }

      scale_out_policy {
        cpu_utilization_percentage = 55
      }
    }
  }

  connector_configuration = {
    "connector.class"      = "io.confluent.connect.s3.S3SinkConnector"
    "tasks.max"            = "2"
    "topics"               = "my-example-topic"
    "s3.region"            = aws_s3_bucket.test.region
    "s3.bucket.name"       = aws_s3_bucket.test.bucket
    "flush.size"           = "1"
    "storage.class"        = "io.confluent.connect.s3.storage.S3Storage"
    "format.class"         = "io.confluent.connect.s3.format.json.JsonFormat"
    "partitioner.class"    = "io.confluent.connect.storage.partitioner.DefaultPartitioner"
    "key.converter"        = "org.apache.kafka.connect.storage.StringConverter"
    "value.converter"      = "org.apache.kafka.connect.storage.StringConverter"
    "schema.compatibility" = "NONE"
  }

  kafka_cluster {
    apache_kafka_cluster {
      bootstrap_servers = %[2]q

      vpc {
        security_groups = [aws_security_group.test.id]
        subnets         = [aws_subnet.test1.id, aws_subnet.test2.id]
      }
    }
  }

  kafka_cluster_client_authentication {
    authentication_type = "NONE"
  }

  kafka_cluster_encryption_in_transit {
    encryption_type = "PLAINTEXT"
  }

  plugin {
    custom_plugin {
      arn      = aws_mskconnect_custom_plugin.test.arn
      revision = aws_mskconnect_custom_plugin.test.latest_revision
    }
  }

  service_execution_role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy.test]
}
`, rName, bootstrapServers))
}

func testAccConnectorConfigAllAttributes(rName, bootstrapServers string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfigBasic(rName),
		testAccConnectorConfigBase(rName),
		fmt.Sprintf(`
resource "aws_mskconnect_connector" "test" {
  name = %[1]q

  kafkaconnect_version = "4.0"

  capacity {
    autoscaling {
      mcu_count        = 4
      min_worker_count = 1
      max_worker_count = 10

      scale_in_policy {
        cpu_utilization_percentage = 55
      }

      scale_out_policy {
        cpu_utilization_percentage = 45
      }
    }
  }

  connector_configuration = {
    Name = %[1]q
  }

  kafka_cluster {
    apache_kafka_cluster {
      bootstrap_servers = %[2]q

      vpc {
        security_groups = [aws_security_group.test.id]
        subnets         = [aws_subnet.test.id]
      }
    }
  }

  client_authentication {
    authentication_type = "NONE"
  }

  encryption_in_transit {
    encryption_type = "PLAINTEXT"
  }

  custom_plugin {
    arn      = aws_mskconnect_custom_plugin.test.arn
    revision = aws_mskconnect_custom_plugin.test.latest_revision
  }

  service_execution_role_arn = aws_iam_role.test.arn

  log_delivery {
    worker_log_delivery {
      s3 {
        enabled = true
        bucket  = aws_s3_bucket.test.id
        prefix  = "connector/"
      }
    }
  }

  worker_configuration {
    arn      = aws_mskconnect_worker_configuration.test.arn
    revision = aws_mskconnect_worker_configuration.test.latest_revision
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, bootstrapServers))
}
