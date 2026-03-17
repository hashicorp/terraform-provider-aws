// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaServerlessCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Cluster
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_serverless_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerlessClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kafka", regexache.MustCompile(`cluster/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "bootstrap_brokers_sasl_iam"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_uuid"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
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

func TestAccKafkaServerlessCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Cluster
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_serverless_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerlessClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkafka.ResourceServerlessCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaServerlessCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Cluster
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_serverless_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerlessClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessClusterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, t, resourceName, &v),
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
				Config: testAccServerlessClusterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccServerlessClusterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKafkaServerlessCluster_securityGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Cluster
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_msk_serverless_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerlessClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessClusterConfig_securityGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
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

func testAccCheckServerlessClusterDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_serverless_cluster" {
				continue
			}

			_, err := tfkafka.FindServerlessClusterByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK Serverless Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServerlessClusterExists(ctx context.Context, t *testing.T, n string, v *types.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Serverless Cluster ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).KafkaClient(ctx)

		output, err := tfkafka.FindServerlessClusterByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccServerlessClusterConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccServerlessClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccServerlessClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_msk_serverless_cluster" "test" {
  cluster_name = %[1]q

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }
}
`, rName))
}

func testAccServerlessClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccServerlessClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_msk_serverless_cluster" "test" {
  cluster_name = %[1]q

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccServerlessClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccServerlessClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_msk_serverless_cluster" "test" {
  cluster_name = %[1]q

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccServerlessClusterConfig_securityGroup(rName string) string {
	return acctest.ConfigCompose(testAccServerlessClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_msk_serverless_cluster" "test" {
  cluster_name = %[1]q

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = aws_subnet.test[*].id
  }
}
`, rName))
}
