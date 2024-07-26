// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayPolicyTableAssociation_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayPolicyTableAssociation
	resourceName := "aws_ec2_transit_gateway_policy_table_association.test"
	transitGatewayPolicyTableResourceName := "aws_ec2_transit_gateway_policy_table.test"
	transitGatewayPeeringResourceName := "aws_networkmanager_transit_gateway_peering.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPolicyTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPolicyTableAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPolicyTableAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceType),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayPeeringResourceName, "transit_gateway_peering_attachment_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_policy_table_id", transitGatewayPolicyTableResourceName, names.AttrID),
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

func testAccTransitGatewayPolicyTableAssociation_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayPolicyTableAssociation
	resourceName := "aws_ec2_transit_gateway_policy_table_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
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

func testAccCheckTransitGatewayPolicyTableAssociationExists(ctx context.Context, n string, v *awstypes.TransitGatewayPolicyTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayPolicyTableAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_policy_table_id"], rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayPolicyTableAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_policy_table_association" {
				continue
			}

			_, err := tfec2.FindTransitGatewayPolicyTableAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_policy_table_id"], rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])

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

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
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
  core_network_id     = aws_networkmanager_core_network_policy_attachment.test.core_network_id
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
