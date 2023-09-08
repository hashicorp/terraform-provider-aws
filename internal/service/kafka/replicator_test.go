package kafka_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
)

func TestAccKafkaReplicator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccReplicatorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
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
				Config: testAccReplicatorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccReplicatorConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_exclude.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_exclude.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccKafkaReplicator_GZIP(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccReplicatorConfig_GZIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}
func TestAccKafkaReplicator_LZ4(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccReplicatorConfig_LZ4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "LZ4"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}
func TestAccKafkaReplicator_SNAPPY(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccReplicatorConfig_SNAPPY(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "SNAPPY"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccKafkaReplicator_ZSTD(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicator kafka.DescribeReplicatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccReplicatorConfig_ZSTD(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
					resource.TestCheckResourceAttr(resourceName, "replicator_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test-description"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.0.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.amazon_msk_cluster.0.msk_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "kafka_clusters.1.vpc_config.0.security_groups_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.source_kafka_cluster_arn", "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.target_compression_type", "ZSTD"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.topic_replication.0.topics_to_replicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_info_list.0.consumer_group_replication.0.consumer_groups_to_replicate.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
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
				Config: testAccReplicatorConfig_basic(rName),
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

			_, err := conn.DescribeReplicator(ctx, &kafka.DescribeReplicatorInput{
				ReplicatorArn: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.NotFoundException](err) {
				return nil
			}
			if err != nil {
				return nil
			}

			return create.Error(names.Kafka, create.ErrActionCheckingDestroyed, tfkafka.ResNameReplicator, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckReplicatorExists(ctx context.Context, name string, replicator *kafka.DescribeReplicatorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameReplicator, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameReplicator, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)
		resp, err := conn.DescribeReplicator(ctx, &kafka.DescribeReplicatorInput{
			ReplicatorArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameReplicator, rs.Primary.ID, err)
		}

		*replicator = *resp

		return nil
	}
}

func testAccReplicatorConfig_basic(rName string) string {
	return fmt.Sprintf(`

resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = "arn:aws:iam::926562225508:role/MaskReplicatorRole"

  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0fa0375b5a6300678", "subnet-0940965bc06ac84a0", "subnet-00b61c8bfca665931"]
      security_groups_ids = ["sg-09c6ec0f574d9eb13"]
    }
  }
  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0c428011e6bef0a18", "subnet-0ab47999ff2951ab5", "subnet-0f80dda415ee7d008"]
      security_groups_ids = ["sg-03377f7d2be71cdb6"]
    }
  }

  replication_info_list {
	target_kafka_cluster_arn = 	"arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    source_kafka_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    target_compression_type  = "NONE"


    topic_replication {
      topics_to_replicate = [".*"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = [".*"]
    }
  }
}
`, rName)
}

func testAccReplicatorConfig_update(rName string) string {
	return fmt.Sprintf(`

resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = "arn:aws:iam::926562225508:role/MaskReplicatorRole"

  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0fa0375b5a6300678", "subnet-0940965bc06ac84a0", "subnet-00b61c8bfca665931"]
      security_groups_ids = ["sg-09c6ec0f574d9eb13"]
    }
  }
  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0c428011e6bef0a18", "subnet-0ab47999ff2951ab5", "subnet-0f80dda415ee7d008"]
      security_groups_ids = ["sg-03377f7d2be71cdb6"]
    }
  }

  replication_info_list {
	target_kafka_cluster_arn = 	"arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    source_kafka_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    target_compression_type  = "NONE"


    topic_replication {
      topics_to_replicate = ["topic1", "topic2", "topic3"]
	  topics_to_exclude   = ["topic-4"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = ["group1", "group2", "group3"]
	  consumer_groups_to_exclude   = ["group-4"]
    }
  }
}
`, rName)
}

func testAccReplicatorConfig_GZIP(rName string) string {
	return fmt.Sprintf(`

resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = "arn:aws:iam::926562225508:role/MaskReplicatorRole"

  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0fa0375b5a6300678", "subnet-0940965bc06ac84a0", "subnet-00b61c8bfca665931"]
      security_groups_ids = ["sg-09c6ec0f574d9eb13"]
    }
  }
  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0c428011e6bef0a18", "subnet-0ab47999ff2951ab5", "subnet-0f80dda415ee7d008"]
      security_groups_ids = ["sg-03377f7d2be71cdb6"]
    }
  }

  replication_info_list {
	target_kafka_cluster_arn = 	"arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    source_kafka_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    target_compression_type  = "GZIP"

    topic_replication {
      topics_to_replicate = ["test-topic"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = ["test-consumer"]
    }
  }
}
`, rName)
}

func testAccReplicatorConfig_LZ4(rName string) string {
	return fmt.Sprintf(`

resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = "arn:aws:iam::926562225508:role/MaskReplicatorRole"

  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0fa0375b5a6300678", "subnet-0940965bc06ac84a0", "subnet-00b61c8bfca665931"]
      security_groups_ids = ["sg-09c6ec0f574d9eb13"]
    }
  }
  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0c428011e6bef0a18", "subnet-0ab47999ff2951ab5", "subnet-0f80dda415ee7d008"]
      security_groups_ids = ["sg-03377f7d2be71cdb6"]
    }
  }

  replication_info_list {
	target_kafka_cluster_arn = 	"arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    source_kafka_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    target_compression_type  = "LZ4"

    topic_replication {
      topics_to_replicate = ["test-topic"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = ["test-consumer"]
    }
  }
}
`, rName)
}

func testAccReplicatorConfig_SNAPPY(rName string) string {
	return fmt.Sprintf(`

resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = "arn:aws:iam::926562225508:role/MaskReplicatorRole"

  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0fa0375b5a6300678", "subnet-0940965bc06ac84a0", "subnet-00b61c8bfca665931"]
      security_groups_ids = ["sg-09c6ec0f574d9eb13"]
    }
  }
  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0c428011e6bef0a18", "subnet-0ab47999ff2951ab5", "subnet-0f80dda415ee7d008"]
      security_groups_ids = ["sg-03377f7d2be71cdb6"]
    }
  }

  replication_info_list {
	target_kafka_cluster_arn = 	"arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    source_kafka_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    target_compression_type  = "SNAPPY"

    topic_replication {
      topics_to_replicate = ["test-topic"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = ["test-consumer"]
    }
  }
}
`, rName)
}

func testAccReplicatorConfig_ZSTD(rName string) string {
	return fmt.Sprintf(`

resource "aws_msk_replicator" "test" {
  replicator_name            = %[1]q
  description                = "test-description"
  service_execution_role_arn = "arn:aws:iam::926562225508:role/MaskReplicatorRole"

  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0fa0375b5a6300678", "subnet-0940965bc06ac84a0", "subnet-00b61c8bfca665931"]
      security_groups_ids = ["sg-09c6ec0f574d9eb13"]
    }
  }
  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0c428011e6bef0a18", "subnet-0ab47999ff2951ab5", "subnet-0f80dda415ee7d008"]
      security_groups_ids = ["sg-03377f7d2be71cdb6"]
    }
  }

  replication_info_list {
	target_kafka_cluster_arn = 	"arn:aws:kafka:us-east-1:926562225508:cluster/Target-MSK-Replicator/e1289fbf-e895-464a-afba-e2aa0735cfd2-14"
    source_kafka_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-MSK-Replicator/9127030a-2c7b-4aea-a5a0-49f978b72f7d-14"
    target_compression_type  = "ZSTD"

    topic_replication {
      topics_to_replicate = ["test-topic"]
    }

    consumer_group_replication {
      consumer_groups_to_replicate = ["test-consumer"]
    }
  }
}
`, rName)
}
