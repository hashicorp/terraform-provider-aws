// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccClientVPNRoute_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnRoute
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test.0"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNRouteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNRouteConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("client_vpn_endpoint_id"), endpointResourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("destination_cidr_block"), knownvalue.StringExact("0.0.0.0/0")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("origin"), knownvalue.StringExact("add-route")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTransitGatewayAttachmentID), knownvalue.StringExact("")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("target_vpc_subnet_id"), subnetResourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), knownvalue.StringExact("Nat")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccClientVPNRoute_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnRoute
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNRouteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNRouteConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceClientVPNRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccClientVPNRoute_transitGateway(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnRoute
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"
	tgwId := acctest.SkipIfEnvVarNotSet(t, "CLIENT_VPN_TGW_ID")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNRouteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNRouteConfig_transitGateway(t, rName, tgwId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("destination_cidr_block"), knownvalue.StringExact("10.2.0.0/24")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_vpc_subnet_id"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTransitGatewayAttachmentID), tfknownvalue.StringNotEmpty()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccClientVPNRoute_description(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.ClientVpnRoute
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNRouteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNRouteConfig_description(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("test client VPN route")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckClientVPNRouteDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_client_vpn_route" {
				continue
			}

			_, err := tfec2.FindClientVPNRouteByThreePartKey(ctx, conn, rs.Primary.Attributes["client_vpn_endpoint_id"], rs.Primary.Attributes["target_vpc_subnet_id"], rs.Primary.Attributes["destination_cidr_block"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Client VPN Route %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClientVPNRouteExists(ctx context.Context, t *testing.T, name string, v *awstypes.ClientVpnRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		output, err := tfec2.FindClientVPNRouteByThreePartKey(ctx, conn, rs.Primary.Attributes["client_vpn_endpoint_id"], rs.Primary.Attributes["target_vpc_subnet_id"], rs.Primary.Attributes["destination_cidr_block"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClientVPNRouteConfig_base(t *testing.T, rName string, subnetCount int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), testAccClientVPNEndpointConfig_basic(t, rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount))
}

func testAccClientVPNRouteConfig_basic(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNRouteConfig_base(t, rName, 1), `
resource "aws_ec2_client_vpn_route" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_network_association.test.client_vpn_endpoint_id
  destination_cidr_block = "0.0.0.0/0"
  target_vpc_subnet_id   = aws_subnet.test[0].id
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[0].id
}
`)
}

func testAccClientVPNRouteConfig_description(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNRouteConfig_base(t, rName, 1), `
resource "aws_ec2_client_vpn_route" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_network_association.test.client_vpn_endpoint_id
  destination_cidr_block = "0.0.0.0/0"
  target_vpc_subnet_id   = aws_subnet.test[0].id
  description            = "test client VPN route"
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[0].id
}
`)
}

func testAccClientVPNRouteConfig_transitGateway(t *testing.T, rName, tgwId string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_transitGatewayConfigurationWithNoAvailabilityZones(t, rName, tgwId), `
resource "aws_ec2_client_vpn_route" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  destination_cidr_block = "10.2.0.0/24"
}
`)
}
