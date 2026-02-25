// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaTopic_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeTopicOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_topic.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kafka", regexache.MustCompile(`topic/.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_arn", "aws_msk_cluster.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "topic_name", rName),
					resource.TestCheckResourceAttr(resourceName, "partition_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "replication_factor", "3"),
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

func TestAccKafkaTopic_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeTopicOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_topic.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkafka.ResourceTopic(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaTopic_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeTopicOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_topic.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "partition_count", "3"),
				),
			},
			{
				Config: testAccTopicConfig_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "partition_count", "6"),
					resource.TestCheckResourceAttrSet(resourceName, "configs"),
				),
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

			parts, err := tfkafka.FindTopicByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster_arn"], rs.Primary.Attributes["topic_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if parts != nil {
				return fmt.Errorf("MSK Topic %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckTopicExists(ctx context.Context, t *testing.T, n string, v *kafka.DescribeTopicOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		output, err := tfkafka.FindTopicByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster_arn"], rs.Primary.Attributes["topic_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTopicConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "3.6.0"
  number_of_broker_nodes = 3

  broker_node_group_info {
    instance_type   = "kafka.m5.large"
    client_subnets  = aws_subnet.test[*].id
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

func testAccTopicConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTopicConfig_base(rName), fmt.Sprintf(`
resource "aws_msk_topic" "test" {
  cluster_arn        = aws_msk_cluster.test.arn
  topic_name         = %[1]q
  partition_count    = 3
  replication_factor = 3
}
`, rName))
}

func testAccTopicConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccTopicConfig_base(rName), fmt.Sprintf(`
resource "aws_msk_topic" "test" {
  cluster_arn        = aws_msk_cluster.test.arn
  topic_name         = %[1]q
  partition_count    = 6
  replication_factor = 3
  configs            = base64encode("retention.ms=86400000")
}
`, rName))
}
