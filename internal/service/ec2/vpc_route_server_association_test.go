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

func TestAccVPCRouteServerAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var vpcrouteserverassociation awstypes.RouteServerAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_route_server_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerAssociationExists(ctx, resourceName, &vpcrouteserverassociation),
					resource.TestCheckResourceAttrSet(resourceName, "route_server_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "route_server_id",
				ImportStateIdFunc:                    testAccVPCRouteServerAssociationImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccVPCRouteServerAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var vpcrouteserverassociation awstypes.RouteServerAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_route_server_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteServerAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerAssociationExists(ctx, resourceName, &vpcrouteserverassociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCRouteServerAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCRouteServerAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_route_server_association" {
				continue
			}

			_, err := tfec2.FindVPCRouteServerAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["route_server_id"], rs.Primary.Attributes[names.AttrVPCID])
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCRouteServerAssociation, rs.Primary.Attributes["route_server_id"], err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCRouteServerAssociation, rs.Primary.Attributes["route_server_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckVPCRouteServerAssociationExists(ctx context.Context, name string, vpcrouteserverassociation *awstypes.RouteServerAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServerAssociation, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["route_server_id"] == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServerAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := tfec2.FindVPCRouteServerAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["route_server_id"], rs.Primary.Attributes[names.AttrVPCID])
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServerAssociation, rs.Primary.Attributes["route_server_id"], err)
		}

		*vpcrouteserverassociation = *resp

		return nil
	}
}

func testAccVPCRouteServerAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["route_server_id"], rs.Primary.Attributes[names.AttrVPCID]), nil
	}
}

func testAccVPCRouteServerAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
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
`, rName)
}
