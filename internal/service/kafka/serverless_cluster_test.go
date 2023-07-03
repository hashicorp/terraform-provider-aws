// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccKafkaServerlessCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_serverless_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kafka.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerlessClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kafka", regexp.MustCompile(`cluster/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_authentication.0.sasl.0.iam.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var v kafka.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_serverless_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kafka.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerlessClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceServerlessCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaServerlessCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_serverless_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kafka.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerlessClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessClusterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServerlessClusterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServerlessClusterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccKafkaServerlessCluster_securityGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_serverless_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, kafka.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerlessClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessClusterConfig_securityGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessClusterExists(ctx, resourceName, &v),
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

func testAccCheckServerlessClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_serverless_cluster" {
				continue
			}

			_, err := tfkafka.FindServerlessClusterByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckServerlessClusterExists(ctx context.Context, n string, v *kafka.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MSK Serverless Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaConn(ctx)

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
