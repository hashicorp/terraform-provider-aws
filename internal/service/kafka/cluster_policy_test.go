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

func TestAccKafkaClusterPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var clusterpolicy kafka.GetClusterPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, t, resourceName, &clusterpolicy),
					resource.TestCheckResourceAttrSet(resourceName, "current_version"),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"kafka:Get\*","kafka:CreateVpcConnection"`)),
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

func TestAccKafkaClusterPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var clusterpolicy kafka.GetClusterPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, t, resourceName, &clusterpolicy),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkafka.ResourceClusterPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaClusterPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var clusterpolicy kafka.GetClusterPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, t, resourceName, &clusterpolicy),
					resource.TestCheckResourceAttrSet(resourceName, "current_version"),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"kafka:Get\*","kafka:CreateVpcConnection"`)),
				),
			},
			{
				Config: testAccClusterPolicyConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, t, resourceName, &clusterpolicy),
					resource.TestCheckResourceAttrSet(resourceName, "current_version"),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"kafka:DescribeCluster","kafka:DescribeClusterV2"`)),
				),
			},
		},
	})
}

func testAccCheckClusterPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_cluster_policy" {
				continue
			}

			_, err := tfkafka.FindClusterPolicyByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK Cluster Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterPolicyExists(ctx context.Context, t *testing.T, n string, v *kafka.GetClusterPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		output, err := tfkafka.FindClusterPolicyByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCConnectionConfig_basic(rName), `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_msk_cluster_policy" "test" {
  cluster_arn = aws_msk_cluster.test.arn

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Sid    = "testMskClusterPolicy"
      Effect = "Allow"
      Principal = {
        "AWS" = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "kafka:Describe*",
        "kafka:Get*",
        "kafka:CreateVpcConnection",
      ]
      Resource = aws_msk_cluster.test.arn
    }]
  })

  depends_on = [aws_msk_vpc_connection.test]
}
`)
}

func testAccClusterPolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccVPCConnectionConfig_basic(rName), `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_msk_cluster_policy" "test" {
  cluster_arn = aws_msk_cluster.test.arn

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Sid    = "testMskClusterPolicy"
      Effect = "Allow"
      Principal = {
        "AWS" = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "kafka:CreateVpcConnection",
        "kafka:GetBootstrapBrokers",
        "kafka:DescribeCluster",
        "kafka:DescribeClusterV2",
      ]
      Resource = aws_msk_cluster.test.arn
    }]
  })

  depends_on = [aws_msk_vpc_connection.test]
}
`)
}
