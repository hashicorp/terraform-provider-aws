// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerTransitGatewayRegistration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:             testAccTransitGatewayRegistration_basic,
		acctest.CtDisappears:        testAccTransitGatewayRegistration_disappears,
		"disappears_TransitGateway": testAccTransitGatewayRegistration_Disappears_transitGateway,
		"crossRegion":               testAccTransitGatewayRegistration_crossRegion,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccTransitGatewayRegistration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_transit_gateway_registration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRegistrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRegistrationExists(ctx, t, resourceName),
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

func testAccTransitGatewayRegistration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_transit_gateway_registration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRegistrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRegistrationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfnetworkmanager.ResourceTransitGatewayRegistration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayRegistration_Disappears_transitGateway(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_transit_gateway_registration.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRegistrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRegistrationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceTransitGateway(), transitGatewayResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayRegistration_crossRegion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_transit_gateway_registration.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRegistrationConfig_crossRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRegistrationExists(ctx, t, resourceName),
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

func testAccCheckTransitGatewayRegistrationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_transit_gateway_registration" {
				continue
			}

			globalNetworkID, transitGatewayARN, err := tfnetworkmanager.TransitGatewayRegistrationParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfnetworkmanager.FindTransitGatewayRegistrationByTwoPartKey(ctx, conn, globalNetworkID, transitGatewayARN)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Transit Gateway Registration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTransitGatewayRegistrationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Transit Gateway Registration ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		globalNetworkID, transitGatewayARN, err := tfnetworkmanager.TransitGatewayRegistrationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfnetworkmanager.FindTransitGatewayRegistrationByTwoPartKey(ctx, conn, globalNetworkID, transitGatewayARN)

		return err
	}
}
func testAccTransitGatewayRegistrationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_transit_gateway_registration" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  transit_gateway_arn = aws_ec2_transit_gateway.test.arn
}
`, rName)
}

func testAccTransitGatewayRegistrationConfig_crossRegion(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  provider = "awsalternate"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_transit_gateway_registration" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  transit_gateway_arn = aws_ec2_transit_gateway.test.arn
}
`, rName))
}
