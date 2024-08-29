// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaVPCConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeVpcConnectionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_vpc_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "authentication", "SASL_IAM"),
					resource.TestCheckResourceAttr(resourceName, "client_subnets.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccKafkaVPCConnection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeVpcConnectionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_vpc_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCConnectionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCConnectionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccKafkaVPCConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v kafka.DescribeVpcConnectionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_vpc_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Kafka),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCConnectionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceVPCConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCConnectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_vpc_connection" {
				continue
			}

			_, err := tfkafka.FindVPCConnectionByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK VPC Connection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
func testAccCheckVPCConnectionExists(ctx context.Context, n string, v *kafka.DescribeVpcConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		output, err := tfkafka.FindVPCConnectionByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCConnectionConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_base(rName),
		testAccClusterConfig_allowEveryoneNoACLFoundFalse(rName),
		fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.8.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.test.id]

    connectivity_info {
      vpc_connectivity {
        client_authentication {
          sasl {
            iam = true
          }
        }
      }
    }

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

resource "aws_vpc" "client" {
  cidr_block = "10.0.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "client" {
  count = 3

  vpc_id            = aws_vpc.client.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.client.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCConnectionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCConnectionConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "client" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.client.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_msk_vpc_connection" "test" {
  authentication     = "SASL_IAM"
  target_cluster_arn = aws_msk_cluster.test.arn
  vpc_id             = aws_vpc.client.id
  client_subnets     = aws_subnet.client[*].id
  security_groups    = aws_security_group.client[*].id
}
`, rName))
}

func testAccVPCConnectionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCConnectionConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "client" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.client.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_msk_vpc_connection" "test" {
  authentication     = "SASL_IAM"
  target_cluster_arn = aws_msk_cluster.test.arn
  vpc_id             = aws_vpc.client.id
  client_subnets     = aws_subnet.client[*].id
  security_groups    = aws_security_group.client[*].id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccVPCConnectionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCConnectionConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "client" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.client.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_msk_vpc_connection" "test" {
  authentication     = "SASL_IAM"
  target_cluster_arn = aws_msk_cluster.test.arn
  vpc_id             = aws_vpc.client.id
  client_subnets     = aws_subnet.client[*].id
  security_groups    = aws_security_group.client[*].id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
