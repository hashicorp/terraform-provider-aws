// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccTransitGatewayPolicyTableAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.TransitGatewayPolicyTableAssociation
	resourceName := "aws_ec2_transit_gateway_policy_table_association.test"
	transitGatewayPolicyTableResourceName := "aws_ec2_transit_gateway_policy_table.test"
	transitGatewayPeeringResourceName := "aws_networkmanager_transit_gateway_peering.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPolicyTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPolicyTableAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_type"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayPeeringResourceName, "transit_gateway_peering_attachment_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_policy_table_id", transitGatewayPolicyTableResourceName, "id"),
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

func testAccTransitGatewayPolicyTableAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.TransitGatewayPolicyTableAssociation
	resourceName := "aws_ec2_transit_gateway_policy_table_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPolicyTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPolicyTableAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableAssociationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayPolicyTableAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTransitGatewayPolicyTableAssociationExists(ctx context.Context, n string, v *ec2.TransitGatewayPolicyTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Policy Table Association ID is set")
		}

		transitGatewayPolicyTableID, transitGatewayAttachmentID, err := tfec2.TransitGatewayPolicyTableAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		output, err := tfec2.FindTransitGatewayPolicyTableAssociationByTwoPartKey(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayPolicyTableAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_policy_table_association" {
				continue
			}

			transitGatewayPolicyTableID, transitGatewayAttachmentID, err := tfec2.TransitGatewayPolicyTableAssociationParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfec2.FindTransitGatewayPolicyTableAssociationByTwoPartKey(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Policy Table Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayPolicyTableAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_policy_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  policy_document   = data.aws_networkmanager_core_network_policy_document.test.json

  tags = {
    Name = %[1]q
  }
}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    # Don't overlap with default TGW ASN: 64512.
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = data.aws_region.current.name
    }
  }

  segments {
    name = "test"
  }
}

resource "aws_networkmanager_transit_gateway_peering" "test" {
  core_network_id     = aws_networkmanager_core_network.test.id
  transit_gateway_arn = aws_ec2_transit_gateway.test.arn

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_transit_gateway_policy_table.test]
}

resource "aws_ec2_transit_gateway_policy_table_association" "test" {
  transit_gateway_attachment_id   = aws_networkmanager_transit_gateway_peering.test.transit_gateway_peering_attachment_id
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
}
`, rName)
}
