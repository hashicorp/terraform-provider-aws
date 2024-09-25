// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2TransitGatewayDefaultRouteTableAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var transitgateway types.TransitGateway
	resourceName := "aws_ec2_transit_gateway_default_route_table_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDefaultRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitgatewayDefaultRouteTableAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayDefaultRouteTableAssociationExists(ctx, resourceName, &transitgateway),
				),
			},
		},
	})
}

func TestAccEC2TransitGatewayDefaultRouteTableAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var transitgateway types.TransitGateway
	resourceName := "aws_ec2_transit_gateway_default_route_table_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDefaultRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitgatewayDefaultRouteTableAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayDefaultRouteTableAssociationExists(ctx, resourceName, &transitgateway),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitgatewayDefaultRouteTableAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTransitGatewayDefaultRouteTableAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_default_route_table_association" {
				continue
			}

			resp, err := tfec2.FindTransitGatewayByID(ctx, conn, rs.Primary.ID)

			if err != nil {
				if errs.IsA[*retry.NotFoundError](err) {
					return nil
				}
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameTransitGatewayDefaultRouteTableAssociation, rs.Primary.ID, err)
			}

			if *resp.Options.PropagationDefaultRouteTableId != *resp.Options.AssociationDefaultRouteTableId {

				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameTransitGatewayDefaultRouteTableAssociation, rs.Primary.ID, errors.New("not destroyed"))
			}
		}

		return nil
	}
}

func testAccCheckTransitGatewayDefaultRouteTableAssociationExists(ctx context.Context, name string, transitgatewaydefaultroutetableassociation *types.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameTransitGatewayDefaultRouteTableAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameTransitGatewayDefaultRouteTableAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		resp, err := tfec2.FindTransitGatewayByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameTransitGatewayDefaultRouteTableAssociation, rs.Primary.ID, err)
		}

		if *resp.Options.PropagationDefaultRouteTableId == *resp.Options.AssociationDefaultRouteTableId {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameTransitGatewayDefaultRouteTableAssociation, name, errors.New("not changed"))
		}

		*transitgatewaydefaultroutetableassociation = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeTransitGatewaysInput{}
	_, err := conn.DescribeTransitGateways(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTransitgatewayDefaultRouteTableAssociationConfig_basic() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

resource "aws_ec2_transit_gateway_default_route_table_association" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  route_table_id     = aws_ec2_transit_gateway_route_table.test.id
}
`
}
