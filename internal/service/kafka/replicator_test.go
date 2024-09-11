// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaReplicator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceCluster := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	targetCluster := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicatorConfig_basic(rName, sourceCluster, targetCluster),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.subnet_ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.security_groups_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.subnet_ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.security_groups_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.starting_position.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", acctest.Ct1),
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

func TestAccKafkaReplicator_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceCluster := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	targetCluster := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicatorConfig_basic(rName, sourceCluster, targetCluster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.subnet_ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.security_groups_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.subnet_ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.security_groups_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.starting_position.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicatorConfig_update(rName, sourceCluster, targetCluster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.subnet_ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.security_groups_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.subnet_ids.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.security_groups_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_exclude.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.synchronise_consumer_group_offsets", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.detect_and_copy_new_consumer_groups", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.starting_position.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.starting_position.0.type", "EARLIEST"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_exclude.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.copy_topic_configurations", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.copy_access_control_lists_for_topics", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.detect_and_copy_new_topics", acctest.CtFalse),
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

func TestAccKafkaReplicator_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceCluster := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	targetCluster := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicatorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1, sourceCluster, targetCluster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
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
				Config: testAccReplicatorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2, sourceCluster, targetCluster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccReplicatorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2, sourceCluster, targetCluster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKafkaReplicator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceCluster := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	targetCluster := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicatorConfig_basic(rName, sourceCluster, targetCluster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceReplicator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReplicatorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_replicator" {
				continue
			}

			_, err := tfkafka.FindReplicatorByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK Replicator %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckReplicatorExists(ctx context.Context, n string, v *kafka.DescribeReplicatorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		output, err := tfkafka.FindReplicatorByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccReplicatorConfig_source(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_allowEveryoneNoACLFoundFalse(rName),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_msk_cluster" "source" {
  cluster_name           = %[1]q
  kafka_version          = "2.8.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.source[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.source.id]

    connectivity_info {
      vpc_connectivity {
        client_authentication {
          sasl {
            iam = true
          }
        }
      }
    }

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.test.arn
    revision = aws_msk_configuration.test.latest_revision
  }

  client_authentication {
    sasl {
      iam = true
    }
  }
}

resource "aws_msk_cluster_policy" "source" {
  cluster_arn = aws_msk_cluster.source.arn

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Sid    = "testMskClusterPolicy"
      Effect = "Allow"
      Principal = {
        "Service" = "kafka.amazonaws.com"
      }
      Action = [
        "kafka:CreateVpcConnection",
        "kafka:GetBootstrapBrokers",
        "kafka:DescribeCluster",
        "kafka:DescribeClusterV2"
      ]
      Resource = aws_msk_cluster.source.arn
    }]
  })

  depends_on = [aws_msk_cluster.source]
}

resource "aws_iam_role" "source" {
  name = "%[1]s"

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "kafka.amazonaws.com"
        },
        "Action" : "sts:AssumeRole",
        "Condition" : {
          "StringEquals" : {
            "aws:SourceAccount" : "${data.aws_caller_identity.current.account_id}"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "source" {
  name = %[1]q
  role = aws_iam_role.source.name

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Resource" : "*",
        "Action" : [
          "kafka-cluster:Connect",
          "kafka-cluster:DescribeCluster",
          "kafka-cluster:AlterCluster",
          "kafka-cluster:ReadData",
          "kafka-cluster:WriteData",
          "kafka-cluster:DescribeTopic",
          "kafka-cluster:CreateTopic",
          "kafka-cluster:AlterTopic",
          "kafka-cluster:AlterGroup",
          "kafka-cluster:DescribeGroup",
          "kafka-cluster:DescribeTopicDynamicConfiguration",
          "kafka-cluster:AlterTopicDynamicConfiguration"
        ]
      }
    ]
  })
}

resource "aws_security_group" "source" {
  name   = %[1]q
  vpc_id = aws_vpc.source.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = -1
    self      = true
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
resource "aws_vpc" "source" {
  cidr_block = "10.0.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "source" {
  count = 3

  vpc_id            = aws_vpc.source.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.source.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccReplicatorConfig_target(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_msk_cluster" "target" {
  cluster_name           = %[1]q
  kafka_version          = "2.8.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.target[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.target.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }
  configuration_info {
    arn      = aws_msk_configuration.test.arn
    revision = aws_msk_configuration.test.latest_revision
  }

  client_authentication {
    sasl {
      iam = true
    }
  }
}

resource "aws_security_group" "target" {
  name   = %[1]q
  vpc_id = aws_vpc.target.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = -1
    self      = true
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

resource "aws_vpc" "target" {
  cidr_block = "10.1.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "target" {
  count = 3

  vpc_id            = aws_vpc.target.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.target.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccReplicatorConfig_basic(rName, sourceCluster, targetCluster string) string {
	return acctest.ConfigCompose(
		testAccReplicatorConfig_source(sourceCluster),
		testAccReplicatorConfig_target(targetCluster),
		fmt.Sprintf(`
resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = aws_iam_role.source.arn

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.source.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.source[*].id
      security_groups_ids = [aws_security_group.source.id]
    }
  }

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.target.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.target[*].id
      security_groups_ids = [aws_security_group.target.id]
    }
  }

  replication_info_list {
    source_kafka_cluster_arn = aws_msk_cluster.source.arn
    target_kafka_cluster_arn = aws_msk_cluster.target.arn
    target_compression_type  = "NONE"


    topic_replication {
      topics_to_replicate = [".*"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [".*"]
    }
  }
}
`, rName, sourceCluster, targetCluster))
}

func testAccReplicatorConfig_update(rName, sourceCluster, targetCluster string) string {
	return acctest.ConfigCompose(
		testAccReplicatorConfig_source(sourceCluster),
		testAccReplicatorConfig_target(targetCluster),
		fmt.Sprintf(`
resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = aws_iam_role.source.arn

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.source.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.source[*].id
      security_groups_ids = [aws_security_group.source.id]
    }
  }

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.target.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.target[*].id
      security_groups_ids = [aws_security_group.target.id]
    }
  }

  replication_info_list {
    source_kafka_cluster_arn = aws_msk_cluster.source.arn
    target_kafka_cluster_arn = aws_msk_cluster.target.arn
    target_compression_type  = "NONE"

    topic_replication {
      detect_and_copy_new_topics           = false
      copy_access_control_lists_for_topics = false
      copy_topic_configurations            = false
      topics_to_replicate                  = ["topic1", "topic2", "topic3"]
      topics_to_exclude                    = ["topic-4"]

      starting_position {
        type = "EARLIEST"
      }
    }

    consumer_group_replication {
      synchronise_consumer_group_offsets  = false
      detect_and_copy_new_consumer_groups = false
      consumer_groups_to_replicate        = ["group1", "group2", "group3"]
      consumer_groups_to_exclude          = ["group-4"]
    }
  }
}
`, rName, sourceCluster, targetCluster))
}

func testAccReplicatorConfig_tags1(rName, tagKey1, tagValue1, sourceCluster, targetCluster string) string {
	return acctest.ConfigCompose(
		testAccReplicatorConfig_source(sourceCluster),
		testAccReplicatorConfig_target(targetCluster),
		fmt.Sprintf(`

resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = aws_iam_role.source.arn

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.source.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.source[*].id
      security_groups_ids = [aws_security_group.source.id]
    }
  }

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.target.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.target[*].id
      security_groups_ids = [aws_security_group.target.id]
    }
  }

  replication_info_list {
    source_kafka_cluster_arn = aws_msk_cluster.source.arn
    target_kafka_cluster_arn = aws_msk_cluster.target.arn
    target_compression_type  = "NONE"


    topic_replication {
      topics_to_replicate = [".*"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [".*"]
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1, sourceCluster, targetCluster))
}

func testAccReplicatorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2, sourceCluster, targetCluster string) string {
	return acctest.ConfigCompose(
		testAccReplicatorConfig_source(sourceCluster),
		testAccReplicatorConfig_target(targetCluster),
		fmt.Sprintf(`
resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = aws_iam_role.source.arn

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.source.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.source[*].id
      security_groups_ids = [aws_security_group.source.id]
    }
  }

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = aws_msk_cluster.target.arn
    }

    vpc_config {
      subnet_ids          = aws_subnet.target[*].id
      security_groups_ids = [aws_security_group.target.id]
    }
  }

  replication_info_list {
    source_kafka_cluster_arn = aws_msk_cluster.source.arn
    target_kafka_cluster_arn = aws_msk_cluster.target.arn
    target_compression_type  = "NONE"


    topic_replication {
      topics_to_replicate = [".*"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [".*"]
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2, sourceCluster, targetCluster))
}
