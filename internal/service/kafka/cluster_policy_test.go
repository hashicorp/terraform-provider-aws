package kafka_test

import (
	"context"
	"errors"
	"regexp"
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

func TestAccKafkaClusterPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var clusterpolicy kafka.GetClusterPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, resourceName, &clusterpolicy),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_arn"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"kafka:CreateVpcConnection","kafka:GetBootstrapBrokers"`)),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_arn", "aws_msk_cluster.test", "arn"),
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

func TestAccKafkaClusterPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var clusterpolicy kafka.GetClusterPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_cluster_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Kafka)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterPolicyExists(ctx, resourceName, &clusterpolicy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceClusterPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_cluster_policy" {
				continue
			}

			_, err := conn.GetClusterPolicy(ctx, &kafka.GetClusterPolicyInput{
				ClusterArn: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.NotFoundException](err) {
				return nil
			}
			if err != nil {
				return nil
			}

			return create.Error(names.Kafka, create.ErrActionCheckingDestroyed, tfkafka.ResNameClusterPolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckClusterPolicyExists(ctx context.Context, name string, clusterpolicy *kafka.GetClusterPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameClusterPolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameClusterPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)
		resp, err := conn.GetClusterPolicy(ctx, &kafka.GetClusterPolicyInput{
			ClusterArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Kafka, create.ErrActionCheckingExistence, tfkafka.ResNameClusterPolicy, rs.Primary.ID, err)
		}

		*clusterpolicy = *resp

		return nil
	}
}

func testAccClusterPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), `
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
        "kafka:GetBootstrapBrokers",
      ]
      Resource = aws_msk_cluster.test.arn
    }]
  })
}
`)
}
