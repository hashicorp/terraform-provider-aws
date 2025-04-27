// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

func TestAccVPCRouteServer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_route_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAsn := sdkacctest.RandIntRange(64512, 65534)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccVPCRouterServerPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterServerConfig_basic(rName, rAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", strconv.Itoa(rAsn)),
					resource.TestCheckResourceAttr(resourceName, "persist_routes", "disable"),
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

func TestAccVPCRouteServer_persistRoutes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_route_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAsn := sdkacctest.RandIntRange(64512, 65534)
	rPersistRoutes := "enable"
	rPersistRoutesDuration := 2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccVPCRouterServerPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterServerConfig_persistRoutes(rName, rAsn, rPersistRoutes, rPersistRoutesDuration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", strconv.Itoa(rAsn)),
					resource.TestCheckResourceAttr(resourceName, "persist_routes", rPersistRoutes),
					resource.TestCheckResourceAttr(resourceName, "persist_routes_duration", strconv.Itoa(rPersistRoutesDuration)),
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

func TestAccVPCRouteServer_updatePersitRoutesSNSNotification(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_route_server.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAsn := sdkacctest.RandIntRange(64512, 65534)
	rPersistRoutes := "enable"
	rPersistRoutesDuration := 1

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccVPCRouterServerPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterServerConfig_persistRoutes(rName, rAsn, rPersistRoutes, rPersistRoutesDuration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", strconv.Itoa(rAsn)),
					resource.TestCheckResourceAttr(resourceName, "persist_routes", rPersistRoutes),
					resource.TestCheckResourceAttr(resourceName, "persist_routes_duration", strconv.Itoa(rPersistRoutesDuration)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCRouterServerConfig_persistRoutesSNSNotification(rName, rAsn, "disable"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", fmt.Sprintf("%d", rAsn)),
					resource.TestCheckResourceAttr(resourceName, "persist_routes", "disable"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrSNSTopicARN),
				),
			},
		},
	})
}

func TestAccVPCRouteServer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_route_server.test"
	rAsn := sdkacctest.RandIntRange(64512, 65534)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccVPCRouterServerPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCRouteServerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouterServerConfig_basic(rName, rAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCRouteServer, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCRouteServerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_route_server" {
				continue
			}

			_, err := tfec2.FindVPCRouteServerByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCRouteServer, rs.Primary.ID, err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVPCRouteServer, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckVPCRouteServerExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServer, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServer, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		resp, err := tfec2.FindVPCRouteServerByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServer, rs.Primary.ID, err)
		}
		if resp == nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVPCRouteServer, rs.Primary.ID, errors.New("not found"))
		}
		return nil
	}
}

func testAccVPCRouterServerPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeRouteServersInput{}

	_, err := conn.DescribeRouteServers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVPCRouterServerConfig_basic(rName string, rAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpc_route_server" "test" {
  amazon_side_asn = %[2]d
  tags = {
    Name = %[1]q
  }
}
`, rName, rAsn)
}

func testAccVPCRouterServerConfig_persistRoutes(rName string, rAsn int, rPersistRoutes string, rPersistRoutesDuration int) string {
	return fmt.Sprintf(`
resource "aws_vpc_route_server" "test" {
  amazon_side_asn         = %[2]d
  persist_routes          = %[3]q
  persist_routes_duration = %[4]d
  tags = {
    Name = %[1]q
  }
}
`, rName, rAsn, rPersistRoutes, rPersistRoutesDuration)
}

func testAccVPCRouterServerConfig_persistRoutesSNSNotification(rName string, rAsn int, rPersistRoutes string) string {
	return fmt.Sprintf(`
resource "aws_vpc_route_server" "test" {
  amazon_side_asn           = %[2]d
  persist_routes            = %[3]q
  sns_notifications_enabled = true
  tags = {
    Name = %[1]q
  }
}
`, rName, rAsn, rPersistRoutes)
}
