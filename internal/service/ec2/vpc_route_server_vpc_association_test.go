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
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCRouteServerVPCAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RouteServerAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpc_route_server_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerVPCAssociationConfig_basic(rName, rAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerVPCAssociationExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "route_server_id",
				ImportStateIdFunc:                    testAccVPCRouteServerVPCAssociationImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccVPCRouteServerVPCAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RouteServerAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpc_route_server_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerVPCAssociationConfig_basic(rName, rAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerVPCAssociationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCRouteServerVPCAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCRouteServerVPCAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_route_server_vpc_association" {
				continue
			}

			_, err := tfec2.FindRouteServerAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["route_server_id"], rs.Primary.Attributes[names.AttrVPCID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Route Server %s VPC %s Association still exists", rs.Primary.Attributes["route_server_id"], rs.Primary.Attributes[names.AttrVPCID])
		}

		return nil
	}
}

func testAccCheckVPCRouteServerVPCAssociationExists(ctx context.Context, n string, v *awstypes.RouteServerAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindRouteServerAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["route_server_id"], rs.Primary.Attributes[names.AttrVPCID])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCRouteServerVPCAssociationImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", n)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["route_server_id"], rs.Primary.Attributes[names.AttrVPCID]), nil
	}
}

func testAccVPCRouteServerVPCAssociationConfig_basic(rName string, rAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_route_server" "test" {
  amazon_side_asn = %[2]d

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_route_server_vpc_association" "test" {
  route_server_id = aws_vpc_route_server.test.route_server_id
  vpc_id          = aws_vpc.test.id
}
`, rName, rAsn)
}
