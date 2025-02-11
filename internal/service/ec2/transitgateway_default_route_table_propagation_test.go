// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayDefaultRouteTablePropagation_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var transitgateway awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway_default_route_table_propagation.test"
	resourceRouteTableName := "aws_ec2_transit_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDefaultRouteTablePropagationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitgatewayDefaultRouteTablePropagationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayDefaultRouteTablePropagationExists(ctx, resourceName, &transitgateway),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", resourceRouteTableName, names.AttrID),
				),
			},
		},
	})
}

func testAccTransitGatewayDefaultRouteTablePropagation_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var transitgateway awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway_default_route_table_propagation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDefaultRouteTablePropagationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitgatewayDefaultRouteTablePropagationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayDefaultRouteTablePropagationExists(ctx, resourceName, &transitgateway),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayDefaultRouteTablePropagation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayDefaultRouteTablePropagation_Disappears_transitGateway(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var transitgateway awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway_default_route_table_propagation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDefaultRouteTablePropagationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitgatewayDefaultRouteTablePropagationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayDefaultRouteTablePropagationExists(ctx, resourceName, &transitgateway),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGateway(), "aws_ec2_transit_gateway.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTransitGatewayDefaultRouteTablePropagationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_default_route_table_propagation" {
				continue
			}

			output, err := tfec2.FindTransitGatewayByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if aws.ToString(output.Options.PropagationDefaultRouteTableId) == aws.ToString(output.Options.AssociationDefaultRouteTableId) {
				continue
			}

			return fmt.Errorf("EC2 Transit Gateway Default Route Table Propagation %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTransitGatewayDefaultRouteTablePropagationExists(ctx context.Context, n string, v *awstypes.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayByID(ctx, conn, rs.Primary.ID)

		if err == nil {
			if aws.ToString(output.Options.PropagationDefaultRouteTableId) == aws.ToString(output.Options.AssociationDefaultRouteTableId) {
				err = errors.New("EC2 Transit Gateway Default Route Table Propagation not found")
			}
		}

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTransitgatewayDefaultRouteTablePropagationConfig_basic() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

resource "aws_ec2_transit_gateway_default_route_table_propagation" "test" {
  transit_gateway_id             = aws_ec2_transit_gateway.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
`
}
