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
	_ "github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccKafkaConnectConnector_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_custom_plugin.test"

	propertiesFileContent := "key.converter=hello\nvalue.converter=world"
	bootstrapServers := fmt.Sprintf("%s:9094,%s:9094", acctest.RandomDomainName(), acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorConfigBasic(rName, bootstrapServers, propertiesFileContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectorExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision"),
					resource.TestCheckResourceAttrSet(resourceName, "location.0.s3.0.bucket_arn"),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.0.file_key", rName),
					resource.TestCheckResourceAttr(resourceName, "location.0.s3.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "state", kafkaconnect.CustomPluginStateActive),
					resource.TestCheckResourceAttr(resourceName, "content_type", kafkaconnect.CustomPluginContentTypeJar),
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

func testAccCheckConnectorExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no MSK Connector ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConnectConn

		_, err := tfkafkaconnect.FindConnectorByARN(context.TODO(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccConnectorConfigBase(name string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.10.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
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
  "Statement": [
    {
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
    }
  ]
}
EOF
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}
`, name))
}

func testAccConnectorConfigBasic(name string, bootstrapServers string, workerConfigurationPropertiesFileContent string) string {
	return acctest.ConfigCompose(
		testAccCustomPluginConfigBasic(name),
		testAccWorkerConfigurationBasic(name, workerConfigurationPropertiesFileContent),
		testAccConnectorConfigBase(name),
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
        cpu_utilization_percentage = 50
      }

      scale_out_policy {
        cpu_utilization_percentage = 50
      }
    }
  }

  configuration = {}

  kafka_cluster {
    apache_kafka_cluster {
      bootstrap_servers = %[2]q

      vpc {
        security_group_ids = [aws_security_group.test.id]
        subnet_ids         = [aws_subnet.test.id]
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
}
`, name, bootstrapServers))
}
