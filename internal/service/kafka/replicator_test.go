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
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

// func TestAccKafkaReplicator_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var replicator kafka.DescribeReplicatorOutput
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_kafka_replicator.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.Kafka)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckReplicatorDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccReplicatorConfig_basic(rName, testAccReplicatorVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckReplicatorExists(ctx, resourceName, &replicator),
// 					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceReplicator(), resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccCheckReplicatorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kafka_replicator" {
				continue
			}

			// input := &kafka.DescribeReplicatorInput{
			// 	ReplicatorArn: aws.String(rs.Primary.ID),
			// }
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
      msk_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-Cluster/7edb97d0-ed0b-42f9-aa35-c70549c3bdf7-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0ab7b209961e28f62", "subnet-0da9b023cd33bbdf2"]
      security_group_ids = ["sg-064bbd53a29a42b93"]
    }
  }
  kafka_clusters {
    amazon_msk_cluster {
      msk_cluster_arn = "arn:aws:kafka:us-west-2:926562225508:cluster/Destination-Cluster/bc2049c9-608e-4d00-8058-8d277340f415-14"
    }

    vpc_config {
      subnet_ids         = ["subnet-0b0126f677f00f9fd", "subnet-09bb16bf115c8d87e"]
      security_group_ids = ["sg-04842606bd0dc0ccc"]
    }
  }

  replication_info_list {
    source_kafka_cluster_arn = "arn:aws:kafka:us-east-1:926562225508:cluster/Source-Cluster/7edb97d0-ed0b-42f9-aa35-c70549c3bdf7-14"
    target_kafka_cluster_arn = "arn:aws:kafka:us-west-2:926562225508:cluster/Destination-Cluster/bc2049c9-608e-4d00-8058-8d277340f415-14"
    target_compression_type  = ["GZI"]

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
