// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfterraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// The ARN format documentation (https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonmanagedstreamingforapachekafka.html#amazonmanagedstreamingforapachekafka-resources-for-iam-policies)
	// shows ARNs having a UUID component, but in testing there is an additional component.
	kafkaUUIDRegexPattern = verify.UUIDRegexPattern + `-\w+` // nosemgrep:ci.kafka-in-const-name,ci.kafka-in-var-name
)

func TestAccKafkaReplicator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Replicator/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kafka", regexache.MustCompile(`replicator/`+rName+`/`+kafkaUUIDRegexPattern+`$`)),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_group_offset_sync_mode", "LEGACY"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.starting_position.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topic_name_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Replicator/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkafka.ResourceReplicator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sourceCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	targetCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicatorConfig_basic(rName, sourceCluster, targetCluster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kafka", regexache.MustCompile(`replicator/`+rName+`/`+kafkaUUIDRegexPattern+`$`)),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_group_offset_sync_mode", "LEGACY"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.starting_position.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topic_name_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "1"),
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
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kafka", regexache.MustCompile(`replicator/`+rName+`/`+kafkaUUIDRegexPattern+`$`)),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_group_offset_sync_mode", "LEGACY"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_exclude.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.synchronise_consumer_group_offsets", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.detect_and_copy_new_consumer_groups", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.starting_position.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.starting_position.0.type", "EARLIEST"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topic_name_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topic_name_configuration.0.type", "IDENTICAL"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_exclude.#", "1"),
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

func TestAccKafkaReplicator_logDelivery(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sourceCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	targetCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicatorConfig_logDelivery(rName, sourceCluster, targetCluster, false),
				ExpectError: regexache.MustCompile(`Error: cannot specify log_group when CloudWatch Logs logging is disabled\n\s*cannot specify delivery_stream when Firehose logging is disabled\n\s*cannot specify bucket when S3 logging is disabled\n\s*cannot specify prefix when S3 logging is disabled`),
			},
			{
				Config: testAccReplicatorConfig_logDelivery(rName, sourceCluster, targetCluster, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.cloudwatch_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName, "log_delivery.0.replicator_log_delivery.0.cloudwatch_logs.0.log_group",
						"aws_cloudwatch_log_group.test", names.AttrName,
					),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.firehose.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.firehose.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName, "log_delivery.0.replicator_log_delivery.0.firehose.0.delivery_stream",
						"aws_kinesis_firehose_delivery_stream.test", names.AttrName,
					),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.s3.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(
						resourceName, "log_delivery.0.replicator_log_delivery.0.s3.0.bucket",
						"aws_s3_bucket.test", names.AttrBucket,
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicatorConfig_logDeliveryDisabled(rName, sourceCluster, targetCluster),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.cloudwatch_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.cloudwatch_logs.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.firehose.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.firehose.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.s3.0.enabled", acctest.CtFalse),
				),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sourceCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	targetCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicatorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1, sourceCluster, targetCluster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccReplicatorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2, sourceCluster, targetCluster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKafkaReplicator_consumerGroupOffsetSyncMode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sourceCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	targetCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicatorConfig_consumerGroupOffsetSyncMode(rName, sourceCluster, targetCluster, "LEGACY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_group_offset_sync_mode", "LEGACY"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topic_name_configuration.0.type", "IDENTICAL"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "3"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicatorConfig_consumerGroupOffsetSyncMode(rName, sourceCluster, targetCluster, "ENHANCED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_group_offset_sync_mode", "ENHANCED"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topic_name_configuration.0.type", "IDENTICAL"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "3"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckReplicatorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_replicator" {
				continue
			}

			_, err := tfkafka.FindReplicatorByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckReplicatorExists(ctx context.Context, t *testing.T, n string, v *kafka.DescribeReplicatorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

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
  kafka_version          = "3.8.x"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.source[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.source.id]

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
            "aws:SourceAccount" : data.aws_caller_identity.current.account_id
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
  kafka_version          = "3.8.x"
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

      topic_name_configuration {
        type = "IDENTICAL"
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

func testAccReplicatorConfig_logDeliveryBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = "%[1]s-log-delivery"
  force_destroy = true
}

resource "aws_s3_bucket" "test_firehose" {
  bucket        = "%[1]s-firehose"
  force_destroy = true
}

resource "aws_iam_role" "firehose_role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Action": "sts:AssumeRole",
     "Principal": {
       "Service": "firehose.amazonaws.com"
     },
     "Effect": "Allow",
     "Sid": ""
   }
 ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose_role.arn
    bucket_arn = aws_s3_bucket.test_firehose.arn
  }

  tags = {
    LogDeliveryEnabled = "placeholder"
  }

  lifecycle {
    ignore_changes = [
      # Ignore changes to LogDeliveryEnabled tag as API adds this tag when broker log delivery is enabled
      tags["LogDeliveryEnabled"],
    ]
  }
}
`, rName)
}

func testAccReplicatorConfig_logDelivery(rName, sourceCluster, targetCluster string, logDeliveryEnabled bool) string {
	return acctest.ConfigCompose(
		testAccReplicatorConfig_source(sourceCluster),
		testAccReplicatorConfig_target(targetCluster),
		testAccReplicatorConfig_logDeliveryBase(rName),
		fmt.Sprintf(`
resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = aws_iam_role.source.arn

  log_delivery {
    replicator_log_delivery {
      cloudwatch_logs {
        enabled   = %[2]t
        log_group = aws_cloudwatch_log_group.test.name
      }
      s3 {
        enabled = %[2]t
        bucket  = aws_s3_bucket.test.bucket
        prefix  = "test/"
      }
      firehose {
        enabled         = %[2]t
        delivery_stream = aws_kinesis_firehose_delivery_stream.test.name
      }
    }
  }

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
`, rName, logDeliveryEnabled))
}

func testAccReplicatorConfig_logDeliveryDisabled(rName, sourceCluster, targetCluster string) string {
	return acctest.ConfigCompose(
		testAccReplicatorConfig_source(sourceCluster),
		testAccReplicatorConfig_target(targetCluster),
		testAccReplicatorConfig_logDeliveryBase(rName),
		fmt.Sprintf(`
resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = aws_iam_role.source.arn

  log_delivery {
    replicator_log_delivery {
      cloudwatch_logs {
        enabled = false
      }
      s3 {
        enabled = false
      }
      firehose {
        enabled = false
      }
    }
  }

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
`, rName))
}

func testAccReplicatorConfig_consumerGroupOffsetSyncMode(rName, sourceCluster, targetCluster, syncMode string) string {
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
      topics_to_replicate = ["topic1", "topic2", "topic3"]

      topic_name_configuration {
        type = "IDENTICAL"
      }
    }

    consumer_group_replication {
      consumer_groups_to_replicate    = ["group1", "group2", "group3"]
      consumer_group_offset_sync_mode = %[4]q
    }
  }
}
`, rName, sourceCluster, targetCluster, syncMode))
}

// propertyTestIterations is the minimum number of randomized iterations each
// property-based test runs.
const propertyTestIterations = 200

// randKafkaString returns a non-empty pseudo-random string for round-trip generators. // nosemgrep:ci.kafka-in-func-name
func randKafkaString(prefix string) string { // nosemgrep:ci.kafka-in-func-name
	return fmt.Sprintf("%s-%d", prefix, rand.IntN(1_000_000))
}

// randKafkaARN returns a non-empty pseudo-random ARN-shaped string. // nosemgrep:ci.kafka-in-func-name
func randKafkaARN(prefix string) string { // nosemgrep:ci.kafka-in-func-name
	return fmt.Sprintf("arn:aws:secretsmanager:us-east-1:123456789012:secret:%s-%d", prefix, rand.IntN(1_000_000))
}

// TestReplicatorApacheKafkaClusterRoundTrip asserts that expanding an apache_kafka_cluster
// config and flattening it back yields the original map, over randomized inputs.
func TestReplicatorApacheKafkaClusterRoundTrip(t *testing.T) { // nosemgrep:ci.kafka-in-func-name
	t.Parallel()

	for i := range propertyTestIterations {
		tfMap := map[string]any{
			"apache_kafka_cluster_id": randKafkaString("cluster"),
			"bootstrap_broker_string": fmt.Sprintf("b-1.%s:9092,b-2:9092", randKafkaString("broker")),
		}

		got := tfkafka.FlattenApacheKafkaCluster(tfkafka.ExpandApacheKafkaCluster(tfMap))

		if !reflect.DeepEqual(tfMap, got) {
			t.Fatalf("iteration %d: apache_kafka_cluster round-trip mismatch:\n input: %#v\noutput: %#v", i, tfMap, got)
		}
	}
}

// TestReplicatorClientAuthenticationRoundTrip asserts that expanding a client_authentication
// config (mTLS only, SASL/SCRAM only, or both) and flattening it back preserves every
// configured value, over randomized inputs.
func TestReplicatorClientAuthenticationRoundTrip(t *testing.T) {
	t.Parallel()

	mechanisms := []string{
		string(awstypes.KafkaClusterSaslScramMechanismSha256),
		string(awstypes.KafkaClusterSaslScramMechanismSha512),
	}

	for i := range propertyTestIterations {
		tfMap := map[string]any{}

		// kind: 0 = mTLS only, 1 = SASL/SCRAM only, 2 = both.
		kind := rand.IntN(3)
		if kind == 0 || kind == 2 {
			tfMap["mtls"] = []any{map[string]any{
				"secret_arn": randKafkaARN("mtls"),
			}}
		}
		if kind == 1 || kind == 2 {
			tfMap["sasl_scram"] = []any{map[string]any{
				"mechanism":  mechanisms[rand.IntN(len(mechanisms))],
				"secret_arn": randKafkaARN("sasl"),
			}}
		}

		got := tfkafka.FlattenKafkaClusterClientAuthentication(tfkafka.ExpandKafkaClusterClientAuthentication(tfMap))

		if !reflect.DeepEqual(tfMap, got) {
			t.Fatalf("iteration %d: client_authentication round-trip mismatch:\n input: %#v\noutput: %#v", i, tfMap, got)
		}
	}
}

// TestReplicatorEncryptionInTransitRoundTrip asserts that expanding an encryption_in_transit
// config and flattening it back preserves the root_ca_certificate secret ARN, over randomized
// inputs. EncryptionType is always TLS and set internally, so it is not part of the round-trip.
func TestReplicatorEncryptionInTransitRoundTrip(t *testing.T) {
	t.Parallel()

	for i := range propertyTestIterations {
		// root_ca_certificate is a Secrets Manager ARN referencing the custom CA chain.
		tfMap := map[string]any{
			"root_ca_certificate": randKafkaARN("secret"),
		}

		got := tfkafka.FlattenKafkaClusterEncryptionInTransit(tfkafka.ExpandKafkaClusterEncryptionInTransit(tfMap))

		if !reflect.DeepEqual(tfMap, got) {
			t.Fatalf("iteration %d: encryption_in_transit round-trip mismatch:\n input: %#v\noutput: %#v", i, tfMap, got)
		}
	}
}

// TestReplicatorKafkaClusterKindMutualExclusivity asserts that the resource diff succeeds only
// when every kafka_cluster entry sets exactly one of amazon_msk_cluster / apache_kafka_cluster,
// and errors when any entry sets both or neither, over randomized combinations.
func TestReplicatorKafkaClusterKindMutualExclusivity(t *testing.T) { // nosemgrep:ci.kafka-in-func-name
	t.Parallel()

	ctx := context.Background()

	amazonBlock := func() any {
		return []any{map[string]any{
			"msk_cluster_arn": "arn:aws:kafka:us-east-1:123456789012:cluster/test/00000000-0000-0000-0000-000000000000-1",
		}}
	}
	apacheBlock := func() any {
		return []any{map[string]any{
			"apache_kafka_cluster_id": "on-prem-1",
			"bootstrap_broker_string": "b-1:9092,b-2:9092",
		}}
	}

	for i := range propertyTestIterations {
		clusters := make([]any, 2)
		validEntries := 0

		for j := range clusters {
			entry := map[string]any{}

			// kind: 0 = amazon only, 1 = apache only, 2 = both, 3 = neither.
			switch rand.IntN(4) {
			case 0:
				entry["amazon_msk_cluster"] = amazonBlock()
				validEntries++
			case 1:
				entry["apache_kafka_cluster"] = apacheBlock()
				validEntries++
			case 2:
				entry["amazon_msk_cluster"] = amazonBlock()
				entry["apache_kafka_cluster"] = apacheBlock()
			case 3:
			}

			clusters[j] = entry
		}

		wantValid := validEntries == len(clusters)

		rc := tfterraform.NewResourceConfigRaw(map[string]any{
			"kafka_cluster": clusters,
		})
		_, err := tfkafka.ResourceReplicator().Diff(ctx, nil, rc, nil)

		switch {
		case wantValid && err != nil:
			t.Fatalf("iteration %d: expected no validation error, got: %v\nconfig: %#v", i, err, clusters)
		case !wantValid && err == nil:
			t.Fatalf("iteration %d: expected a validation error, got none\nconfig: %#v", i, clusters)
		}
	}
}

// TestReplicatorKafkaClusterIdentifier asserts that kafkaClusterIdentifier returns the MSK ARN
// for an Amazon MSK description, the cluster ID for an Apache description, and nil when neither
// is set, without panicking, over randomized inputs.
func TestReplicatorKafkaClusterIdentifier(t *testing.T) { // nosemgrep:ci.kafka-in-func-name
	t.Parallel()

	for i := range propertyTestIterations {
		var (
			desc     awstypes.KafkaClusterDescription
			wantARN  *string
			mskARN   = randKafkaARN("cluster")
			apacheID = randKafkaString("on-prem")
		)

		// kind: 0 = amazon, 1 = apache, 2 = neither.
		switch rand.IntN(3) {
		case 0:
			desc.AmazonMskCluster = &awstypes.AmazonMskCluster{MskClusterArn: aws.String(mskARN)}
			wantARN = aws.String(mskARN)
		case 1:
			desc.ApacheKafkaCluster = &awstypes.ApacheKafkaCluster{
				ApacheKafkaClusterId:  aws.String(apacheID),
				BootstrapBrokerString: aws.String("b-1:9092"),
			}
			wantARN = aws.String(apacheID)
		case 2:
			wantARN = nil
		}

		got := tfkafka.KafkaClusterIdentifier(desc)

		if aws.ToString(got) != aws.ToString(wantARN) {
			t.Fatalf("iteration %d: identifier mismatch: got %v, want %v", i, aws.ToString(got), aws.ToString(wantARN))
		}
	}
}

// TestExpandKafkaCluster_amazonMSKCluster verifies that a kafka_cluster with an
// amazon_msk_cluster block expands to a KafkaCluster with AmazonMskCluster set and no
// ApacheKafkaCluster.
func TestExpandKafkaCluster_amazonMSKCluster(t *testing.T) { // nosemgrep:ci.kafka-in-func-name,ci.msk-in-func-name
	t.Parallel()

	const arn = "arn:aws:kafka:us-east-1:123456789012:cluster/test/00000000-0000-0000-0000-000000000000-1"
	tfMap := map[string]any{
		"amazon_msk_cluster": []any{map[string]any{
			"msk_cluster_arn": arn,
		}},
	}

	apiObject := tfkafka.ExpandKafkaCluster(tfMap)

	if apiObject.AmazonMskCluster == nil {
		t.Fatal("expected AmazonMskCluster to be non-nil")
	}
	if got := aws.ToString(apiObject.AmazonMskCluster.MskClusterArn); got != arn {
		t.Fatalf("MskClusterArn: got %q, want %q", got, arn)
	}
	if apiObject.ApacheKafkaCluster != nil {
		t.Fatal("expected ApacheKafkaCluster to be nil")
	}
}

// TestReplicatorSecretARNValidation verifies that a malformed secret_arn is rejected by the
// schema ARN validator.
func TestReplicatorSecretARNValidation(t *testing.T) {
	t.Parallel()

	s := tfkafka.ResourceReplicator().SchemaFunc()
	clientAuth := s["kafka_cluster"].Elem.(*schema.Resource).Schema["client_authentication"].Elem.(*schema.Resource).Schema
	secretARN := clientAuth["mtls"].Elem.(*schema.Resource).Schema["secret_arn"]

	if _, errs := secretARN.ValidateFunc("not-an-arn", "secret_arn"); len(errs) == 0 {
		t.Fatal("expected a validation error for a malformed secret_arn, got none")
	}

	validARN := "arn:aws:secretsmanager:us-east-1:123456789012:secret:test-1"
	if _, errs := secretARN.ValidateFunc(validARN, "secret_arn"); len(errs) != 0 {
		t.Fatalf("expected no validation error for a well-formed secret_arn, got: %v", errs)
	}
}

// TestReplicatorRootCaCertificateValidation verifies that root_ca_certificate is required and
// must be a well-formed ARN (Secrets Manager secret ARN holding the custom CA chain).
func TestReplicatorRootCaCertificateValidation(t *testing.T) {
	t.Parallel()

	s := tfkafka.ResourceReplicator().SchemaFunc()
	encryptionInTransit := s["kafka_cluster"].Elem.(*schema.Resource).Schema["encryption_in_transit"].Elem.(*schema.Resource).Schema
	rootCA := encryptionInTransit["root_ca_certificate"]

	if !rootCA.Required {
		t.Fatal("expected root_ca_certificate to be required")
	}

	if _, errs := rootCA.ValidateFunc("not-an-arn", "root_ca_certificate"); len(errs) == 0 {
		t.Fatal("expected a validation error for a malformed root_ca_certificate, got none")
	}

	validARN := "arn:aws:secretsmanager:us-east-1:123456789012:secret:test-ca"
	if _, errs := rootCA.ValidateFunc(validARN, "root_ca_certificate"); len(errs) != 0 {
		t.Fatalf("expected no validation error for a well-formed root_ca_certificate ARN, got: %v", errs)
	}
}

// Environment variables gating the self-managed Apache Kafka replicator acceptance test.
// The test provisions only the replicator and wires it to pre-existing, mutually-reachable
// infrastructure whose values are supplied here (the source is an MSK cluster impersonating
// a self-managed cluster, so its public-CA TLS is trusted without a custom root CA). See
// internal/service/kafka/testdata/replicator_self_managed for a config that provisions that
// infrastructure and emits these values as outputs.
const (
	envVarOnPremKafkaEnabled          = "MSK_ONPREM_KAFKA_ENABLED"
	envVarOnPremKafkaBootstrap        = "MSK_ONPREM_KAFKA_BOOTSTRAP_BROKERS"
	envVarOnPremKafkaClusterID        = "MSK_ONPREM_KAFKA_CLUSTER_ID"
	envVarOnPremKafkaSASLSCRAMSecr    = "MSK_ONPREM_KAFKA_SASL_SCRAM_SECRET_ARN"
	envVarOnPremKafkaTargetCluster    = "MSK_ONPREM_KAFKA_TARGET_CLUSTER_ARN"
	envVarOnPremKafkaSubnetIDs        = "MSK_ONPREM_KAFKA_SUBNET_IDS"
	envVarOnPremKafkaSecurityGroupIDs = "MSK_ONPREM_KAFKA_SECURITY_GROUP_IDS"
)

// TestAccKafkaReplicator_selfManagedSASLSCRAM exercises the self-managed source path
// end-to-end using SASL/SCRAM. The source is an MSK cluster impersonating a self-managed
// Apache Kafka cluster; the test provisions only the replicator (plus its execution role)
// and wires it to reachable infrastructure supplied via env vars. See
// testdata/replicator_self_managed for a setup config that produces these values.
func TestAccKafkaReplicator_selfManagedSASLSCRAM(t *testing.T) { // nosemgrep:ci.kafka-in-func-name
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.SkipIfEnvVarNotSet(t, envVarOnPremKafkaEnabled)
	bootstrap := acctest.SkipIfEnvVarNotSet(t, envVarOnPremKafkaBootstrap)
	clusterID := acctest.SkipIfEnvVarNotSet(t, envVarOnPremKafkaClusterID)
	secretARN := acctest.SkipIfEnvVarNotSet(t, envVarOnPremKafkaSASLSCRAMSecr)
	targetARN := acctest.SkipIfEnvVarNotSet(t, envVarOnPremKafkaTargetCluster)
	subnetIDs := acctest.SkipIfEnvVarNotSet(t, envVarOnPremKafkaSubnetIDs)
	securityGroupIDs := acctest.SkipIfEnvVarNotSet(t, envVarOnPremKafkaSecurityGroupIDs)

	var replicator kafka.DescribeReplicatorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicatorConfig_selfManagedSASLSCRAM(rName, clusterID, bootstrap, secretARN, targetARN, subnetIDs, securityGroupIDs),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.apache_kafka_cluster_id", clusterID),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.apache_kafka_cluster.0.bootstrap_broker_string", bootstrap),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.client_authentication.0.sasl_scram.0.mechanism", "SHA512"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.client_authentication.0.sasl_scram.0.secret_arn", secretARN),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_id", clusterID),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", targetARN),
				),
			},
		},
	})
}

// TestAccKafkaReplicator_amazonMskClusterOnly_noDiff verifies backward compatibility: an
// existing amazon_msk_cluster-only configuration produces no plan differences after these
// changes.
func TestAccKafkaReplicator_amazonMskClusterOnly_noDiff(t *testing.T) { // nosemgrep:ci.kafka-in-func-name,ci.msk-in-func-name
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	sourceCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	targetCluster := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_replicator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicatorConfig_basic(rName, sourceCluster, targetCluster),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.amazon_msk_cluster.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.amazon_msk_cluster.#", "1"),
				),
			},
			{
				Config: testAccReplicatorConfig_basic(rName, sourceCluster, targetCluster),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccReplicatorConfig_selfManagedSASLSCRAM(rName, clusterID, bootstrap, secretARN, targetARN, subnetIDs, securityGroupIDs string) string { // nosemgrep:ci.kafka-in-func-name
	hclList := func(csv string) string {
		var quoted []string
		for _, p := range strings.Split(csv, ",") {
			if p = strings.TrimSpace(p); p != "" {
				quoted = append(quoted, fmt.Sprintf("%q", p))
			}
		}
		return strings.Join(quoted, ", ")
	}

	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "kafka.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSMSKReplicatorExecutionRole"
}

resource "aws_iam_role_policy" "test" {
  name = "secrets-and-kms"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["secretsmanager:GetSecretValue"]
        Resource = [%[4]q]
      },
      {
        Effect   = "Allow"
        Action   = ["kms:Decrypt", "kms:DescribeKey"]
        Resource = ["*"]
        Condition = {
          StringLike = { "kms:ViaService" = "secretsmanager.*.amazonaws.com" }
        }
      }
    ]
  })
}

resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  service_execution_role_arn = aws_iam_role.test.arn

  # Source: an MSK cluster impersonating a self-managed Apache Kafka cluster. Referenced by
  # its real Kafka cluster.id + SASL/SCRAM bootstrap brokers. No encryption_in_transit is
  # needed because the source's server certificate is signed by a public CA.
  kafka_cluster {
    apache_kafka_cluster {
      apache_kafka_cluster_id = %[2]q
      bootstrap_broker_string = %[3]q
    }

    client_authentication {
      sasl_scram {
        mechanism  = "SHA512"
        secret_arn = %[4]q
      }
    }
  }

  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = %[5]q
    }

    vpc_config {
      subnet_ids          = [%[6]s]
      security_groups_ids = [%[7]s]
    }
  }

  replication_info_list {
    source_kafka_cluster_id  = %[2]q
    target_kafka_cluster_arn = %[5]q
    target_compression_type  = "NONE"

    topic_replication {
      topics_to_replicate = [".*"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [".*"]
    }
  }
}
`, rName, clusterID, bootstrap, secretARN, targetARN, hclList(subnetIDs), hclList(securityGroupIDs))
}
