// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCRouteServer_basic(t *testing.T) {
	ctx := acctest.Context(t)
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
				Config: testAccVPCRouterServerConfig_basic(rAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("amazon_side_asn"), knownvalue.Int64Exact(int64(rAsn))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`route-server/rs-[a-z0-9]+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("persist_routes"), tfknownvalue.StringExact(awstypes.RouteServerPersistRoutesActionDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("persist_routes_duration"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("route_server_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sns_notifications_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrSNSTopicARN), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "route_server_id"),
				ImportStateVerifyIdentifierAttribute: "route_server_id",
			},
		},
	})
}

func TestAccVPCRouteServer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
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
				Config: testAccVPCRouterServerConfig_basic(rAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCRouteServer, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCRouteServer_persistRoutes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_route_server.test"
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
				Config: testAccVPCRouterServerConfig_persistRoutes(rAsn, rPersistRoutes, rPersistRoutesDuration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("persist_routes"), tfknownvalue.StringExact(awstypes.RouteServerPersistRoutesActionEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("persist_routes_duration"), knownvalue.Int64Exact(int64(rPersistRoutesDuration))),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "route_server_id"),
				ImportStateVerifyIdentifierAttribute: "route_server_id",
			},
		},
	})
}

func TestAccVPCRouteServer_updatePersitRoutesSNSNotification(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_route_server.test"
	rAsn := sdkacctest.RandIntRange(64512, 65534)
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
				Config: testAccVPCRouterServerConfig_persistRoutes(rAsn, "enable", rPersistRoutesDuration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("persist_routes"), tfknownvalue.StringExact(awstypes.RouteServerPersistRoutesActionEnable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("persist_routes_duration"), knownvalue.Int64Exact(int64(rPersistRoutesDuration))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sns_notifications_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrSNSTopicARN), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "route_server_id"),
				ImportStateVerifyIdentifierAttribute: "route_server_id",
			},
			{
				Config: testAccVPCRouterServerConfig_persistRoutesSNSNotification(rAsn, "disable"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("persist_routes"), tfknownvalue.StringExact(awstypes.RouteServerPersistRoutesActionDisable)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("persist_routes_duration"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("sns_notifications_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrSNSTopicARN), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccVPCRouteServer_tags(t *testing.T) {
	ctx := acctest.Context(t)
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
				Config: testAccVPCRouterServerConfig_tags1(rAsn, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "route_server_id"),
				ImportStateVerifyIdentifierAttribute: "route_server_id",
			},
			{
				Config: testAccVPCRouterServerConfig_tags2(rAsn, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccVPCRouterServerConfig_tags1(rAsn, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCRouteServerExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
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

			_, err := tfec2.FindRouteServerByID(ctx, conn, rs.Primary.Attributes["route_server_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Route Server %s still exists", rs.Primary.Attributes["route_server_id"])
		}

		return nil
	}
}

func testAccCheckVPCRouteServerExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindRouteServerByID(ctx, conn, rs.Primary.Attributes["route_server_id"])

		return err
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

func testAccVPCRouterServerConfig_basic(rAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpc_route_server" "test" {
  amazon_side_asn = %[1]d
}
`, rAsn)
}

func testAccVPCRouterServerConfig_persistRoutes(rAsn int, rPersistRoutes string, rPersistRoutesDuration int) string {
	return fmt.Sprintf(`
resource "aws_vpc_route_server" "test" {
  amazon_side_asn         = %[1]d
  persist_routes          = %[2]q
  persist_routes_duration = %[3]d
}
`, rAsn, rPersistRoutes, rPersistRoutesDuration)
}

func testAccVPCRouterServerConfig_persistRoutesSNSNotification(rAsn int, rPersistRoutes string) string {
	return fmt.Sprintf(`
resource "aws_vpc_route_server" "test" {
  amazon_side_asn           = %[1]d
  persist_routes            = %[2]q
  sns_notifications_enabled = true
}
`, rAsn, rPersistRoutes)
}

func testAccVPCRouterServerConfig_tags1(rAsn int, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_vpc_route_server" "test" {
  amazon_side_asn = %[1]d

  tags = {
    %[2]q = %[3]q
  }
}
`, rAsn, tag1Key, tag1Value)
}

func testAccVPCRouterServerConfig_tags2(rAsn int, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_vpc_route_server" "test" {
  amazon_side_asn = %[1]d

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rAsn, tag1Key, tag1Value, tag2Key, tag2Value)
}
