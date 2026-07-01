// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2OutpostsLocalGatewayRouteTable_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_local_gateway_route_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteTableConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ec2", regexache.MustCompile(`local-gateway-route-table/lgw-rtb-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "local_gateway_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrMode, "direct-vpc-routing"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOutpostARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "available"),
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

func TestAccEC2OutpostsLocalGatewayRouteTable_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_local_gateway_route_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteTableConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceLocalGatewayRouteTable, resourceName),
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

func TestAccEC2OutpostsLocalGatewayRouteTable_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_local_gateway_route_table.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteTableDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLocalGatewayRouteTableConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableExists(ctx, t, resourceName),
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
				Config: testAccLocalGatewayRouteTableConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocalGatewayRouteTableConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckLocalGatewayRouteTableExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		_, err := tfec2.FindLocalGatewayRouteTableByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckLocalGatewayRouteTableDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_local_gateway_route_table" {
				continue
			}

			_, err := tfec2.FindLocalGatewayRouteTableByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Local Gateway Route Table still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccLocalGatewayRouteTableBaseConfig() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway" "test" {
  id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}
`
}

func testAccLocalGatewayRouteTableConfig_basic() string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableBaseConfig(),
		`
resource "aws_ec2_local_gateway_route_table" "test" {
  local_gateway_id = data.aws_ec2_local_gateway.test.id
  mode             = "direct-vpc-routing"
}
`)
}

func testAccLocalGatewayRouteTableConfig_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableBaseConfig(),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table" "test" {
  local_gateway_id = data.aws_ec2_local_gateway.test.id
  mode             = "direct-vpc-routing"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccLocalGatewayRouteTableConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableBaseConfig(),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table" "test" {
  local_gateway_id = data.aws_ec2_local_gateway.test.id
  mode             = "direct-vpc-routing"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
