// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKafkaScramSecretAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_scram_secret_association.test"
	clusterResourceName := "aws_msk_cluster.test"
	secretResourceName := "aws_secretsmanager_secret.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScramSecretAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScramSecretAssociationConfig_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScramSecretAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_arn", clusterResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "secret_arn_list.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName, names.AttrARN),
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

func TestAccKafkaScramSecretAssociation_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_scram_secret_association.test"
	secretResourceName := "aws_secretsmanager_secret.test.0"
	secretResourceName2 := "aws_secretsmanager_secret.test.1"
	secretResourceName3 := "aws_secretsmanager_secret.test.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScramSecretAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScramSecretAssociationConfig_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScramSecretAssociationExists(ctx, resourceName),
				),
			},
			{
				Config: testAccScramSecretAssociationConfig_basic(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScramSecretAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "secret_arn_list.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName2, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName3, names.AttrARN),
				),
			},
			{
				Config: testAccScramSecretAssociationConfig_basic(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScramSecretAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "secret_arn_list.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName2, names.AttrARN),
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

func TestAccKafkaScramSecretAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_scram_secret_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScramSecretAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScramSecretAssociationConfig_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScramSecretAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceSCRAMSecretAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKafkaScramSecretAssociation_Disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_scram_secret_association.test"
	clusterResourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScramSecretAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScramSecretAssociationConfig_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScramSecretAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckScramSecretAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_scram_secret_association" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

			_, err := tfkafka.FindSCRAMSecretsByClusterARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckScramSecretAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		_, err := tfkafka.FindSCRAMSecretsByClusterARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccScramSecretAssociationConfig_base(rName string, count int) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.8.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.test.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }

  client_authentication {
    sasl {
      scram = true
    }
  }
}

resource "aws_kms_key" "test" {
  count       = %[2]d
  description = "%[1]s-${count.index + 1}"
}

resource "aws_secretsmanager_secret" "test" {
  count      = %[2]d
  name       = "AmazonMSK_%[1]s-${count.index + 1}"
  kms_key_id = aws_kms_key.test[count.index].id
}

resource "aws_secretsmanager_secret_version" "test" {
  count         = %[2]d
  secret_id     = aws_secretsmanager_secret.test[count.index].id
  secret_string = jsonencode({ username = "user", password = "pass" })
}

resource "aws_secretsmanager_secret_policy" "test" {
  count      = %[2]d
  secret_arn = aws_secretsmanager_secret.test[count.index].arn
  policy     = <<POLICY
{
  "Version" : "2012-10-17",
  "Statement" : [ {
    "Sid": "AWSKafkaResourcePolicy",
    "Effect" : "Allow",
    "Principal" : {
      "Service" : "kafka.${data.aws_partition.current.dns_suffix}"
    },
    "Action" : "secretsmanager:getSecretValue",
    "Resource" : "${aws_secretsmanager_secret.test[count.index].arn}"
  } ]
}
POLICY
}
`, rName, count))
}

func testAccScramSecretAssociationConfig_basic(rName string, count int) string {
	return acctest.ConfigCompose(testAccScramSecretAssociationConfig_base(rName, count), `
resource "aws_msk_scram_secret_association" "test" {
  cluster_arn     = aws_msk_cluster.test.arn
  secret_arn_list = aws_secretsmanager_secret.test[*].arn

  depends_on = [aws_secretsmanager_secret_version.test]
}
`)
}
