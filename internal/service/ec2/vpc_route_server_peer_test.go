// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCRouteServerPeer_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var VPCRouteServerPeer awstypes.RouteServerPeer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_route_server_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerPeerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerPeerExists(ctx, resourceName, &VPCRouteServerPeer),
					resource.TestCheckResourceAttrSet(resourceName, "route_server_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_eni_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_eni_address"),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_options.peer_asn"),
					resource.TestCheckResourceAttrSet(resourceName, "bfd_status.status"),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_status.status"),
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

func TestAccVPCRouteServerPeer_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var VPCRouteServerPeer awstypes.RouteServerPeer
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_route_server_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerPeerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerPeerExists(ctx, resourceName, &VPCRouteServerPeer),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCRouteServerPeer, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCRouteServerPeerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_route_server_peer" {
				continue
			}

			_, err := tfec2.FindVPCRouteServerPeerByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCRouteServerPeer, rs.Primary.ID, err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCRouteServerPeer, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckVPCRouteServerPeerExists(ctx context.Context, name string, VPCRouteServerPeer *awstypes.RouteServerPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServerPeer, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServerPeer, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := tfec2.FindVPCRouteServerPeerByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServerPeer, rs.Primary.ID, err)
		}

		*VPCRouteServerPeer = *resp

		return nil
	}
}

func testAccVPCRouteServerPeerConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.0.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_route_server" "test" {
  amazon_side_asn = 4294967294
  tags = {
	Name = %[1]q
  }
}

resource "aws_vpc_route_server_association" "test" {
  route_server_id = aws_vpc_route_server.test.id
  vpc_id          = aws_vpc.test.id
}

resource "aws_vpc_route_server_endpoint" "test" {
  route_server_id = aws_vpc_route_server.test.id
  subnet_id      = aws_subnet.test.id

  tags = {
	Name = %[1]q
  }

  depends_on     = [aws_vpc_route_server_association.test]
}

resource "aws_vpc_route_server_peer" "test" {
  route_server_endpoint_id = aws_vpc_route_server_endpoint.test.id
  peer_address    = "10.0.1.250"
  bgp_options {
	peer_asn = 65000
  }

  tags = {
	Name = %[1]q
  }
}

`, rName)
}
