package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccNetworkManagerCustomerGatewayAssociation_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":                      testAccCustomerGatewayAssociation_basic,
		"disappears":                 testAccCustomerGatewayAssociation_disappears,
		"disappears_CustomerGateway": testAccCustomerGatewayAssociation_Disappears_customerGateway,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccCustomerGatewayAssociation_basic(t *testing.T) {
	resourceName := "aws_networkmanager_customer_gateway_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomerGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayAssociationExists(resourceName),
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

func testAccCustomerGatewayAssociation_disappears(t *testing.T) {
	resourceName := "aws_networkmanager_customer_gateway_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomerGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfnetworkmanager.ResourceCustomerGatewayAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCustomerGatewayAssociation_Disappears_customerGateway(t *testing.T) {
	resourceName := "aws_networkmanager_customer_gateway_association.test"
	vpnConnectionResourceName := "aws_vpn_connection.test"
	customerGatewayResourceName := "aws_customer_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCustomerGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPNConnection(), vpnConnectionResourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceCustomerGateway(), customerGatewayResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomerGatewayAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_customer_gateway_association" {
			continue
		}

		globalNetworkID, customerGatewayARN, err := tfnetworkmanager.CustomerGatewayAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfnetworkmanager.FindCustomerGatewayAssociationByTwoPartKey(context.TODO(), conn, globalNetworkID, customerGatewayARN)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Network Manager Customer Gateway Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCustomerGatewayAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Customer Gateway Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

		globalNetworkID, customerGatewayARN, err := tfnetworkmanager.CustomerGatewayAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfnetworkmanager.FindCustomerGatewayAssociationByTwoPartKey(context.TODO(), conn, globalNetworkID, customerGatewayARN)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCustomerGatewayAssociationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = 65534
  ip_address = "12.1.2.3"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type
  static_routes_only  = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_transit_gateway_registration" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  transit_gateway_arn = aws_ec2_transit_gateway.test.arn

  depends_on = [aws_vpn_connection.test]
}

resource "aws_networkmanager_customer_gateway_association" "test" {
  global_network_id    = aws_networkmanager_global_network.test.id
  customer_gateway_arn = aws_customer_gateway.test.arn
  device_id            = aws_networkmanager_device.test.id

  depends_on = [aws_networkmanager_transit_gateway_registration.test]
}
`, rName)
}
