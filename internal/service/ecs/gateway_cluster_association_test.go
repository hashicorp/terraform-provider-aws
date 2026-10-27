// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSGatewayClusterAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway_cluster_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayClusterAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayClusterAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayClusterAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "cluster", "aws_ecs_cluster.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_arn", "aws_ecs_cluster.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "gateway", "aws_ecs_gateway.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", "aws_ecs_gateway.test", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGatewayClusterAssociationImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "cluster",
			},
		},
	})
}

func TestAccECSGatewayClusterAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway_cluster_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayClusterAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayClusterAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayClusterAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecs.ResourceGatewayClusterAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccECSGatewayClusterAssociation_Disappears_gateway(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway_cluster_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayClusterAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayClusterAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayClusterAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecs.ResourceGatewayClusterAssociation, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecs.ResourceGateway, "aws_ecs_gateway.test"),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_ecs_gateway.test", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckGatewayClusterAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_gateway_cluster_association" {
				continue
			}

			_, err := tfecs.FindGatewayClusterAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster"], rs.Primary.Attributes["gateway"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECS Gateway Cluster Association %s,%s still exists", rs.Primary.Attributes["cluster"], rs.Primary.Attributes["gateway"])
		}

		return nil
	}
}

func testAccCheckGatewayClusterAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		_, err := tfecs.FindGatewayClusterAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster"], rs.Primary.Attributes["gateway"])

		return err
	}
}

func testAccGatewayClusterAssociationImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "cluster", "gateway")
}

func testAccGatewayClusterAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_basic(rName), fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_gateway_cluster_association" "test" {
  cluster = aws_ecs_cluster.test.name
  gateway = aws_ecs_gateway.test.arn
}
`, rName))
}
