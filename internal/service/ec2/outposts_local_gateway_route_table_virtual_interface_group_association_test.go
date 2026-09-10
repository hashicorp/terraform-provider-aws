// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2OutpostsLocalGatewayRouteTableVirtualInterfaceGroupAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_local_gateway_route_table_virtual_interface_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteTableVIFGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteTableVIFGroupAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVIFGroupAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "local_gateway_id"),
					resource.TestCheckResourceAttrSet(resourceName, "local_gateway_route_table_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "local_gateway_route_table_id"),
					resource.TestCheckResourceAttrSet(resourceName, "local_gateway_virtual_interface_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccEC2OutpostsLocalGatewayRouteTableVirtualInterfaceGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_local_gateway_route_table_virtual_interface_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteTableVIFGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteTableVIFGroupAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVIFGroupAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceLocalGatewayRouteTableVIFGroupAssociation, resourceName),
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

func TestAccEC2OutpostsLocalGatewayRouteTableVirtualInterfaceGroupAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_local_gateway_route_table_virtual_interface_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteTableVIFGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteTableVIFGroupAssociationConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVIFGroupAssociationExists(ctx, t, resourceName),
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
				Config: testAccLocalGatewayRouteTableVIFGroupAssociationConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVIFGroupAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocalGatewayRouteTableVIFGroupAssociationConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVIFGroupAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckLocalGatewayRouteTableVIFGroupAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		_, err := tfec2.FindLocalGatewayRouteTableVIFGroupAssociationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckLocalGatewayRouteTableVIFGroupAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_local_gateway_route_table_virtual_interface_group_association" {
				continue
			}

			_, err := tfec2.FindLocalGatewayRouteTableVIFGroupAssociationByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Local Gateway Route Table Virtual Interface Group Association still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccLocalGatewayRouteTableVIFGroupAssociationBaseConfig() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = data.aws_ec2_local_gateway_route_table.test.local_gateway_id
}
`
}

func testAccLocalGatewayRouteTableVIFGroupAssociationConfig_basic() string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableVIFGroupAssociationBaseConfig(),
		`
resource "aws_ec2_local_gateway_route_table_virtual_interface_group_association" "test" {
  local_gateway_route_table_id             = data.aws_ec2_local_gateway_route_table.test.id
  local_gateway_virtual_interface_group_id = data.aws_ec2_local_gateway_virtual_interface_group.test.id
}
`)
}

func testAccLocalGatewayRouteTableVIFGroupAssociationConfig_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableVIFGroupAssociationBaseConfig(),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table_virtual_interface_group_association" "test" {
  local_gateway_route_table_id             = data.aws_ec2_local_gateway_route_table.test.id
  local_gateway_virtual_interface_group_id = data.aws_ec2_local_gateway_virtual_interface_group.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccLocalGatewayRouteTableVIFGroupAssociationConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableVIFGroupAssociationBaseConfig(),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table_virtual_interface_group_association" "test" {
  local_gateway_route_table_id             = data.aws_ec2_local_gateway_route_table.test.id
  local_gateway_virtual_interface_group_id = data.aws_ec2_local_gateway_virtual_interface_group.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
