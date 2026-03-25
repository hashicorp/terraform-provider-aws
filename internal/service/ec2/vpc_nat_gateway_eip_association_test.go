// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNATGatewayEIPAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var v types.NatGatewayAddress
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_nat_gateway_eip_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCNATGatewayEIPAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayEIPAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCNATGatewayEIPAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAssociationID), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "nat_gateway_id",
				ImportStateIdFunc:                    testAccVPCNATGatewayEIPAssociationImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccVPCNATGatewayEIPAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var v types.NatGatewayAddress
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_nat_gateway_eip_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCNATGatewayEIPAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayEIPAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCNATGatewayEIPAssociationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceNATGatewayEIPAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCNATGatewayEIPAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_nat_gateway_eip_association" {
				continue
			}

			_, err := tfec2.FindNATGatewayAddressByNATGatewayIDAndAllocationIDSucceeded(ctx, conn, rs.Primary.Attributes["nat_gateway_id"], rs.Primary.Attributes["allocation_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC NAT Gateway %s EIP %s Association still exists", rs.Primary.Attributes["nat_gateway_id"], rs.Primary.Attributes["allocation_id"])
		}

		return nil
	}
}

func testAccCheckVPCNATGatewayEIPAssociationExists(ctx context.Context, t *testing.T, n string, v *types.NatGatewayAddress) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		output, err := tfec2.FindNATGatewayAddressByNATGatewayIDAndAllocationIDSucceeded(ctx, conn, rs.Primary.Attributes["nat_gateway_id"], rs.Primary.Attributes["allocation_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCNATGatewayEIPAssociationImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", n)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["nat_gateway_id"], rs.Primary.Attributes["allocation_id"]), nil
	}
}

func testAccVPCNATGatewayEIPAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCNATGatewayConfig_basic(rName), fmt.Sprintf(`
resource "aws_eip" "secondary" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway_eip_association" "test" {
  allocation_id  = aws_eip.secondary.id
  nat_gateway_id = aws_nat_gateway.test.id
}
`, rName))
}
