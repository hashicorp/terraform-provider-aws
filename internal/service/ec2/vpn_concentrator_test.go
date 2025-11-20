// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2VPNConcentrator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpn_concentrator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPNConcentratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConcentratorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNConcentratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ipsec.1"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_attachment_id"),
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

func TestAccEC2VPNConcentrator_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpn_concentrator.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPNConcentratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConcentratorConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNConcentratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPNConcentratorConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNConcentratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPNConcentratorConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNConcentratorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckVPNConcentratorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpn_concentrator" {
				continue
			}

			_, err := tfec2.FindVPNConcentratorByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPN Concentrator %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPNConcentratorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindVPNConcentratorByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccVPNConcentratorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_concentrator" "test" {
  type               = "ipsec.1"
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}
`, rName)
}

func testAccVPNConcentratorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_concentrator" "test" {
  type               = "ipsec.1"
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPNConcentratorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_concentrator" "test" {
  type               = "ipsec.1"
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
