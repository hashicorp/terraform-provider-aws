// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"maps"
	"math/rand/v2"
	"reflect"
	"slices"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
	return fmt.Sprintf("arn:aws:secretsmanager:us-east-1:123456789012:secret:%s-%d", prefix, rand.IntN(1_000_000)) //lintignore:AWSAT003,AWSAT005
}

// randKafkaClusterARN returns a pseudo-random MSK cluster ARN. // nosemgrep:ci.kafka-in-func-name
func randKafkaClusterARN(prefix string) string { // nosemgrep:ci.kafka-in-func-name
	return fmt.Sprintf("arn:aws:kafka:us-east-1:123456789012:cluster/%s/%d-1", prefix, rand.IntN(1_000_000)) //lintignore:AWSAT003,AWSAT005
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

	mechanisms := enum.Slice(
		awstypes.KafkaClusterSaslScramMechanismSha256,
		awstypes.KafkaClusterSaslScramMechanismSha512,
	)

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
			"msk_cluster_arn": "arn:aws:kafka:us-east-1:123456789012:cluster/test/00000000-0000-0000-0000-000000000000-1", //lintignore:AWSAT003,AWSAT005
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
			desc       awstypes.KafkaClusterDescription
			wantARN    *string
			clusterARN = randKafkaARN("cluster")
			apacheID   = randKafkaString("on-prem")
		)

		// kind: 0 = amazon, 1 = apache, 2 = neither.
		switch rand.IntN(3) {
		case 0:
			desc.AmazonMskCluster = &awstypes.AmazonMskCluster{MskClusterArn: aws.String(clusterARN)}
			wantARN = aws.String(clusterARN)
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

// TestReplicatorUpdateReplicationInfoInput asserts that an UpdateReplicationInfo request
// identifies each side of the replication by the identifier its cluster kind uses -- an Amazon
// MSK cluster by ARN, a self-managed Apache Kafka cluster by ID -- and leaves the unused field
// of each pair unset, over randomized combinations of the two kinds. Sending an empty string
// instead of omitting the unused field leaves the API unable to resolve the cluster.
func TestReplicatorUpdateReplicationInfoInput(t *testing.T) {
	t.Parallel()

	const (
		currentVersion = "1"
		replicatorARN  = "arn:aws:kafka:us-east-1:123456789012:replicator/test/00000000-0000-0000-0000-000000000000-1" //lintignore:AWSAT003,AWSAT005
	)

	for i := range propertyTestIterations {
		var (
			tfMap         = map[string]any{"target_compression_type": "NONE"}
			wantSourceARN *string
			wantSourceID  *string
			wantTargetARN *string
			wantTargetID  *string
		)

		// kind: 0 = Amazon MSK, referenced by ARN; 1 = Apache Kafka, referenced by ID.
		if rand.IntN(2) == 0 {
			v := randKafkaClusterARN(names.AttrSource)
			tfMap["source_kafka_cluster_arn"], wantSourceARN = v, aws.String(v)
		} else {
			v := randKafkaString("on-prem-source")
			tfMap["source_kafka_cluster_id"], wantSourceID = v, aws.String(v)
		}

		if rand.IntN(2) == 0 {
			v := randKafkaClusterARN(names.AttrTarget)
			tfMap["target_kafka_cluster_arn"], wantTargetARN = v, aws.String(v)
		} else {
			v := randKafkaString("on-prem-target")
			tfMap["target_kafka_cluster_id"], wantTargetID = v, aws.String(v)
		}

		d := schema.TestResourceDataRaw(t, tfkafka.ResourceReplicator().SchemaFunc(), map[string]any{
			"replication_info_list": []any{tfMap},
		})
		d.SetId(replicatorARN)
		if err := d.Set("current_version", currentVersion); err != nil {
			t.Fatalf("iteration %d: setting current_version: %s", i, err)
		}

		input := tfkafka.NewUpdateReplicationInfoInput(d)

		if got := aws.ToString(input.ReplicatorArn); got != replicatorARN {
			t.Fatalf("iteration %d: ReplicatorArn: got %q, want %q", i, got, replicatorARN)
		}
		if got := aws.ToString(input.CurrentVersion); got != currentVersion {
			t.Fatalf("iteration %d: CurrentVersion: got %q, want %q", i, got, currentVersion)
		}

		for _, identifier := range []struct {
			name string
			got  *string
			want *string
		}{
			{"SourceKafkaClusterArn", input.SourceKafkaClusterArn, wantSourceARN},
			{"SourceKafkaClusterId", input.SourceKafkaClusterId, wantSourceID},
			{"TargetKafkaClusterArn", input.TargetKafkaClusterArn, wantTargetARN},
			{"TargetKafkaClusterId", input.TargetKafkaClusterId, wantTargetID},
		} {
			switch {
			case identifier.want == nil && identifier.got != nil:
				t.Fatalf("iteration %d: %s: got %q, want unset\nconfig: %#v", i, identifier.name, aws.ToString(identifier.got), tfMap)
			case identifier.want != nil && identifier.got == nil:
				t.Fatalf("iteration %d: %s: got unset, want %q\nconfig: %#v", i, identifier.name, aws.ToString(identifier.want), tfMap)
			case aws.ToString(identifier.got) != aws.ToString(identifier.want):
				t.Fatalf("iteration %d: %s: got %q, want %q\nconfig: %#v", i, identifier.name, aws.ToString(identifier.got), aws.ToString(identifier.want), tfMap)
			}
		}
	}
}

// TestReplicatorFlattenReplicationInfoDescription documents which replication_info_list
// identifier attribute Read produces for each side of a replication: an Amazon MSK cluster
// yields the ARN attribute, a self-managed Apache Kafka cluster yields the ID attribute, and
// neither is produced when DescribeReplicator returns no cluster details for that side.
func TestReplicatorFlattenReplicationInfoDescription(t *testing.T) {
	t.Parallel()

	const (
		sourceARN = "arn:aws:kafka:us-east-1:123456789012:cluster/source/00000000-0000-0000-0000-000000000000-1" //lintignore:AWSAT003,AWSAT005
		targetARN = "arn:aws:kafka:us-east-1:123456789012:cluster/target/00000000-0000-0000-0000-000000000000-2" //lintignore:AWSAT003,AWSAT005
		sourceID  = "on-prem-source"
		targetID  = "on-prem-target"
	)

	amazonCluster := func(arn string) *awstypes.KafkaClusterDescription {
		return &awstypes.KafkaClusterDescription{
			AmazonMskCluster: &awstypes.AmazonMskCluster{MskClusterArn: aws.String(arn)},
		}
	}
	apacheCluster := func(id string) *awstypes.KafkaClusterDescription {
		return &awstypes.KafkaClusterDescription{
			ApacheKafkaCluster: &awstypes.ApacheKafkaCluster{
				ApacheKafkaClusterId:  aws.String(id),
				BootstrapBrokerString: aws.String("b-1:9092,b-2:9092"),
			},
		}
	}
	// clusterWithoutDetails is a description that carries neither cluster kind, which is what
	// DescribeReplicator returns for a self-managed cluster if it omits ApacheKafkaCluster.
	clusterWithoutDetails := func() *awstypes.KafkaClusterDescription {
		return &awstypes.KafkaClusterDescription{KafkaClusterAlias: aws.String(names.AttrAlias)}
	}

	// want lists the identifier attributes expected in the flattened map. Any identifier
	// attribute absent from want must be absent from the map.
	testCases := map[string]struct {
		source *awstypes.KafkaClusterDescription
		target *awstypes.KafkaClusterDescription
		want   map[string]string
	}{
		"Amazon MSK source and target": {
			source: amazonCluster(sourceARN),
			target: amazonCluster(targetARN),
			want: map[string]string{
				"source_kafka_cluster_arn": sourceARN,
				"target_kafka_cluster_arn": targetARN,
			},
		},
		"Apache Kafka source, Amazon MSK target": {
			source: apacheCluster(sourceID),
			target: amazonCluster(targetARN),
			want: map[string]string{
				"source_kafka_cluster_id":  sourceID,
				"target_kafka_cluster_arn": targetARN,
			},
		},
		"Amazon MSK source, Apache Kafka target": {
			source: amazonCluster(sourceARN),
			target: apacheCluster(targetID),
			want: map[string]string{
				"source_kafka_cluster_arn": sourceARN,
				"target_kafka_cluster_id":  targetID,
			},
		},
		"Apache Kafka source and target": {
			source: apacheCluster(sourceID),
			target: apacheCluster(targetID),
			want: map[string]string{
				"source_kafka_cluster_id": sourceID,
				"target_kafka_cluster_id": targetID,
			},
		},
		"source details omitted": {
			source: clusterWithoutDetails(),
			target: amazonCluster(targetARN),
			want: map[string]string{
				"target_kafka_cluster_arn": targetARN,
			},
		},
		"neither side resolved to a cluster": {
			source: nil,
			target: nil,
			want:   map[string]string{},
		},
	}

	identifierAttributes := []string{
		"source_kafka_cluster_arn",
		"source_kafka_cluster_id",
		"target_kafka_cluster_arn",
		"target_kafka_cluster_id",
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			apiObject := awstypes.ReplicationInfoDescription{
				SourceKafkaClusterAlias: aws.String("source-alias"),
				TargetKafkaClusterAlias: aws.String("target-alias"),
			}

			tfMap := tfkafka.FlattenReplicationInfoDescription(apiObject, testCase.source, testCase.target)

			for _, attribute := range identifierAttributes {
				got, set := tfMap[attribute]
				want, expected := testCase.want[attribute]

				switch {
				case expected && !set:
					t.Errorf("%s: absent, want %q", attribute, want)
				case !expected && set:
					t.Errorf("%s: got %q, want absent", attribute, got)
				case expected && got != want:
					t.Errorf("%s: got %q, want %q", attribute, got, want)
				}
			}

			// The computed aliases are what Read matches against output.KafkaClusters to pick
			// the source and target descriptions, so they must always survive the flatten.
			for attribute, want := range map[string]string{
				"source_kafka_cluster_alias": "source-alias",
				"target_kafka_cluster_alias": "target-alias",
			} {
				if got := tfMap[attribute]; got != want {
					t.Errorf("%s: got %v, want %q", attribute, got, want)
				}
			}
		})
	}
}

// TestExpandKafkaCluster_amazonMSKCluster verifies that a kafka_cluster with an
// amazon_msk_cluster block expands to a KafkaCluster with AmazonMskCluster set and no
// ApacheKafkaCluster.
func TestExpandKafkaCluster_amazonMSKCluster(t *testing.T) { // nosemgrep:ci.kafka-in-func-name,ci.msk-in-func-name
	t.Parallel()

	const arn = "arn:aws:kafka:us-east-1:123456789012:cluster/test/00000000-0000-0000-0000-000000000000-1" //lintignore:AWSAT003,AWSAT005
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

	validARN := "arn:aws:secretsmanager:us-east-1:123456789012:secret:test-1" //lintignore:AWSAT003,AWSAT005
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

	validARN := "arn:aws:secretsmanager:us-east-1:123456789012:secret:test-ca" //lintignore:AWSAT003,AWSAT005
	if _, errs := rootCA.ValidateFunc(validARN, "root_ca_certificate"); len(errs) != 0 {
		t.Fatalf("expected no validation error for a well-formed root_ca_certificate ARN, got: %v", errs)
	}
}

// Plan-shape and request-shape tests for the update path.
//
// resourceReplicatorUpdate, and therefore UpdateReplicationInfo, is only reached for changes
// Terraform plans in place. These tests pin down which changes plan in place and which force
// replacement, so the update path cannot silently become unreachable, and cover the request
// bodies the update expanders build. None of them need AWS.

const (
	testReplicatorARN       = "arn:aws:kafka:us-east-1:123456789012:replicator/test/00000000-0000-0000-0000-000000000000-1" //lintignore:AWSAT003,AWSAT005
	testReplicatorSourceARN = "arn:aws:kafka:us-east-1:123456789012:cluster/source/00000000-0000-0000-0000-000000000000-1"  //lintignore:AWSAT003,AWSAT005
	testReplicatorTargetARN = "arn:aws:kafka:us-east-1:123456789012:cluster/target/00000000-0000-0000-0000-000000000000-2"  //lintignore:AWSAT003,AWSAT005
	testReplicatorRoleARN   = "arn:aws:iam::123456789012:role/test"                                                         //lintignore:AWSAT005
	testReplicatorSecretARN = "arn:aws:secretsmanager:us-east-1:123456789012:secret:scram-1"                                //lintignore:AWSAT003,AWSAT005
	testReplicatorClusterID = "on-prem-1"
	testReplicatorBootstrap = "b-1.example.com:9096,b-2.example.com:9096"
)

// testReplicatorSetApacheTarget rewrites a raw config to replicate from the Amazon MSK source
// to a self-managed Apache Kafka target, which is referenced by its Kafka cluster.id instead of
// an ARN.
func testReplicatorSetApacheTarget(t *testing.T, raw map[string]any) {
	t.Helper()

	clusters, ok := raw["kafka_cluster"].([]any)
	if !ok || len(clusters) < 2 {
		t.Fatal("kafka_cluster missing from raw config")
	}

	clusters[1] = map[string]any{
		"apache_kafka_cluster": []any{map[string]any{
			"apache_kafka_cluster_id": testReplicatorClusterID,
			"bootstrap_broker_string": testReplicatorBootstrap,
		}},
		"client_authentication": []any{map[string]any{
			"sasl_scram": []any{map[string]any{
				"mechanism":  string(awstypes.KafkaClusterSaslScramMechanismSha512),
				"secret_arn": testReplicatorSecretARN,
			}},
		}},
	}

	replicationInfo := testReplicatorReplicationInfo(t, raw)
	delete(replicationInfo, "target_kafka_cluster_arn")
	replicationInfo["target_kafka_cluster_id"] = testReplicatorClusterID
}

// testReplicatorRawConfig returns a minimal valid configuration: an Amazon MSK source and
// target, one topic, one consumer group. Callers mutate the result to express the change under
// test. Each call builds fresh maps so mutations cannot leak between test cases.
func testReplicatorRawConfig() map[string]any {
	return map[string]any{
		"replicator_name":            "test",
		"service_execution_role_arn": testReplicatorRoleARN,
		"kafka_cluster": []any{
			map[string]any{
				"amazon_msk_cluster": []any{map[string]any{"msk_cluster_arn": testReplicatorSourceARN}},
			},
			map[string]any{
				"amazon_msk_cluster": []any{map[string]any{"msk_cluster_arn": testReplicatorTargetARN}},
			},
		},
		"replication_info_list": []any{map[string]any{
			"source_kafka_cluster_arn": testReplicatorSourceARN,
			"target_kafka_cluster_arn": testReplicatorTargetARN,
			"target_compression_type":  "NONE",
			"topic_replication": []any{map[string]any{
				"topics_to_replicate": []any{"topic-a"},
			}},
			"consumer_group_replication": []any{map[string]any{
				"consumer_groups_to_replicate": []any{"group-a"},
			}},
		}},
	}
}

// testReplicatorReplicationInfo returns the single replication_info_list entry of a raw config
// so a test can modify it in place.
func testReplicatorReplicationInfo(t *testing.T, raw map[string]any) map[string]any {
	t.Helper()

	list, ok := raw["replication_info_list"].([]any)
	if !ok || len(list) == 0 {
		t.Fatal("replication_info_list missing from raw config")
	}

	entry, ok := list[0].(map[string]any)
	if !ok {
		t.Fatal("replication_info_list entry is not a map")
	}

	return entry
}

// testReplicatorNestedBlock returns the single nested block named key from a replication_info
// entry, for example topic_replication or consumer_group_replication.
func testReplicatorNestedBlock(t *testing.T, replicationInfo map[string]any, key string) map[string]any {
	t.Helper()

	list, ok := replicationInfo[key].([]any)
	if !ok || len(list) == 0 {
		t.Fatalf("%s missing from replication_info_list entry", key)
	}

	block, ok := list[0].(map[string]any)
	if !ok {
		t.Fatalf("%s entry is not a map", key)
	}

	return block
}

// testReplicatorDiff computes the plan diff from oldRaw as prior state to newRaw as the new
// configuration.
func testReplicatorDiff(t *testing.T, oldRaw, newRaw map[string]any) *tfterraform.InstanceDiff {
	t.Helper()

	resource := tfkafka.ResourceReplicator()

	d := schema.TestResourceDataRaw(t, resource.SchemaFunc(), oldRaw)
	d.SetId(testReplicatorARN)

	state := d.State()
	if state == nil {
		t.Fatal("prior state: got nil")
	}

	diff, err := resource.Diff(context.Background(), state, tfterraform.NewResourceConfigRaw(newRaw), nil)
	if err != nil {
		t.Fatalf("computing diff: %s", err)
	}

	return diff
}

// testReplicatorDiffCase describes one plan-shape case. prepare applies to both the prior state
// and the new configuration, so a case can hold an attribute constant while changing a sibling;
// change applies only to the new configuration.
type testReplicatorDiffCase struct {
	prepare func(*testing.T, map[string]any)
	change  func(*testing.T, map[string]any)
}

func (testCase testReplicatorDiffCase) diff(t *testing.T) *tfterraform.InstanceDiff {
	t.Helper()

	oldRaw, newRaw := testReplicatorRawConfig(), testReplicatorRawConfig()

	if testCase.prepare != nil {
		testCase.prepare(t, oldRaw)
		testCase.prepare(t, newRaw)
	}
	testCase.change(t, newRaw)

	return testReplicatorDiff(t, oldRaw, newRaw)
}

// testReplicatorSetVPCConfig returns a mutator that places vpc_config on the first
// kafka_cluster entry.
func testReplicatorSetVPCConfig(subnetID, securityGroupID string) func(*testing.T, map[string]any) {
	return func(t *testing.T, raw map[string]any) {
		t.Helper()

		clusters, ok := raw["kafka_cluster"].([]any)
		if !ok || len(clusters) == 0 {
			t.Fatal("kafka_cluster missing from raw config")
		}

		clusters[0].(map[string]any)[names.AttrVPCConfig] = []any{map[string]any{
			names.AttrSubnetIDs:   []any{subnetID},
			"security_groups_ids": []any{securityGroupID},
		}}
	}
}

// TestReplicatorDiffUpdatesInPlace asserts that the replication settings MSK can actually
// update plan as in-place changes. This is the premise the cluster identifier fix rests on:
// UpdateReplicationInfo is only called for changes planned in place, so if replication_info_list
// being ForceNew propagated to its nested attributes, every topic change would replace the
// replicator and the identifiers the request carries would never matter.
func TestReplicatorDiffUpdatesInPlace(t *testing.T) {
	t.Parallel()

	testCases := map[string]testReplicatorDiffCase{
		"topics_to_replicate": {
			change: func(t *testing.T, raw map[string]any) {
				block := testReplicatorNestedBlock(t, testReplicatorReplicationInfo(t, raw), "topic_replication")
				block["topics_to_replicate"] = []any{"topic-a", "topic-b"}
			},
		},
		"topics_to_exclude": {
			change: func(t *testing.T, raw map[string]any) {
				block := testReplicatorNestedBlock(t, testReplicatorReplicationInfo(t, raw), "topic_replication")
				block["topics_to_exclude"] = []any{"topic-z"}
			},
		},
		"detect_and_copy_new_topics": {
			change: func(t *testing.T, raw map[string]any) {
				block := testReplicatorNestedBlock(t, testReplicatorReplicationInfo(t, raw), "topic_replication")
				block["detect_and_copy_new_topics"] = false
			},
		},
		"consumer_groups_to_replicate": {
			change: func(t *testing.T, raw map[string]any) {
				block := testReplicatorNestedBlock(t, testReplicatorReplicationInfo(t, raw), "consumer_group_replication")
				block["consumer_groups_to_replicate"] = []any{"group-a", "group-b"}
			},
		},
		"log_delivery": {
			change: func(t *testing.T, raw map[string]any) {
				raw["log_delivery"] = []any{map[string]any{
					"replicator_log_delivery": []any{map[string]any{
						"s3": []any{map[string]any{
							names.AttrEnabled: true,
							names.AttrBucket:  "test",
						}},
					}},
				}}
			},
		},
		// The same change against a self-managed Apache Kafka target, which the replication
		// references by cluster ID. The update then has to carry TargetKafkaClusterId, so this
		// confirms that topology reaches Update rather than replacing the replicator.
		"topics_to_replicate with an Apache Kafka target": {
			prepare: testReplicatorSetApacheTarget,
			change: func(t *testing.T, raw map[string]any) {
				block := testReplicatorNestedBlock(t, testReplicatorReplicationInfo(t, raw), "topic_replication")
				block["topics_to_replicate"] = []any{"topic-a", "topic-b"}
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diff := testCase.diff(t)

			switch {
			case diff == nil || diff.Empty():
				t.Fatal("expected a diff, got none: the change would not reach Update at all")
			case diff.RequiresNew():
				t.Fatal("expected an in-place update, got a replacement: UpdateReplicationInfo is never called for a replacement")
			}
		})
	}
}

// TestReplicatorDiffForcesReplacement asserts that the settings MSK cannot update force a
// replacement instead of reaching Update with a request the API would reject.
func TestReplicatorDiffForcesReplacement(t *testing.T) {
	t.Parallel()

	testCases := map[string]testReplicatorDiffCase{
		// Switching a side from an Amazon MSK cluster to a self-managed Apache Kafka cluster
		// changes which identifier the replication uses, which the API cannot do in place.
		"source cluster referenced by ID instead of ARN": {
			change: func(t *testing.T, raw map[string]any) {
				replicationInfo := testReplicatorReplicationInfo(t, raw)
				delete(replicationInfo, "source_kafka_cluster_arn")
				replicationInfo["source_kafka_cluster_id"] = testReplicatorClusterID
			},
		},
		"target_compression_type": {
			change: func(t *testing.T, raw map[string]any) {
				testReplicatorReplicationInfo(t, raw)["target_compression_type"] = "GZIP"
			},
		},
		// subnet_ids declares ForceNew, unlike its sibling security_groups_ids.
		"vpc_config subnet_ids": {
			prepare: testReplicatorSetVPCConfig("subnet-12345678", "sg-12345678"),
			change:  testReplicatorSetVPCConfig("subnet-87654321", "sg-12345678"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diff := testCase.diff(t)

			switch {
			case diff == nil || diff.Empty():
				t.Fatal("expected a diff, got none")
			case !diff.RequiresNew():
				t.Fatal("expected a replacement, got an in-place update: Update cannot apply this change")
			}
		})
	}
}

// TestReplicatorDiffAttributesTheAPICannotUpdate documents a known gap rather than asserting
// desirable behavior. UpdateReplicationInfo is the only replicator update operation, and it
// carries no replicator name, description, execution role, Kafka cluster, or VPC configuration.
// All of those nevertheless plan as in-place updates, so changing one issues a request that
// cannot apply it; Read then restores the old value from the API and the diff never converges.
//
// For the kafka_cluster attributes the cause is that ForceNew on a parent list does not reach
// its nested attributes: kafka_cluster declares ForceNew, but amazon_msk_cluster.msk_cluster_arn
// and vpc_config.security_groups_ids do not, so only their siblings that declare it explicitly
// (apache_kafka_cluster, client_authentication, encryption_in_transit, subnet_ids) replace the
// replicator. Note that reconcileKafkaClusters assumes the opposite.
//
// Marking these ForceNew is tracked separately. This test pins current behavior so that change
// is deliberate and visible, and it skips rather than fails once an attribute is fixed.
func TestReplicatorDiffAttributesTheAPICannotUpdate(t *testing.T) {
	t.Parallel()

	testCases := map[string]testReplicatorDiffCase{
		names.AttrDescription: {
			change: func(t *testing.T, raw map[string]any) {
				raw[names.AttrDescription] = "updated"
			},
		},
		"replicator_name": {
			change: func(t *testing.T, raw map[string]any) {
				raw["replicator_name"] = "updated"
			},
		},
		"service_execution_role_arn": {
			change: func(t *testing.T, raw map[string]any) {
				raw["service_execution_role_arn"] = "arn:aws:iam::123456789012:role/updated" //lintignore:AWSAT005
			},
		},
		"kafka_cluster msk_cluster_arn": {
			change: func(t *testing.T, raw map[string]any) {
				clusters, ok := raw["kafka_cluster"].([]any)
				if !ok || len(clusters) == 0 {
					t.Fatal("kafka_cluster missing from raw config")
				}
				clusters[0].(map[string]any)["amazon_msk_cluster"] = []any{map[string]any{
					"msk_cluster_arn": testReplicatorTargetARN,
				}}
			},
		},
		// Only the security group changes; subnet_ids is held constant because it declares
		// ForceNew and would otherwise replace the replicator.
		"vpc_config security_groups_ids": {
			prepare: testReplicatorSetVPCConfig("subnet-12345678", "sg-12345678"),
			change:  testReplicatorSetVPCConfig("subnet-12345678", "sg-87654321"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diff := testCase.diff(t)

			if diff == nil || diff.Empty() {
				t.Fatal("expected a diff, got none")
			}

			if diff.RequiresNew() {
				t.Skip("attribute now forces replacement: the API-cannot-update gap is fixed, so this test can be replaced with a ForceNew assertion")
			}
		})
	}
}

// TestReplicatorClusterIdentifierExactlyOneOf asserts the schema guarantee that
// newUpdateReplicationInfoInput depends on: exactly one of the ARN/ID pair is set for each side
// of the replication. Were both accepted, an update would send an ARN and an ID together for
// the same cluster.
func TestReplicatorClusterIdentifierExactlyOneOf(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		identifiers map[string]any
		valid       bool
	}{
		"ARN for both sides": {
			identifiers: map[string]any{
				"source_kafka_cluster_arn": testReplicatorSourceARN,
				"target_kafka_cluster_arn": testReplicatorTargetARN,
			},
			valid: true,
		},
		"ID for the source, ARN for the target": {
			identifiers: map[string]any{
				"source_kafka_cluster_id":  testReplicatorClusterID,
				"target_kafka_cluster_arn": testReplicatorTargetARN,
			},
			valid: true,
		},
		"ID for both sides": {
			identifiers: map[string]any{
				"source_kafka_cluster_id": testReplicatorClusterID,
				"target_kafka_cluster_id": "on-prem-2",
			},
			valid: true,
		},
		"both ARN and ID for the source": {
			identifiers: map[string]any{
				"source_kafka_cluster_arn": testReplicatorSourceARN,
				"source_kafka_cluster_id":  testReplicatorClusterID,
				"target_kafka_cluster_arn": testReplicatorTargetARN,
			},
			valid: false,
		},
		"both ARN and ID for the target": {
			identifiers: map[string]any{
				"source_kafka_cluster_arn": testReplicatorSourceARN,
				"target_kafka_cluster_arn": testReplicatorTargetARN,
				"target_kafka_cluster_id":  "on-prem-2",
			},
			valid: false,
		},
		"neither ARN nor ID for the source": {
			identifiers: map[string]any{
				"target_kafka_cluster_arn": testReplicatorTargetARN,
			},
			valid: false,
		},
		"neither ARN nor ID for the target": {
			identifiers: map[string]any{
				"source_kafka_cluster_arn": testReplicatorSourceARN,
			},
			valid: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			raw := testReplicatorRawConfig()
			replicationInfo := testReplicatorReplicationInfo(t, raw)
			for _, key := range []string{
				"source_kafka_cluster_arn",
				"source_kafka_cluster_id",
				"target_kafka_cluster_arn",
				"target_kafka_cluster_id",
			} {
				delete(replicationInfo, key)
			}
			maps.Copy(replicationInfo, testCase.identifiers)

			diags := tfkafka.ResourceReplicator().Validate(tfterraform.NewResourceConfigRaw(raw))

			if got := diags.HasError(); got != !testCase.valid {
				t.Fatalf("validation error = %t, want %t: %v", got, !testCase.valid, diags)
			}
		})
	}
}

// TestReplicatorFlattenKafkaClusterDescriptionsImport asserts the kafka_cluster state built
// straight from a DescribeReplicator response, which is what import produces: reconcile only
// runs when there is prior state, so on import the flattened response is written to state as is.
//
// DescribeReplicator echoes the same VpcConfig onto every entry, including self-managed Apache
// Kafka entries, where no configuration can express it -- validateKafkaClusterKind rejects
// vpc_config on an apache_kafka_cluster. If it reaches state, the following plan proposes
// removing it, which is an in-place change UpdateReplicationInfo cannot apply, so the diff never
// converges. The flatten therefore has to surface vpc_config only on the Amazon MSK entry that
// configures it.
//
// The response shape here mirrors a live MSK-to-Apache replicator: the Apache entry carries its
// cluster details, client authentication, and encryption, and both entries carry the same VPC.
func TestReplicatorFlattenKafkaClusterDescriptionsImport(t *testing.T) { // nosemgrep:ci.kafka-in-func-name
	t.Parallel()

	vpcConfig := func() *awstypes.KafkaClusterClientVpcConfig {
		return &awstypes.KafkaClusterClientVpcConfig{
			SecurityGroupIds: []string{"sg-12345678"},
			SubnetIds:        []string{"subnet-12345678"},
		}
	}

	apiObjects := []awstypes.KafkaClusterDescription{
		{
			KafkaClusterAlias: aws.String("msk-alias"),
			AmazonMskCluster:  &awstypes.AmazonMskCluster{MskClusterArn: aws.String(testReplicatorSourceARN)},
			VpcConfig:         vpcConfig(),
		},
		{
			KafkaClusterAlias: aws.String("apache-alias"),
			ApacheKafkaCluster: &awstypes.ApacheKafkaCluster{
				ApacheKafkaClusterId:  aws.String(testReplicatorClusterID),
				BootstrapBrokerString: aws.String(testReplicatorBootstrap),
			},
			ClientAuthentication: &awstypes.KafkaClusterClientAuthentication{
				SaslScram: &awstypes.KafkaClusterSaslScramAuthentication{
					Mechanism: awstypes.KafkaClusterSaslScramMechanismSha512,
					SecretArn: aws.String(testReplicatorSecretARN),
				},
			},
			EncryptionInTransit: &awstypes.KafkaClusterEncryptionInTransit{
				EncryptionType:    awstypes.KafkaClusterEncryptionInTransitTypeTls,
				RootCaCertificate: aws.String(testReplicatorSecretARN),
			},
			// The echo: the API reports the MSK entry's VPC on the Apache entry too.
			VpcConfig: vpcConfig(),
		},
	}

	tfList := tfkafka.FlattenKafkaClusterDescriptions(apiObjects)

	if got, want := len(tfList), 2; got != want {
		t.Fatalf("kafka_cluster entries: got %d, want %d", got, want)
	}

	mskEntry, ok := tfList[0].(map[string]any)
	if !ok {
		t.Fatal("kafka_cluster.0 is not a map")
	}
	if _, set := mskEntry[names.AttrVPCConfig]; !set {
		t.Error("kafka_cluster.0 (Amazon MSK): vpc_config absent, want the configured VPC")
	}

	apacheEntry, ok := tfList[1].(map[string]any)
	if !ok {
		t.Fatal("kafka_cluster.1 is not a map")
	}
	if v, set := apacheEntry[names.AttrVPCConfig]; set {
		t.Errorf("kafka_cluster.1 (Apache Kafka): vpc_config = %v, want absent -- no configuration can set it, so importing it leaves a diff that cannot be applied", v)
	}

	// The same response proves DescribeReplicator does return the Apache cluster details and
	// its client authentication, so import populates them without help from prior state.
	if _, set := apacheEntry["apache_kafka_cluster"]; !set {
		t.Error("kafka_cluster.1: apache_kafka_cluster absent, want the cluster details from the response")
	}
	if _, set := apacheEntry["client_authentication"]; !set {
		t.Error("kafka_cluster.1: client_authentication absent, want the SASL/SCRAM details from the response")
	}
}

// TestReplicatorDiffVPCConfigOnApacheEntry shows why the echoed vpc_config must never reach
// state. With it in state and absent from the configuration -- what import produced before the
// flatten dropped it -- the plan does not merely propose an unappliable update: vpc_config
// declares subnet_ids as ForceNew, so removing the block replaces the replicator. Importing a
// replicator with a self-managed cluster therefore set up a plan that destroyed and recreated
// it, interrupting replication, purely because state held a field no configuration can express.
func TestReplicatorDiffVPCConfigOnApacheEntry(t *testing.T) {
	t.Parallel()

	withApacheTarget := func(t *testing.T) map[string]any {
		t.Helper()

		raw := testReplicatorRawConfig()
		testReplicatorSetApacheTarget(t, raw)

		return raw
	}

	oldRaw := withApacheTarget(t)
	clusters, ok := oldRaw["kafka_cluster"].([]any)
	if !ok || len(clusters) < 2 {
		t.Fatal("kafka_cluster missing from raw config")
	}
	clusters[1].(map[string]any)[names.AttrVPCConfig] = []any{map[string]any{
		names.AttrSubnetIDs:   []any{"subnet-12345678"},
		"security_groups_ids": []any{"sg-12345678"},
	}}

	diff := testReplicatorDiff(t, oldRaw, withApacheTarget(t))

	switch {
	case diff == nil || diff.Empty():
		t.Fatal("expected a diff proposing to remove the vpc_config that no configuration can set")
	case !diff.RequiresNew():
		t.Fatal("expected removing vpc_config to force replacement, got an in-place update")
	}
}

// TestReplicatorExpandReplicationInfoApacheTarget asserts the CreateReplicator request for a
// replication running from an Amazon MSK source to a self-managed Apache Kafka target: the
// source carries an ARN, the target carries a cluster ID, and neither unused field is sent.
// This is the mirror of the source-side topology and the one the update fix also has to build,
// since the target identifier is chosen by the same logic.
func TestReplicatorExpandReplicationInfoApacheTarget(t *testing.T) {
	t.Parallel()

	tfMap := map[string]any{
		"source_kafka_cluster_arn": testReplicatorSourceARN,
		"target_kafka_cluster_id":  testReplicatorClusterID,
		"target_compression_type":  string(awstypes.TargetCompressionTypeNone),
		"topic_replication": []any{map[string]any{
			"topics_to_replicate": schema.NewSet(schema.HashString, []any{"topic-a"}),
		}},
		"consumer_group_replication": []any{map[string]any{
			"consumer_groups_to_replicate": schema.NewSet(schema.HashString, []any{"group-a"}),
		}},
	}

	apiObject := tfkafka.ExpandReplicationInfo(tfMap)

	if got, want := aws.ToString(apiObject.SourceKafkaClusterArn), testReplicatorSourceARN; got != want {
		t.Errorf("SourceKafkaClusterArn: got %q, want %q", got, want)
	}
	if apiObject.SourceKafkaClusterId != nil {
		t.Errorf("SourceKafkaClusterId: got %q, want unset", aws.ToString(apiObject.SourceKafkaClusterId))
	}
	if got, want := aws.ToString(apiObject.TargetKafkaClusterId), testReplicatorClusterID; got != want {
		t.Errorf("TargetKafkaClusterId: got %q, want %q", got, want)
	}
	if apiObject.TargetKafkaClusterArn != nil {
		t.Errorf("TargetKafkaClusterArn: got %q, want unset", aws.ToString(apiObject.TargetKafkaClusterArn))
	}
}

// TestReplicatorExpandKafkaClustersApacheTarget asserts the kafka_cluster list sent for the same
// topology: an Amazon MSK entry carrying the VPC the replicator uses to reach the Apache
// cluster, and an Apache entry carrying its bootstrap brokers and SASL/SCRAM credentials.
func TestReplicatorExpandKafkaClustersApacheTarget(t *testing.T) { // nosemgrep:ci.kafka-in-func-name
	t.Parallel()

	tfList := []any{
		map[string]any{
			"amazon_msk_cluster": []any{map[string]any{"msk_cluster_arn": testReplicatorSourceARN}},
			names.AttrVPCConfig: []any{map[string]any{
				names.AttrSubnetIDs:   schema.NewSet(schema.HashString, []any{"subnet-12345678"}),
				"security_groups_ids": schema.NewSet(schema.HashString, []any{"sg-12345678"}),
			}},
		},
		map[string]any{
			"apache_kafka_cluster": []any{map[string]any{
				"apache_kafka_cluster_id": testReplicatorClusterID,
				"bootstrap_broker_string": testReplicatorBootstrap,
			}},
			"client_authentication": []any{map[string]any{
				"sasl_scram": []any{map[string]any{
					"mechanism":  string(awstypes.KafkaClusterSaslScramMechanismSha512),
					"secret_arn": testReplicatorSecretARN,
				}},
			}},
		},
	}

	apiObjects := tfkafka.ExpandKafkaClusters(tfList)

	if got, want := len(apiObjects), 2; got != want {
		t.Fatalf("kafka clusters: got %d, want %d", got, want)
	}

	source := apiObjects[0]
	if source.AmazonMskCluster == nil {
		t.Fatal("source AmazonMskCluster: got nil, want the MSK cluster ARN")
	}
	if got, want := aws.ToString(source.AmazonMskCluster.MskClusterArn), testReplicatorSourceARN; got != want {
		t.Errorf("source MskClusterArn: got %q, want %q", got, want)
	}
	if source.VpcConfig == nil {
		t.Error("source VpcConfig: got nil, want the VPC the replicator reaches the Apache cluster through")
	}

	target := apiObjects[1]
	if target.ApacheKafkaCluster == nil {
		t.Fatal("target ApacheKafkaCluster: got nil, want the Apache cluster details")
	}
	if got, want := aws.ToString(target.ApacheKafkaCluster.ApacheKafkaClusterId), testReplicatorClusterID; got != want {
		t.Errorf("target ApacheKafkaClusterId: got %q, want %q", got, want)
	}
	if got, want := aws.ToString(target.ApacheKafkaCluster.BootstrapBrokerString), testReplicatorBootstrap; got != want {
		t.Errorf("target BootstrapBrokerString: got %q, want %q", got, want)
	}
	if target.ClientAuthentication == nil || target.ClientAuthentication.SaslScram == nil {
		t.Fatal("target SASL/SCRAM authentication: got nil, want the SCRAM secret")
	}
	if got, want := aws.ToString(target.ClientAuthentication.SaslScram.SecretArn), testReplicatorSecretARN; got != want {
		t.Errorf("target SCRAM SecretArn: got %q, want %q", got, want)
	}
	if target.AmazonMskCluster != nil {
		t.Error("target AmazonMskCluster: got a value, want nil")
	}
}

// TestReplicatorExpandTopicReplicationUpdate pins the request body the update path builds, and
// the one deliberate difference from the create path: the update expander forwards empty sets
// so that clearing topics_to_exclude reaches the API as "no exclusions" rather than being
// omitted, where an omitted field would leave the existing exclusions in place.
func TestReplicatorExpandTopicReplicationUpdate(t *testing.T) {
	t.Parallel()

	tfMap := map[string]any{
		"copy_access_control_lists_for_topics": true,
		"copy_topic_configurations":            true,
		"detect_and_copy_new_topics":           false,
		"topics_to_replicate":                  schema.NewSet(schema.HashString, []any{"topic-a", "topic-b"}),
		"topics_to_exclude":                    schema.NewSet(schema.HashString, []any{}),
	}

	apiObject := tfkafka.ExpandTopicReplicationUpdate(tfMap)

	if got, want := len(apiObject.TopicsToReplicate), 2; got != want {
		t.Fatalf("TopicsToReplicate: got %d entries, want %d", got, want)
	}
	if !slices.Contains(apiObject.TopicsToReplicate, "topic-a") || !slices.Contains(apiObject.TopicsToReplicate, "topic-b") {
		t.Fatalf("TopicsToReplicate: got %v, want topic-a and topic-b", apiObject.TopicsToReplicate)
	}
	if apiObject.TopicsToExclude == nil {
		t.Fatal("TopicsToExclude: got nil, want an empty list so the API clears existing exclusions")
	}
	if got := len(apiObject.TopicsToExclude); got != 0 {
		t.Fatalf("TopicsToExclude: got %d entries, want 0", got)
	}
	if !aws.ToBool(apiObject.CopyTopicConfigurations) {
		t.Error("CopyTopicConfigurations: got false, want true")
	}
	if aws.ToBool(apiObject.DetectAndCopyNewTopics) {
		t.Error("DetectAndCopyNewTopics: got true, want false")
	}

	// The create expander omits an empty set instead, leaving the field unset on a brand new
	// replication where there is nothing to clear.
	if created := tfkafka.ExpandTopicReplication(tfMap); created.TopicsToExclude != nil {
		t.Errorf("create path TopicsToExclude: got %v, want nil", created.TopicsToExclude)
	}
}

// TestReplicatorExpandConsumerGroupReplicationUpdate pins the consumer group request body,
// including that the update expander sends no ConsumerGroupOffsetSyncMode: the attribute is
// ForceNew, so a change to it replaces the replicator and never reaches this code.
func TestReplicatorExpandConsumerGroupReplicationUpdate(t *testing.T) {
	t.Parallel()

	tfMap := map[string]any{
		"consumer_group_offset_sync_mode":     string(awstypes.ConsumerGroupOffsetSyncModeLegacy),
		"consumer_groups_to_replicate":        schema.NewSet(schema.HashString, []any{"group-a"}),
		"consumer_groups_to_exclude":          schema.NewSet(schema.HashString, []any{}),
		"detect_and_copy_new_consumer_groups": true,
		"synchronise_consumer_group_offsets":  false,
	}

	apiObject := tfkafka.ExpandConsumerGroupReplicationUpdate(tfMap)

	if got, want := apiObject.ConsumerGroupsToReplicate, []string{"group-a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ConsumerGroupsToReplicate: got %v, want %v", got, want)
	}
	if apiObject.ConsumerGroupsToExclude == nil {
		t.Fatal("ConsumerGroupsToExclude: got nil, want an empty list so the API clears existing exclusions")
	}
	if !aws.ToBool(apiObject.DetectAndCopyNewConsumerGroups) {
		t.Error("DetectAndCopyNewConsumerGroups: got false, want true")
	}
	if aws.ToBool(apiObject.SynchroniseConsumerGroupOffsets) {
		t.Error("SynchroniseConsumerGroupOffsets: got true, want false")
	}

	// The create expander does carry the sync mode.
	if got, want := tfkafka.ExpandConsumerGroupReplication(tfMap).ConsumerGroupOffsetSyncMode, awstypes.ConsumerGroupOffsetSyncModeLegacy; got != want {
		t.Errorf("create path ConsumerGroupOffsetSyncMode: got %q, want %q", got, want)
	}
}

// Environment variables gating the self-managed Apache Kafka replicator acceptance test.
// The test provisions only the replicator and wires it to pre-existing, mutually-reachable
// infrastructure whose values are supplied here (the source is an MSK cluster impersonating
// a self-managed cluster, so its public-CA TLS is trusted without a custom root CA). See
// internal/service/kafka/testdata/replicator_self_managed for a config that provisions that
// infrastructure and emits these values as outputs.
const (
	envVarOnPremEnabled          = "MSK_ONPREM_KAFKA_ENABLED"
	envVarOnPremBootstrap        = "MSK_ONPREM_KAFKA_BOOTSTRAP_BROKERS"
	envVarOnPremClusterID        = "MSK_ONPREM_KAFKA_CLUSTER_ID"
	envVarOnPremSASLSCRAMSecr    = "MSK_ONPREM_KAFKA_SASL_SCRAM_SECRET_ARN"
	envVarOnPremTargetCluster    = "MSK_ONPREM_KAFKA_TARGET_CLUSTER_ARN"
	envVarOnPremSubnetIDs        = "MSK_ONPREM_KAFKA_SUBNET_IDS"
	envVarOnPremSecurityGroupIDs = "MSK_ONPREM_KAFKA_SECURITY_GROUP_IDS"
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

	acctest.SkipIfEnvVarNotSet(t, envVarOnPremEnabled)
	bootstrap := acctest.SkipIfEnvVarNotSet(t, envVarOnPremBootstrap)
	clusterID := acctest.SkipIfEnvVarNotSet(t, envVarOnPremClusterID)
	secretARN := acctest.SkipIfEnvVarNotSet(t, envVarOnPremSASLSCRAMSecr)
	targetARN := acctest.SkipIfEnvVarNotSet(t, envVarOnPremTargetCluster)
	subnetIDs := acctest.SkipIfEnvVarNotSet(t, envVarOnPremSubnetIDs)
	securityGroupIDs := acctest.SkipIfEnvVarNotSet(t, envVarOnPremSecurityGroupIDs)

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
				Config: testAccReplicatorConfig_selfManagedSASLSCRAM(rName, clusterID, bootstrap, secretARN, targetARN, subnetIDs, securityGroupIDs, ".*", ".*", false),
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
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "1"),
				),
			},
			{
				// Changing the topic list goes through UpdateReplicationInfo, which has to
				// identify the self-managed source by ID. An empty source ARN leaves the API
				// unable to resolve the cluster and the update fails.
				Config: testAccReplicatorConfig_selfManagedSASLSCRAM(rName, clusterID, bootstrap, secretARN, targetARN, subnetIDs, securityGroupIDs, "topic-a,topic-b", ".*", false),
				// Assert the change is applied in place. Without this the step would also pass
				// if the plan replaced the replicator, and a replacement never calls
				// UpdateReplicationInfo -- so the step would stop covering the update path.
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_id", clusterID),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", targetARN),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.*", "topic-a"),
					resource.TestCheckTypeSetElemAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.*", "topic-b"),
				),
			},
			{
				// consumer_group_replication travels in the same UpdateReplicationInfo request
				// as the topic settings and needs the same source cluster identifier, so it is
				// the second update shape that would have failed against a self-managed source.
				Config: testAccReplicatorConfig_selfManagedSASLSCRAM(rName, clusterID, bootstrap, secretARN, targetARN, subnetIDs, securityGroupIDs, "topic-a,topic-b", "group-a,group-b", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_id", clusterID),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.*", "group-a"),
					resource.TestCheckTypeSetElemAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.*", "group-b"),
				),
			},
			{
				// log_delivery is the third field UpdateReplicationInfo carries, and it is sent
				// alongside the same cluster identifiers.
				Config: testAccReplicatorConfig_selfManagedSASLSCRAM(rName, clusterID, bootstrap, secretARN, targetARN, subnetIDs, securityGroupIDs, "topic-a,topic-b", "group-a,group-b", true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_id", clusterID),
					resource.TestCheckResourceAttr(resourceName, "log_delivery.0.replicator_log_delivery.0.cloudwatch_logs.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "log_delivery.0.replicator_log_delivery.0.cloudwatch_logs.0.log_group", "aws_cloudwatch_log_group.test", names.AttrName),
				),
			},
		},
	})
}

// TestAccKafkaReplicator_selfManagedTarget exercises the opposite direction from
// TestAccKafkaReplicator_selfManagedSASLSCRAM: an Amazon MSK source replicating to a
// self-managed Apache Kafka target, which the replication references by cluster ID. It reuses
// the same infrastructure, so the env var names describe the source-direction roles -- the
// cluster the other test uses as its Apache source is the Apache target here, and its MSK
// target is the MSK source.
//
// The second step changes the topic list, which sends TargetKafkaClusterId through
// UpdateReplicationInfo. That is the target-side counterpart of the source-side bug, and the
// only place it is exercised against the real API.
func TestAccKafkaReplicator_selfManagedTarget(t *testing.T) { // nosemgrep:ci.kafka-in-func-name
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.SkipIfEnvVarNotSet(t, envVarOnPremEnabled)
	bootstrap := acctest.SkipIfEnvVarNotSet(t, envVarOnPremBootstrap)
	clusterID := acctest.SkipIfEnvVarNotSet(t, envVarOnPremClusterID)
	secretARN := acctest.SkipIfEnvVarNotSet(t, envVarOnPremSASLSCRAMSecr)
	mskClusterARN := acctest.SkipIfEnvVarNotSet(t, envVarOnPremTargetCluster)
	subnetIDs := acctest.SkipIfEnvVarNotSet(t, envVarOnPremSubnetIDs)
	securityGroupIDs := acctest.SkipIfEnvVarNotSet(t, envVarOnPremSecurityGroupIDs)

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
				Config: testAccReplicatorConfig_selfManagedTarget(rName, clusterID, bootstrap, secretARN, mskClusterARN, subnetIDs, securityGroupIDs, ".*"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.0.amazon_msk_cluster.0.msk_cluster_arn", mskClusterARN),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.apache_kafka_cluster.0.apache_kafka_cluster_id", clusterID),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.apache_kafka_cluster.0.bootstrap_broker_string", bootstrap),
					resource.TestCheckResourceAttr(resourceName, "kafka_cluster.1.client_authentication.0.sasl_scram.0.secret_arn", secretARN),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_arn", mskClusterARN),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_id", clusterID),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", ""),
				),
			},
			{
				Config: testAccReplicatorConfig_selfManagedTarget(rName, clusterID, bootstrap, secretARN, mskClusterARN, subnetIDs, securityGroupIDs, "topic-a,topic-b"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicatorExists(ctx, t, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_id", clusterID),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.*", "topic-a"),
					resource.TestCheckTypeSetElemAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.*", "topic-b"),
				),
			},
		},
	})
}

// TestAccKafkaReplicator_amazonMSKClusterOnly_noDiff verifies backward compatibility: an
// existing amazon_msk_cluster-only configuration produces no plan differences after these
// changes.
func TestAccKafkaReplicator_amazonMSKClusterOnly_noDiff(t *testing.T) { // nosemgrep:ci.kafka-in-func-name,ci.msk-in-func-name
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

// testAccReplicatorHCLList renders a comma-separated environment variable value as the elements
// of an HCL list.
func testAccReplicatorHCLList(csv string) string {
	var quoted []string
	for p := range strings.SplitSeq(csv, ",") {
		if p = strings.TrimSpace(p); p != "" {
			quoted = append(quoted, fmt.Sprintf("%q", p))
		}
	}
	return strings.Join(quoted, ", ")
}

func testAccReplicatorConfig_selfManagedSASLSCRAM(rName, clusterID, bootstrap, secretARN, targetARN, subnetIDs, securityGroupIDs, topics, consumerGroups string, logDelivery bool) string { // nosemgrep:ci.kafka-in-func-name
	hclList := testAccReplicatorHCLList

	// log_delivery is only rendered when the step exercises it, so the earlier steps keep a
	// configuration with no log_delivery block at all.
	logGroup, logDeliveryBlock := "", ""
	if logDelivery {
		logGroup = fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
		logDeliveryBlock = `
  log_delivery {
    replicator_log_delivery {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.test.name
      }
    }
  }
`
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
      topics_to_replicate = [%[8]s]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [%[9]s]
    }
  }
%[10]s
}
%[11]s
`, rName, clusterID, bootstrap, secretARN, targetARN, hclList(subnetIDs), hclList(securityGroupIDs), hclList(topics), hclList(consumerGroups), logDeliveryBlock, logGroup)
}

func testAccReplicatorConfig_selfManagedTarget(rName, clusterID, bootstrap, secretARN, mskClusterARN, subnetIDs, securityGroupIDs, topics string) string { // nosemgrep:ci.kafka-in-func-name
	hclList := testAccReplicatorHCLList

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

  # Source: an Amazon MSK cluster, referenced by ARN. vpc_config is supplied here because the
  # replicator reaches the Apache cluster through this VPC.
  kafka_cluster {
    amazon_msk_cluster {
      msk_cluster_arn = %[5]q
    }

    vpc_config {
      subnet_ids          = [%[6]s]
      security_groups_ids = [%[7]s]
    }
  }

  # Target: a cluster registered as self-managed Apache Kafka, referenced by its Kafka
  # cluster.id and reached over SASL/SCRAM.
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

  replication_info_list {
    source_kafka_cluster_arn = %[5]q
    target_kafka_cluster_id  = %[2]q
    target_compression_type  = "NONE"

    topic_replication {
      topics_to_replicate = [%[8]s]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [".*"]
    }
  }
}
`, rName, clusterID, bootstrap, secretARN, mskClusterARN, hclList(subnetIDs), hclList(securityGroupIDs), hclList(topics))
}
