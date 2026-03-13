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
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaTopic_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var topic kafka.DescribeTopicOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("kafka", regexache.MustCompile(fmt.Sprintf(`topic/.+/.+/%s$`, rName)))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("configs_actual"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateIdFunc:                    testAccTopicImportStateIDFunc(resourceName),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"configs"},
			},
			{
				Config: testAccTopicConfig_basic(rName, clusterName, 3, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &topic),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccKafkaTopic_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var topic kafka.DescribeTopicOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfkafka.ResourceTopic, resourceName),
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

func TestAccKafkaTopic_configs(t *testing.T) {
	ctx := acctest.Context(t)
	var topic kafka.DescribeTopicOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clusterName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateIdFunc:                    testAccTopicImportStateIDFunc(resourceName),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"configs"},
			},
			{
				Config: testAccTopicConfig_configsUpdate(rName, clusterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &topic),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccTopicConfig_basic(rName, clusterName, 2, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &topic),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccTopicImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "cluster_arn", names.AttrName)
}

func testAccCheckTopicDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_topic" {
				continue
			}

			_, err := tfkafka.FindTopicByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster_arn"], rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK Topic %s still exists", rs.Primary.Attributes[names.AttrName])
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

		output, err := tfkafka.FindTopicByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster_arn"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTopicConfig_basic(rName, clusterName string, partitionCount, replicationFactor int) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(clusterName), fmt.Sprintf(`
resource "aws_msk_topic" "test" {
  name               = %[1]q
  cluster_arn        = aws_msk_cluster.test.arn
  partition_count    = %[2]d
  replication_factor = %[3]d
}
`, rName, partitionCount, replicationFactor))
}

func testAccTopicConfig_configs(rName, clusterName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(clusterName), fmt.Sprintf(`
resource "aws_msk_topic" "test" {
  name               = %[1]q
  cluster_arn        = aws_msk_cluster.test.arn
  partition_count    = 2
  replication_factor = 2

  configs = jsonencode({
    "retention.ms"        = "604800000"
    "retention.bytes"     = "-1",
    "cleanup.policy"      = "delete",
    "min.insync.replicas" = "2"
  })
}
`, rName))
}

func testAccTopicConfig_configsUpdate(rName, clusterName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(clusterName), fmt.Sprintf(`
resource "aws_msk_topic" "test" {
  name               = %[1]q
  cluster_arn        = aws_msk_cluster.test.arn
  partition_count    = 2
  replication_factor = 2

  configs = jsonencode({
    "retention.ms"        = "604800000"
    "compression.type"    = "snappy"
    "retention.bytes"     = "-1",
    "segment.bytes"       = "1073741824",
    "min.insync.replicas" = "3",
  })
}
`, rName))
}
