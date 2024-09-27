// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayDefaultRouteTablePropagation_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var transitgateway types.TransitGateway
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

	var transitgateway types.TransitGateway
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

func testAccCheckTransitGatewayDefaultRouteTablePropagationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_default_route_table_propagation" {
				continue
			}

			resp, err := tfec2.FindTransitGatewayByID(ctx, conn, rs.Primary.ID)

			if err != nil {
				if errs.IsA[*retry.NotFoundError](err) {
					return nil
				}
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameTransitGatewayDefaultRouteTablePropagation, rs.Primary.ID, err)
			}

			if *resp.Options.PropagationDefaultRouteTableId != *resp.Options.AssociationDefaultRouteTableId {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameTransitGatewayDefaultRouteTablePropagation, rs.Primary.ID, errors.New("not destroyed"))
			}
		}

		return nil
	}
}

func testAccCheckTransitGatewayDefaultRouteTablePropagationExists(ctx context.Context, name string, transitgatewaydefaultroutetableassociation *types.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameTransitGatewayDefaultRouteTablePropagation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameTransitGatewayDefaultRouteTablePropagation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		resp, err := tfec2.FindTransitGatewayByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameTransitGatewayDefaultRouteTablePropagation, rs.Primary.ID, err)
		}

		if *resp.Options.PropagationDefaultRouteTableId == *resp.Options.AssociationDefaultRouteTableId {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameTransitGatewayDefaultRouteTablePropagation, name, errors.New("not changed"))
		}

		*transitgatewaydefaultroutetableassociation = *resp

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
