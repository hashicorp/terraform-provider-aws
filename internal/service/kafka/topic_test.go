// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaTopic_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var topic kafka.DescribeTopicOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_topic.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_basic(rName, clusterName, 2, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &topic),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kafka", regexache.MustCompile(fmt.Sprintf(`topic/.+/.+/%s$`, rName))),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrName, "cluster_arn"),
				ImportState:                          true,
				ImportStateVerify:                    true,
			},
		},
	})
}

func TestAccKafkaTopic_configs(t *testing.T) {
	ctx := acctest.Context(t)

	var topic kafka.DescribeTopicOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_topic.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_configs(rName, clusterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &topic),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kafka", regexache.MustCompile(fmt.Sprintf(`topic/.+/.+/%s$`, rName))),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrName, "cluster_arn"),
			},
		},
	})
}

func testAccCheckTopicDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_topic" {
				continue
			}

			_, err := tfkafka.FindTopicByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["cluster_arn"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Kafka, create.ErrActionCheckingDestroyed, tfkafka.ResNameTopic, rs.Primary.ID, err)
			}

			return create.Error(names.Kafka, create.ErrActionCheckingDestroyed, tfkafka.ResNameTopic, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTopicExists(ctx context.Context, t *testing.T, name string, topic *kafka.DescribeTopicOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameTopic, name, errors.New("not found"))
		}

		name := rs.Primary.Attributes[names.AttrName]
		clusterARN := rs.Primary.Attributes["cluster_arn"]
		if name == "" {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameTopic, name, errors.New("topic name not set"))
		}
		if clusterARN == "" {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameTopic, name, errors.New("topic cluster ARN not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		resp, err := tfkafka.FindTopicByTwoPartKey(ctx, conn, name, clusterARN)
		if err != nil {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameTopic, rs.Primary.ID, err)
		}

		*topic = *resp

		return nil
	}
}

func testAccTopicConfig_basic(rName, clusterName string, partitionCount, replicationFactor int) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(clusterName), fmt.Sprintf(`
resource "aws_msk_topic" "test" {
  name               = %[1]q
  cluster_arn        = aws_msk_cluster.test.arn
  partition_count    = %[3]d
  replication_factor = %[4]d
}

`, rName, clusterName, partitionCount, replicationFactor))
}

func testAccTopicConfig_configs(rName, clusterName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(clusterName), fmt.Sprintf(`
resource "aws_msk_topic" "test" {
  name               = %[1]q
  cluster_arn        = aws_msk_cluster.test.arn
  partition_count    = 2
  replication_factor = 2

  configs = base64encode(jsonencode({
    "retention.ms"        = "604800000"
    "retention.bytes"     = "-1",
    "cleanup.policy"      = "delete",
    "min.insync.replicas" = "2"
  }))
}

`, rName, clusterName))
}
