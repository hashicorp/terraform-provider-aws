package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccClientVPNRoute_basic(t *testing.T) {
	var v ec2.ClientVpnRoute
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnRouteConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "origin", "add-route"),
					resource.TestCheckResourceAttrPair(resourceName, "target_vpc_subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "Nat"),
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

func testAccClientVPNRoute_disappears(t *testing.T) {
	var v ec2.ClientVpnRoute
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnRouteConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceClientVPNRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClientVPNRoute_description(t *testing.T) {
	var v ec2.ClientVpnRoute
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnRouteConfigDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "test client VPN route"),
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

func testAccCheckClientVPNRouteDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_route" {
			continue
		}

		endpointID, targetSubnetID, destinationCIDR, err := tfec2.ClientVPNRouteParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindClientVPNRouteByThreePartKey(conn, endpointID, targetSubnetID, destinationCIDR)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Client VPN Route %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckClientVPNRouteExists(name string, v *ec2.ClientVpnRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Client VPN Route ID is set")
		}

		endpointID, targetSubnetID, destinationCIDR, err := tfec2.ClientVPNRouteParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindClientVPNRouteByThreePartKey(conn, endpointID, targetSubnetID, destinationCIDR)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEc2ClientVpnRouteConfigBase(rName string, subnetCount int) string {
	return acctest.ConfigCompose(
		testAccEc2ClientVpnEndpointConfig(rName),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count                   = %[2]d
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

func testAccEc2ClientVpnRouteConfigBasic(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnRouteConfigBase(rName, 1), `
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

func testAccEc2ClientVpnRouteConfigDescription(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnRouteConfigBase(rName, 1), `
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
