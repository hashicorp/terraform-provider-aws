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

func TestAccKafkaSingleSCRAMSecretAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_single_scram_secret_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSingleSCRAMSecretAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSingleSCRAMSecretAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSingleSCRAMSecretAssociationExists(ctx, resourceName),
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

func TestAccKafkaSingleSCRAMSecretAssociation_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_msk_single_scram_secret_association.test1"
	resource2Name := "aws_msk_single_scram_secret_association.test2"
	resource3Name := "aws_msk_single_scram_secret_association.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSingleSCRAMSecretAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSingleSCRAMSecretAssociationConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSingleSCRAMSecretAssociationExists(ctx, resource1Name),
					testAccCheckSingleSCRAMSecretAssociationExists(ctx, resource2Name),
					testAccCheckSingleSCRAMSecretAssociationExists(ctx, resource3Name),
				),
			},
		},
	})
}

func TestAccKafkaSingleSCRAMSecretAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_msk_single_scram_secret_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KafkaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSingleSCRAMSecretAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSingleSCRAMSecretAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSingleSCRAMSecretAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfkafka.ResourceSingleSCRAMSecretAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSingleSCRAMSecretAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_msk_single_scram_secret_association" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

			err := tfkafka.FindSingleSCRAMSecretAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster_arn"], rs.Primary.Attributes["secret_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MSK Single SCRAM Secret Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSingleSCRAMSecretAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KafkaClient(ctx)

		return tfkafka.FindSingleSCRAMSecretAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster_arn"], rs.Primary.Attributes["secret_arn"])
	}
}

func testAccSingleSCRAMSecretAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSCRAMSecretAssociationConfig_base(rName, 1), `
resource "aws_msk_single_scram_secret_association" "test" {
  cluster_arn = aws_msk_cluster.test.arn
  secret_arn  = aws_secretsmanager_secret.test[0].arn

  depends_on = [aws_secretsmanager_secret_version.test]
}
`)
}

func testAccSingleSCRAMSecretAssociationConfig_multiple(rName string) string {
	return acctest.ConfigCompose(testAccSCRAMSecretAssociationConfig_base(rName, 3), `
resource "aws_msk_single_scram_secret_association" "test1" {
  cluster_arn = aws_msk_cluster.test.arn
  secret_arn  = aws_secretsmanager_secret.test[0].arn

  depends_on = [aws_secretsmanager_secret_version.test]
}

resource "aws_msk_single_scram_secret_association" "test2" {
  cluster_arn = aws_msk_cluster.test.arn
  secret_arn  = aws_secretsmanager_secret.test[1].arn

  depends_on = [aws_secretsmanager_secret_version.test]
}

resource "aws_msk_single_scram_secret_association" "test3" {
  cluster_arn = aws_msk_cluster.test.arn
  secret_arn  = aws_secretsmanager_secret.test[2].arn

  depends_on = [aws_secretsmanager_secret_version.test]
}
`)
}
